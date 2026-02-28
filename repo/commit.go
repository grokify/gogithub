package repo

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v84/github"
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
func CreateCommit(ctx context.Context, gh *github.Client, owner, repo, branch, message string, files []FileContent) (string, error) {
	// Get the current commit SHA
	ref, _, err := gh.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}
	parentSHA := ref.GetObject().GetSHA()

	// Get the tree of the parent commit
	parentCommit, _, err := gh.Git.GetCommit(ctx, owner, repo, parentSHA)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}
	baseTreeSHA := parentCommit.GetTree().GetSHA()

	// Create tree entries for files
	var entries []*github.TreeEntry
	for _, f := range files {
		entries = append(entries, &github.TreeEntry{
			Path:    github.Ptr(f.Path),
			Mode:    github.Ptr("100644"),
			Type:    github.Ptr("blob"),
			Content: github.Ptr(string(f.Content)),
		})
	}

	// Create the tree
	tree, _, err := gh.Git.CreateTree(ctx, owner, repo, baseTreeSHA, entries)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Create the commit
	commit, _, err := gh.Git.CreateCommit(ctx, owner, repo, github.Commit{
		Message: github.Ptr(message),
		Tree:    tree,
		Parents: []*github.Commit{{SHA: github.Ptr(parentSHA)}},
	}, nil)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	// Update the branch reference
	updateRef := github.UpdateRef{
		SHA:   *commit.SHA,
		Force: github.Ptr(false),
	}
	_, _, err = gh.Git.UpdateRef(ctx, owner, repo, *ref.Ref, updateRef)
	if err != nil {
		return "", &CommitError{Message: message, Err: err}
	}

	return commit.GetSHA(), nil
}

// ReadLocalFiles reads all files from a local directory recursively.
// The prefix is prepended to relative paths for the destination.
func ReadLocalFiles(dir, prefix string) ([]FileContent, error) {
	var files []FileContent

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Combine with prefix
		fullPath := filepath.Join(prefix, relPath)
		// Normalize to forward slashes for GitHub
		fullPath = strings.ReplaceAll(fullPath, "\\", "/")

		files = append(files, FileContent{
			Path:    fullPath,
			Content: content,
		})

		return nil
	})

	return files, err
}
