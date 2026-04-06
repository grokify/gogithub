package repo

// Git reference prefixes for GitHub API operations.
const (
	// RefHeadsPrefix is the prefix for branch references.
	RefHeadsPrefix = "refs/heads/"
	// RefTagsPrefix is the prefix for tag references.
	RefTagsPrefix = "refs/tags/"
)

// Git file modes for tree entries.
const (
	// FileModeRegular is the mode for regular files (644 permissions).
	FileModeRegular = "100644"
	// FileModeExecutable is the mode for executable files (755 permissions).
	FileModeExecutable = "100755"
	// FileModeSubmodule is the mode for submodule entries.
	FileModeSubmodule = "160000"
	// FileModeSymlink is the mode for symbolic links.
	FileModeSymlink = "120000"
)

// Error message fragments for error detection.
const (
	// ErrAlreadyExists is the error message fragment indicating a resource already exists.
	ErrAlreadyExists = "already exists"
)
