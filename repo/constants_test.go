package repo

import (
	"testing"
)

func TestRefPrefixConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"RefHeadsPrefix", RefHeadsPrefix, "refs/heads/"},
		{"RefTagsPrefix", RefTagsPrefix, "refs/tags/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestFileModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FileModeRegular", FileModeRegular, "100644"},
		{"FileModeExecutable", FileModeExecutable, "100755"},
		{"FileModeSubmodule", FileModeSubmodule, "160000"},
		{"FileModeSymlink", FileModeSymlink, "120000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestErrAlreadyExists(t *testing.T) {
	if ErrAlreadyExists != "already exists" {
		t.Errorf("ErrAlreadyExists = %q, want %q", ErrAlreadyExists, "already exists")
	}
}
