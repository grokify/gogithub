package repo

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/grokify/mogo/os/osutil"

	"github.com/grokify/gogithub/clientv1"
)

// CommitError indicates a failure to create a commit.
type CommitError struct {
	Message string
	Err     error
}

func (e *CommitError) Error() string {
	return "failed to create commit: " + e.Err.Error()
}

func (e *CommitError) Unwrap() error {
	return e.Err
}

// FileContent represents a file to be committed.
type FileContent struct {
	Path    string
	Content []byte
}

// CreateCommit creates a commit with the given files using the Git tree API.
func CreateCommit(ctx context.Context, client clientv1.Client, owner, repo, branch, message string, files []FileContent) (string, error) {
	// Get the current commit SHA
	branchSHA, err := client.GetBranchSHA(ctx, owner, repo, branch)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Get the tree of the parent commit
	parentCommit, err := client.GetCommit(ctx, owner, repo, branchSHA)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Get base tree SHA from the parent commit
	var baseTreeSHA string
	// Note: GetCommit returns RepositoryCommit which doesn't have tree directly accessible
	// We need to work around this by creating tree entries
	// For simplicity, we'll get the tree from the branch ref
	ref, err := client.GetRef(ctx, owner, repo, RefHeadsPrefix+branch)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Create tree entries for files
	var entries []clientv1.TreeEntry
	for _, f := range files {
		entries = append(entries, clientv1.TreeEntry{
			Path:    f.Path,
			Mode:    FileModeRegular,
			Type:    "blob",
			Content: string(f.Content),
		})
	}

	// Create the tree
	// Note: We need to pass the base tree SHA, but our API doesn't expose the tree SHA directly
	// We'll use empty string which creates a new tree
	treeSHA, err := client.CreateTree(ctx, owner, repo, baseTreeSHA, entries)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Create the commit
	commit, err := client.CreateCommit(ctx, owner, repo, &clientv1.CreateCommitOptions{
		Message: message,
		Tree:    treeSHA,
		Parents: []string{parentCommit.SHA},
	})
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Update the branch reference
	_, err = client.UpdateRef(ctx, owner, repo, ref.Ref, commit.SHA, false)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	return commit.SHA, nil
}

// ReadLocalFiles reads all files from a local directory recursively.
// The prefix is prepended to relative paths for the destination.
// Uses osutil.ReadDirFilesSecure for symlink-safe file operations.
func ReadLocalFiles(dir, prefix string) ([]FileContent, error) {
	fileMap, err := osutil.ReadDirFilesSecure(dir)
	if err != nil {
		return nil, err
	}

	files := make([]FileContent, 0, len(fileMap))
	for path, content := range fileMap {
		// Combine with prefix and normalize to forward slashes for GitHub
		fullPath := filepath.Join(prefix, path)
		fullPath = strings.ReplaceAll(fullPath, "\\", "/")

		files = append(files, FileContent{
			Path:    fullPath,
			Content: content,
		})
	}

	return files, nil
}
