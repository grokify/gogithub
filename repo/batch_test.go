package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/grokify/gogithub/pathutil"
)

func TestBatchWrite(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	// Test valid write
	err := batch.Write("file.txt", []byte("content"))
	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
	}

	if batch.Len() != 1 {
		t.Errorf("Len() = %d, want 1", batch.Len())
	}

	ops := batch.Operations()
	if len(ops) != 1 {
		t.Fatalf("Operations() len = %d, want 1", len(ops))
	}
	if ops[0].Type != BatchOpWrite {
		t.Errorf("Operation type = %v, want BatchOpWrite", ops[0].Type)
	}
	if ops[0].Path != "file.txt" {
		t.Errorf("Operation path = %q, want %q", ops[0].Path, "file.txt")
	}
}

func TestBatchWriteNormalizesPath(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	err := batch.Write("/dir/file.txt", []byte("content"))
	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
	}

	ops := batch.Operations()
	if ops[0].Path != "dir/file.txt" {
		t.Errorf("Operation path = %q, want %q (normalized)", ops[0].Path, "dir/file.txt")
	}
}

func TestBatchWriteInvalidPath(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	// Test path traversal
	err := batch.Write("../etc/passwd", []byte("content"))
	if !errors.Is(err, pathutil.ErrPathTraversal) {
		t.Errorf("Write() error = %v, want ErrPathTraversal", err)
	}

	// Test empty path
	err = batch.Write("", []byte("content"))
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Write() error = %v, want ErrEmptyPath", err)
	}
}

func TestBatchDelete(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	err := batch.Delete("file.txt")
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	if batch.Len() != 1 {
		t.Errorf("Len() = %d, want 1", batch.Len())
	}

	ops := batch.Operations()
	if ops[0].Type != BatchOpDelete {
		t.Errorf("Operation type = %v, want BatchOpDelete", ops[0].Type)
	}
}

func TestBatchDeleteInvalidPath(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	// Test path traversal
	err := batch.Delete("../etc/passwd")
	if !errors.Is(err, pathutil.ErrPathTraversal) {
		t.Errorf("Delete() error = %v, want ErrPathTraversal", err)
	}

	// Test empty path
	err = batch.Delete("")
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Delete() error = %v, want ErrEmptyPath", err)
	}
}

func TestBatchCommitted(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
		committed:  true,
	}

	// Test write after commit
	err := batch.Write("file.txt", []byte("content"))
	if !errors.Is(err, ErrBatchCommitted) {
		t.Errorf("Write() error = %v, want ErrBatchCommitted", err)
	}

	// Test delete after commit
	err = batch.Delete("file.txt")
	if !errors.Is(err, ErrBatchCommitted) {
		t.Errorf("Delete() error = %v, want ErrBatchCommitted", err)
	}
}

func TestBatchCommitEmpty(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	sha, err := batch.Commit(context.Background())
	if err != nil {
		t.Errorf("Commit() error = %v, want nil", err)
	}
	if sha != "" {
		t.Errorf("Commit() sha = %q, want empty", sha)
	}
	if !batch.Committed() {
		t.Error("Committed() = false, want true")
	}
}

func TestBatchCommitAlreadyCommitted(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
		committed:  true,
	}

	_, err := batch.Commit(context.Background())
	if !errors.Is(err, ErrBatchCommitted) {
		t.Errorf("Commit() error = %v, want ErrBatchCommitted", err)
	}
}

func TestBatchCommitCanceledContext(t *testing.T) {
	batch := &Batch{
		operations: []BatchOperation{
			{Type: BatchOpWrite, Path: "file.txt", Content: []byte("content")},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := batch.Commit(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Commit() error = %v, want context.Canceled", err)
	}
}

func TestBatchOperationsIsolation(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	_ = batch.Write("file.txt", []byte("content"))

	// Get operations and modify the returned slice
	ops := batch.Operations()
	ops[0].Path = "modified.txt"

	// Original should be unchanged
	originalOps := batch.Operations()
	if originalOps[0].Path != "file.txt" {
		t.Errorf("Operations() was modified, path = %q, want %q", originalOps[0].Path, "file.txt")
	}
}

func TestWithCommitAuthor(t *testing.T) {
	batch := &Batch{
		operations: make([]BatchOperation, 0),
	}

	opt := WithCommitAuthor("Test User", "test@example.com")
	opt(batch)

	if batch.author == nil {
		t.Fatal("author is nil")
	}
	if batch.author.GetName() != "Test User" {
		t.Errorf("author.Name = %q, want %q", batch.author.GetName(), "Test User")
	}
	if batch.author.GetEmail() != "test@example.com" {
		t.Errorf("author.Email = %q, want %q", batch.author.GetEmail(), "test@example.com")
	}
}

func TestNewBatch(t *testing.T) {
	ctx := context.Background()

	// Test with default message
	batch, err := NewBatch(ctx, nil, "owner", "repo", "main", "")
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if batch.message != "Batch update" {
		t.Errorf("message = %q, want %q", batch.message, "Batch update")
	}

	// Test with custom message
	batch, err = NewBatch(ctx, nil, "owner", "repo", "main", "Custom message")
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if batch.message != "Custom message" {
		t.Errorf("message = %q, want %q", batch.message, "Custom message")
	}

	// Test with author option
	batch, err = NewBatch(ctx, nil, "owner", "repo", "main", "Test", WithCommitAuthor("Test", "test@example.com"))
	if err != nil {
		t.Fatalf("NewBatch() error = %v", err)
	}
	if batch.author == nil {
		t.Error("author is nil")
	}
}

func TestNewBatchCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewBatch(ctx, nil, "owner", "repo", "main", "Test")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("NewBatch() error = %v, want context.Canceled", err)
	}
}

func TestBatchErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	batchErr := &BatchError{Op: "test", Err: innerErr}

	if !errors.Is(batchErr, innerErr) {
		t.Error("BatchError should unwrap to inner error")
	}

	msg := batchErr.Error()
	if msg != "batch test failed: inner error" {
		t.Errorf("Error() = %q, want %q", msg, "batch test failed: inner error")
	}
}
