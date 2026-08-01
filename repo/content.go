package repo

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ContentOptions specifies options for fetching repository content.
// Deprecated: Use gogithub.ContentOptions instead.
type ContentOptions = gogithub.ContentOptions

// FileInfo represents information about a file or directory in a repository.
type FileInfo struct {
	Path        string
	Name        string
	Type        string // "file" or "dir"
	Size        int
	SHA         string
	DownloadURL string
}

// GetFileContent fetches the content of a single file from a repository.
// Returns the decoded file content as bytes.
func GetFileContent(ctx context.Context, client clientv1.Client, owner, repo, filePath string, opts *ContentOptions) ([]byte, error) {
	return client.GetFileContent(ctx, owner, repo, filePath, opts)
}

// GetFileContentString fetches the content of a single file as a string.
func GetFileContentString(ctx context.Context, client clientv1.Client, owner, repo, filePath string, opts *ContentOptions) (string, error) {
	return client.GetFileContentString(ctx, owner, repo, filePath, opts)
}

// ListDirectory lists the contents of a directory in a repository.
func ListDirectory(ctx context.Context, client clientv1.Client, owner, repo, dirPath string, opts *ContentOptions) ([]FileInfo, error) {
	contents, err := client.ListDirectory(ctx, owner, repo, dirPath, opts)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, item := range contents {
		files = append(files, FileInfo{
			Path:        item.Path,
			Name:        item.Name,
			Type:        item.Type,
			Size:        item.Size,
			SHA:         item.SHA,
			DownloadURL: item.DownloadURL,
		})
	}

	return files, nil
}

// ListDirectoryRecursive lists all files in a directory recursively.
func ListDirectoryRecursive(ctx context.Context, client clientv1.Client, owner, repo, dirPath string, opts *ContentOptions) ([]FileInfo, error) {
	var allFiles []FileInfo

	items, err := ListDirectory(ctx, client, owner, repo, dirPath, opts)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		switch item.Type {
		case "file":
			allFiles = append(allFiles, item)
		case "dir":
			subFiles, err := ListDirectoryRecursive(ctx, client, owner, repo, item.Path, opts)
			if err != nil {
				return nil, err
			}
			allFiles = append(allFiles, subFiles...)
		}
	}

	return allFiles, nil
}

// GetMultipleFiles fetches multiple files from a repository.
// Returns a map of file path to content.
func GetMultipleFiles(ctx context.Context, client clientv1.Client, owner, repo string, filePaths []string, opts *ContentOptions) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for _, filePath := range filePaths {
		content, err := client.GetFileContent(ctx, owner, repo, filePath, opts)
		if err != nil {
			return nil, err
		}
		result[filePath] = content
	}

	return result, nil
}

// DownloadDirectory downloads all files from a directory recursively.
// Returns a map of relative file path to content.
func DownloadDirectory(ctx context.Context, client clientv1.Client, owner, repo, dirPath string, opts *ContentOptions) (map[string][]byte, error) {
	files, err := ListDirectoryRecursive(ctx, client, owner, repo, dirPath, opts)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for _, file := range files {
		content, err := client.GetFileContent(ctx, owner, repo, file.Path, opts)
		if err != nil {
			return nil, fmt.Errorf("download %s: %w", file.Path, err)
		}
		// Use relative path from dirPath
		relPath := strings.TrimPrefix(file.Path, dirPath)
		relPath = strings.TrimPrefix(relPath, "/")
		result[relPath] = content
	}

	return result, nil
}

// FileExists checks if a file exists in a repository.
func FileExists(ctx context.Context, client clientv1.Client, owner, repo, filePath string, opts *ContentOptions) (bool, error) {
	return client.FileExists(ctx, owner, repo, filePath, opts)
}

// GetRawFileURL returns the raw content URL for a file.
// This URL can be used to download the file without authentication for public repos.
func GetRawFileURL(owner, repo, ref, filePath string) string {
	if ref == "" {
		ref = "main"
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, filePath)
}

// ParseRepoURL parses a GitHub repository URL and returns owner and repo.
// Supports formats:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo.git
//   - git@github.com:owner/repo.git
//   - owner/repo
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	// Handle owner/repo format
	if !strings.Contains(repoURL, "://") && !strings.Contains(repoURL, "@") {
		parts := strings.Split(repoURL, "/")
		if len(parts) == 2 {
			return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
		}
		return "", "", fmt.Errorf("invalid repo format: %s", repoURL)
	}

	// Handle git@github.com:owner/repo.git
	if strings.HasPrefix(repoURL, "git@") {
		repoURL = strings.TrimPrefix(repoURL, "git@github.com:")
		repoURL = strings.TrimSuffix(repoURL, ".git")
		parts := strings.Split(repoURL, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
		return "", "", fmt.Errorf("invalid git SSH URL: %s", repoURL)
	}

	// Handle https://github.com/owner/repo
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "http://")
	repoURL = strings.TrimPrefix(repoURL, "github.com/")
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimSuffix(repoURL, "/")

	parts := strings.Split(repoURL, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid GitHub URL: %s", repoURL)
}

// JoinPath joins path segments for repository file paths.
func JoinPath(segments ...string) string {
	return path.Join(segments...)
}
