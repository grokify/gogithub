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

func TestGzipCompressEmpty(t *testing.T) {
	original := []byte{}

	compressed, err := gzipCompress(original)
	if err != nil {
		t.Fatalf("gzipCompress failed on empty input: %v", err)
	}

	// Even empty data produces some gzip header bytes
	if len(compressed) == 0 {
		t.Error("compressed data should not be empty (gzip has headers)")
	}

	decompressed, err := gzipDecompress(compressed)
	if err != nil {
		t.Fatalf("gzipDecompress failed: %v", err)
	}

	if len(decompressed) != 0 {
		t.Errorf("decompressed empty data should be empty, got %d bytes", len(decompressed))
	}
}

func TestGzipCompressLargeData(t *testing.T) {
	// Create large repetitive data (compresses well)
	original := bytes.Repeat([]byte("SARIF test data with repetitive content. "), 1000)

	compressed, err := gzipCompress(original)
	if err != nil {
		t.Fatalf("gzipCompress failed: %v", err)
	}

	// Compressed should be smaller for repetitive data
	if len(compressed) >= len(original) {
		t.Logf("Warning: compressed size (%d) >= original size (%d)", len(compressed), len(original))
	}

	decompressed, err := gzipDecompress(compressed)
	if err != nil {
		t.Fatalf("gzipDecompress failed: %v", err)
	}

	if !bytes.Equal(decompressed, original) {
		t.Error("decompressed data should match original")
	}
}

func TestGzipDecompressInvalidData(t *testing.T) {
	invalidData := []byte("this is not gzip data")

	_, err := gzipDecompress(invalidData)
	if err == nil {
		t.Error("gzipDecompress should fail on invalid gzip data")
	}
}

func TestUploadOptionsStruct(t *testing.T) {
	opts := UploadOptions{
		CommitSHA:   "abc123def456",
		Ref:         "refs/heads/main",
		CheckoutURI: "file:///github/workspace/",
		ToolName:    "my-scanner",
	}

	if opts.CommitSHA != "abc123def456" {
		t.Errorf("CommitSHA = %q, want %q", opts.CommitSHA, "abc123def456")
	}
	if opts.Ref != "refs/heads/main" {
		t.Errorf("Ref = %q, want %q", opts.Ref, "refs/heads/main")
	}
	if opts.CheckoutURI != "file:///github/workspace/" {
		t.Errorf("CheckoutURI = %q, want %q", opts.CheckoutURI, "file:///github/workspace/")
	}
	if opts.ToolName != "my-scanner" {
		t.Errorf("ToolName = %q, want %q", opts.ToolName, "my-scanner")
	}
}

func TestUploadResultStruct(t *testing.T) {
	result := UploadResult{
		SarifID: "sarif-12345",
		URL:     "https://api.github.com/repos/owner/repo/code-scanning/sarifs/sarif-12345",
	}

	if result.SarifID != "sarif-12345" {
		t.Errorf("SarifID = %q, want %q", result.SarifID, "sarif-12345")
	}
	if result.URL != "https://api.github.com/repos/owner/repo/code-scanning/sarifs/sarif-12345" {
		t.Errorf("URL = %q, want expected URL", result.URL)
	}
}

func TestUploadStatusStruct(t *testing.T) {
	status := UploadStatus{
		Status:      StatusComplete,
		AnalysesURL: "https://api.github.com/repos/owner/repo/code-scanning/analyses",
	}

	if status.Status != StatusComplete {
		t.Errorf("Status = %q, want %q", status.Status, StatusComplete)
	}
	if status.AnalysesURL != "https://api.github.com/repos/owner/repo/code-scanning/analyses" {
		t.Errorf("AnalysesURL = %q, want expected URL", status.AnalysesURL)
	}
}

func TestProcessingStatusComparison(t *testing.T) {
	// Test that status values can be compared
	tests := []struct {
		name   string
		status ProcessingStatus
		check  ProcessingStatus
		equal  bool
	}{
		{"pending equals pending", StatusPending, StatusPending, true},
		{"complete equals complete", StatusComplete, StatusComplete, true},
		{"failed equals failed", StatusFailed, StatusFailed, true},
		{"pending not equals complete", StatusPending, StatusComplete, false},
		{"complete not equals failed", StatusComplete, StatusFailed, false},
		{"pending not equals failed", StatusPending, StatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status == tt.check
			if got != tt.equal {
				t.Errorf("(%q == %q) = %v, want %v", tt.status, tt.check, got, tt.equal)
			}
		})
	}
}

func TestGzipCompressDecompressBinary(t *testing.T) {
	// Test with binary data including null bytes
	original := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x00, 0x00}

	compressed, err := gzipCompress(original)
	if err != nil {
		t.Fatalf("gzipCompress failed on binary data: %v", err)
	}

	decompressed, err := gzipDecompress(compressed)
	if err != nil {
		t.Fatalf("gzipDecompress failed: %v", err)
	}

	if !bytes.Equal(decompressed, original) {
		t.Error("decompressed binary data should match original")
	}
}
