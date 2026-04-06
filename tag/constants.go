package tag

// Git reference constants for tag operations.
const (
	// refTagsPrefix is the full prefix for tag references in Git.
	refTagsPrefix = "refs/tags/"
	// tagsPrefix is the short prefix used by some GitHub API endpoints.
	tagsPrefix = "tags/"
)

// Error message fragments for error detection.
const (
	// errAlreadyExists is the error message fragment indicating a resource already exists.
	errAlreadyExists = "already exists"
)
