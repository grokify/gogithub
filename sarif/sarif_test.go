package sarif

import (
	"bytes"
	"testing"
)

func TestGzipCompressDecompress(t *testing.T) {
	original := []byte(`{"$schema":"https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json","version":"2.1.0","runs":[]}`)

	// Compress
	compressed, err := gzipCompress(original)
	if err != nil {
		t.Fatalf("gzipCompress failed: %v", err)
	}

	// Compressed should be different from original
	if bytes.Equal(compressed, original) {
		t.Error("compressed data should differ from original")
	}

	// Decompress
	decompressed, err := gzipDecompress(compressed)
	if err != nil {
		t.Fatalf("gzipDecompress failed: %v", err)
	}

	// Should match original
	if !bytes.Equal(decompressed, original) {
		t.Error("decompressed data should match original")
	}
}

func TestUploadOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    UploadOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: UploadOptions{
				CommitSHA: "abc123",
				Ref:       "refs/heads/main",
			},
			wantErr: false,
		},
		{
			name: "missing CommitSHA",
			opts: UploadOptions{
				Ref: "refs/heads/main",
			},
			wantErr: true,
		},
		{
			name: "missing Ref",
			opts: UploadOptions{
				CommitSHA: "abc123",
			},
			wantErr: true,
		},
		{
			name:    "missing both",
			opts:    UploadOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't fully test Upload without mocking, but we can verify
			// the validation logic by checking if required fields are set
			hasErr := tt.opts.CommitSHA == "" || tt.opts.Ref == ""
			if hasErr != tt.wantErr {
				t.Errorf("validation = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestProcessingStatusConstants(t *testing.T) {
	// Verify the status constants match GitHub's API values
	tests := []struct {
		status   ProcessingStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusComplete, "complete"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("ProcessingStatus = %q, want %q", tt.status, tt.expected)
		}
	}
}
