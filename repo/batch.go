package repo

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/gogithub/pathutil"
)

// BatchOperationType indicates the type of batch operation.
type BatchOperationType int

const (
	// BatchOpWrite represents a write (create/update) operation.
	BatchOpWrite BatchOperationType = iota
	// BatchOpDelete represents a delete operation.
	BatchOpDelete
)

// BatchOperation represents a single operation in a batch.
type BatchOperation struct {
	Type    BatchOperationType
	Path    string
	Content []byte
}

// BatchError indicates a failure during batch operations.
type BatchError struct {
	Op  string
	Err error
}

func (e *BatchError) Error() string {
	return fmt.Sprintf("batch %s failed: %v", e.Op, e.Err)
}

func (e *BatchError) Unwrap() error {
	return e.Err
}

// ErrBatchCommitted is returned when operations are attempted on a committed batch.
var ErrBatchCommitted = errors.New("batch already committed")

// ErrEmptyPath is returned when an empty path is provided.
var ErrEmptyPath = errors.New("empty path not allowed")

// Batch accumulates multiple file operations to be committed atomically.
// Use NewBatch to create a batch, then call Write/Delete to queue operations,
// and finally Commit to apply all changes in a single commit.
type Batch struct {
	client     *github.Client
	owner      string
	repo       string
	branch     string
	message    string
	author     *github.CommitAuthor
	operations []BatchOperation
	committed  bool
	mu         sync.Mutex
}

// BatchOption configures a Batch.
type BatchOption func(*Batch)

// WithCommitAuthor sets the commit author for the batch.
func WithCommitAuthor(name, email string) BatchOption {
	return func(b *Batch) {
		b.author = &github.CommitAuthor{
			Name:  github.Ptr(name),
			Email: github.Ptr(email),
		}
	}
}

// NewBatch creates a new batch for accumulating file operations.
// The message is used as the commit message when Commit is called.
func NewBatch(ctx context.Context, gh *github.Client, owner, repo, branch, message string, opts ...BatchOption) (*Batch, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if message == "" {
		message = "Batch update"
	}

	b := &Batch{
		client:     gh,
		owner:      owner,
		repo:       repo,
		branch:     branch,
		message:    message,
		operations: make([]BatchOperation, 0),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

// Write queues a file write operation.
// The file will be created or updated when Commit is called.
func (b *Batch) Write(filePath string, content []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.committed {
		return ErrBatchCommitted
	}

	if err := pathutil.Validate(filePath); err != nil {
		return err
	}

	if filePath == "" {
		return ErrEmptyPath
	}

	b.operations = append(b.operations, BatchOperation{
		Type:    BatchOpWrite,
		Path:    pathutil.Normalize(filePath),
		Content: content,
	})

	return nil
}

// Delete queues a file delete operation.
// The file will be deleted when Commit is called.
// If the file doesn't exist at commit time, it is ignored (no error).
func (b *Batch) Delete(filePath string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.committed {
		return ErrBatchCommitted
	}

	if err := pathutil.Validate(filePath); err != nil {
		return err
	}

	if filePath == "" {
		return ErrEmptyPath
	}

	b.operations = append(b.operations, BatchOperation{
		Type: BatchOpDelete,
		Path: pathutil.Normalize(filePath),
	})

	return nil
}

// Len returns the number of operations in the batch.
func (b *Batch) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.operations)
}

// Operations returns a copy of the queued operations.
func (b *Batch) Operations() []BatchOperation {
	b.mu.Lock()
	defer b.mu.Unlock()

	ops := make([]BatchOperation, len(b.operations))
	copy(ops, b.operations)
	return ops
}

// Committed returns whether the batch has been committed.
func (b *Batch) Committed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.committed
}

