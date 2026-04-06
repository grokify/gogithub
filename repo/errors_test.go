package repo

import (
	"errors"
	"testing"
)

func TestCommitError(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &CommitError{
		Message: "test commit",
		Err:     innerErr,
	}

	// Test Error() method
	expected := "failed to create commit: inner error"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap() method
	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}

	// Test error chain compatibility
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestBranchError(t *testing.T) {
	innerErr := errors.New("branch creation failed")
	err := &BranchError{
		Branch: "feature-branch",
		Err:    innerErr,
	}

	// Test Error() method
	expected := "failed to create branch feature-branch: branch creation failed"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap() method
	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}

	// Test error chain compatibility
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestForkError(t *testing.T) {
	innerErr := errors.New("fork failed")
	err := &ForkError{
		Owner: "owner",
		Repo:  "repo",
		Err:   innerErr,
	}

	// Test Error() method
	expected := "failed to fork owner/repo: fork failed"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap() method
	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}

	// Test error chain compatibility
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestBatchError(t *testing.T) {
	innerErr := errors.New("operation failed")
	err := &BatchError{
		Op:  "create blob",
		Err: innerErr,
	}

	// Test Error() method
	expected := "batch create blob failed: operation failed"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap() method
	if err.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), innerErr)
	}

	// Test error chain compatibility
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}
