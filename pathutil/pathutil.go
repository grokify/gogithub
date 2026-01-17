// Package pathutil provides path validation and normalization utilities
// for GitHub repository paths.
package pathutil

import (
	"errors"
	"path"
	"strings"
)

// ErrInvalidPath is returned when a path contains invalid characters or patterns.
var ErrInvalidPath = errors.New("invalid path")

// ErrPathTraversal is returned when a path contains directory traversal attempts.
var ErrPathTraversal = errors.New("path traversal not allowed")

// Validate checks if a path is valid for GitHub repository operations.
// It returns an error if the path contains:
//   - Directory traversal sequences (..)
//   - Null bytes or other control characters
//   - Invalid characters for GitHub paths
//
// An empty path is considered valid (represents root).
func Validate(p string) error {
	if p == "" {
		return nil // Empty path is valid (root)
	}

	// Check for path traversal before cleaning
	if strings.Contains(p, "..") {
		return ErrPathTraversal
	}

	// Check for null bytes
	if strings.ContainsRune(p, 0) {
		return ErrInvalidPath
	}

	return nil
}

// Normalize normalizes a path for GitHub API operations.
// It performs the following transformations:
//   - Cleans the path (removes redundant separators, etc.)
//   - Removes leading slashes
//   - Converts backslashes to forward slashes
//   - Returns empty string for root paths (".", "/", "")
func Normalize(p string) string {
	if p == "" {
		return ""
	}

	// Convert backslashes to forward slashes (Windows compatibility)
	p = strings.ReplaceAll(p, "\\", "/")

	// Clean the path
	p = path.Clean(p)

	// Remove leading slash
	p = strings.TrimPrefix(p, "/")

	// Handle root representations
	if p == "." {
		return ""
	}

	return p
}

// ValidateAndNormalize validates and normalizes a path in one operation.
// Returns the normalized path and any validation error.
func ValidateAndNormalize(p string) (string, error) {
	if err := Validate(p); err != nil {
		return "", err
	}
	return Normalize(p), nil
}

// Join joins path elements and normalizes the result for GitHub.
// Empty elements are ignored.
func Join(elem ...string) string {
	// Filter empty elements
	var filtered []string
	for _, e := range elem {
		if e != "" {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	return Normalize(path.Join(filtered...))
}

// Split splits a path into directory and file components.
// Returns ("", filename) for files in the root directory.
func Split(p string) (dir, file string) {
	p = Normalize(p)
	return path.Split(p)
}

// Dir returns the directory portion of a path.
// Returns "" for files in the root directory.
func Dir(p string) string {
	dir, _ := Split(p)
	// Remove trailing slash from dir
	return strings.TrimSuffix(dir, "/")
}

// Base returns the last element of a path.
// Returns "" for empty paths.
func Base(p string) string {
	p = Normalize(p)
	if p == "" {
		return ""
	}
	return path.Base(p)
}

// Ext returns the file extension of a path.
// Returns "" if there is no extension.
func Ext(p string) string {
	return path.Ext(p)
}

// HasPrefix reports whether the path has the given prefix.
// Both paths are normalized before comparison.
func HasPrefix(p, prefix string) bool {
	p = Normalize(p)
	prefix = Normalize(prefix)

	if prefix == "" {
		return true
	}

	if p == prefix {
		return true
	}

	// Check if p starts with prefix followed by a slash
	return strings.HasPrefix(p, prefix+"/")
}