// Commit applies all queued operations in a single commit.
// This uses the Git Data API (Trees and Commits) to create an atomic commit.
//
// The process:
//  1. Get the current commit SHA from the branch reference
//  2. Get the current tree SHA from that commit
//  3. Create blobs for all new/updated file contents
//  4. Create a new tree with the changes
//  5. Create a new commit pointing to the new tree
//  6. Update the branch reference to the new commit
//
// Returns the new commit SHA on success.
func (b *Batch) Commit(ctx context.Context) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.committed {
		return "", ErrBatchCommitted
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	if len(b.operations) == 0 {
		b.committed = true
		return "", nil // Nothing to commit
	}

	// Step 1: Get the current branch reference
	ref, _, err := b.client.Git.GetRef(ctx, b.owner, b.repo, "refs/heads/"+b.branch)
	if err != nil {
		return "", &BatchError{Op: "get ref", Err: err}
	}

	currentCommitSHA := ref.Object.GetSHA()

	// Step 2: Get the current commit to find the tree SHA
	currentCommit, _, err := b.client.Git.GetCommit(ctx, b.owner, b.repo, currentCommitSHA)
	if err != nil {
		return "", &BatchError{Op: "get commit", Err: err}
	}

	baseTreeSHA := currentCommit.Tree.GetSHA()

	// Step 3: Build tree entries for all operations
	treeEntries, err := b.buildTreeEntries(ctx)
	if err != nil {
		return "", err
	}

	// If no entries (all deletes were for non-existent files), nothing to commit
	if len(treeEntries) == 0 {
		b.committed = true
		return "", nil
	}

	// Step 4: Create the new tree
	newTree, _, err := b.client.Git.CreateTree(ctx, b.owner, b.repo, baseTreeSHA, treeEntries)
	if err != nil {
		return "", &BatchError{Op: "create tree", Err: err}
	}

	// Step 5: Create the new commit
	commitOpts := github.Commit{
		Message: github.Ptr(b.message),
		Tree:    newTree,
		Parents: []*github.Commit{{SHA: github.Ptr(currentCommitSHA)}},
	}

	if b.author != nil {
		commitOpts.Author = b.author
	}

	newCommit, _, err := b.client.Git.CreateCommit(ctx, b.owner, b.repo, commitOpts, nil)
	if err != nil {
		return "", &BatchError{Op: "create commit", Err: err}
	}

	// Step 6: Update the branch reference
	newCommitSHA := newCommit.GetSHA()
	updateRef := github.UpdateRef{
		SHA:   newCommitSHA,
		Force: github.Ptr(false),
	}

	_, _, err = b.client.Git.UpdateRef(ctx, b.owner, b.repo, ref.GetRef(), updateRef)
	if err != nil {
		return "", &BatchError{Op: "update ref", Err: err}
	}

	b.committed = true
	return newCommitSHA, nil
}

// buildTreeEntries creates GitHub tree entries for all operations.
func (b *Batch) buildTreeEntries(ctx context.Context) ([]*github.TreeEntry, error) {
	entries := make([]*github.TreeEntry, 0, len(b.operations))

	for _, op := range b.operations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		switch op.Type {
		case BatchOpWrite:
			// Create a blob for the content
			blob, _, err := b.client.Git.CreateBlob(ctx, b.owner, b.repo, github.Blob{
				Content:  github.Ptr(string(op.Content)),
				Encoding: github.Ptr("utf-8"),
			})
			if err != nil {
				return nil, &BatchError{Op: "create blob", Err: err}
			}

			entries = append(entries, &github.TreeEntry{
				Path: github.Ptr(op.Path),
				Mode: github.Ptr("100644"), // Regular file
				Type: github.Ptr("blob"),
				SHA:  blob.SHA,
			})

		case BatchOpDelete:
			// Check if the file exists first
			exists, err := b.fileExists(ctx, op.Path)
			if err != nil {
				return nil, err
			}
			if exists {
				// Use a nil SHA to indicate deletion
				entries = append(entries, &github.TreeEntry{
					Path: github.Ptr(op.Path),
					Mode: github.Ptr("100644"),
					Type: github.Ptr("blob"),
					SHA:  nil, // nil SHA means delete
				})
			}
			// If file doesn't exist, skip it (idempotent)
		}
	}

	return entries, nil
}

// fileExists checks if a file exists in the repository.
func (b *Batch) fileExists(ctx context.Context, path string) (bool, error) {
	_, _, resp, err := b.client.Repositories.GetContents(ctx, b.owner, b.repo, path, &github.RepositoryContentGetOptions{
		Ref: b.branch,
	})
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if errResp.Response != nil && errResp.Response.StatusCode == 404 {
				return false, nil
			}
		}
		return false, &BatchError{Op: "check file exists", Err: err}
	}
	return true, nil
}
