// Package sarif provides helpers for uploading SARIF files to GitHub Code Scanning.
package sarif

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/go-github/v84/github"
)

// UploadOptions configures the SARIF upload.
type UploadOptions struct {
	// CommitSHA is the SHA of the commit to associate the upload with.
	// Required.
	CommitSHA string

	// Ref is the Git reference (branch or tag) to associate the upload with.
	// For branches, use "refs/heads/<branch>". For tags, use "refs/tags/<tag>".
	// Required.
	Ref string

	// CheckoutURI is the URI to the root of the repository checkout.
	// Optional. Example: "file:///github/workspace/"
	CheckoutURI string

	// ToolName is the name of the tool that generated the SARIF file.
	// Optional. If not set, GitHub will extract it from the SARIF file.
	ToolName string

	// StartedAt is when the analysis started.
	// Optional. Defaults to current time.
	StartedAt *time.Time
}

// UploadResult contains the result of a SARIF upload.
type UploadResult struct {
	// SarifID is the identifier for the uploaded SARIF.
	SarifID string

	// URL is the API URL for checking upload status.
	URL string
}

// UploadFile reads a SARIF file, compresses it, and uploads to GitHub Code Scanning.
//
// The file is gzip-compressed and base64-encoded as required by the GitHub API.
//
// GitHub API docs: https://docs.github.com/rest/code-scanning/code-scanning#upload-an-analysis-as-sarif-data
func UploadFile(ctx context.Context, gh *github.Client, owner, repo, filePath string, opts UploadOptions) (*UploadResult, error) {
	// Read the SARIF file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SARIF file: %w", err)
	}

	return Upload(ctx, gh, owner, repo, data, opts)
}

// Upload uploads SARIF data to GitHub Code Scanning.
//
// The data is gzip-compressed and base64-encoded as required by the GitHub API.
//
// GitHub API docs: https://docs.github.com/rest/code-scanning/code-scanning#upload-an-analysis-as-sarif-data
func Upload(ctx context.Context, gh *github.Client, owner, repo string, sarifData []byte, opts UploadOptions) (*UploadResult, error) {
	if opts.CommitSHA == "" {
		return nil, fmt.Errorf("CommitSHA is required")
	}
	if opts.Ref == "" {
		return nil, fmt.Errorf("Ref is required")
	}

	// Compress with gzip
	compressed, err := gzipCompress(sarifData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress SARIF data: %w", err)
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(compressed)

	// Build the upload request
	analysis := &github.SarifAnalysis{
		CommitSHA: github.Ptr(opts.CommitSHA),
		Ref:       github.Ptr(opts.Ref),
		Sarif:     github.Ptr(encoded),
	}

	if opts.CheckoutURI != "" {
		analysis.CheckoutURI = github.Ptr(opts.CheckoutURI)
	}
	if opts.ToolName != "" {
		analysis.ToolName = github.Ptr(opts.ToolName)
	}
	if opts.StartedAt != nil {
		analysis.StartedAt = &github.Timestamp{Time: *opts.StartedAt}
	}

	// Upload
	sarifID, _, err := gh.CodeScanning.UploadSarif(ctx, owner, repo, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to upload SARIF: %w", err)
	}

	result := &UploadResult{
		SarifID: sarifID.GetID(),
		URL:     sarifID.GetURL(),
	}

	return result, nil
}

// ProcessingStatus represents the status of a SARIF upload.
type ProcessingStatus string

const (
	// StatusPending indicates the SARIF file is being processed.
	StatusPending ProcessingStatus = "pending"

	// StatusComplete indicates processing is complete.
	StatusComplete ProcessingStatus = "complete"

	// StatusFailed indicates processing failed.
	StatusFailed ProcessingStatus = "failed"
)

// UploadStatus contains information about a SARIF upload's processing status.
type UploadStatus struct {
	// Status is the processing status: "pending", "complete", or "failed".
	Status ProcessingStatus

	// AnalysesURL is the URL to fetch the analyses associated with the upload.
	// Only available when status is "complete".
	AnalysesURL string
}

// GetUploadStatus retrieves the processing status of a SARIF upload.
//
// GitHub API docs: https://docs.github.com/rest/code-scanning/code-scanning#get-information-about-a-sarif-upload
func GetUploadStatus(ctx context.Context, gh *github.Client, owner, repo, sarifID string) (*UploadStatus, error) {
	upload, _, err := gh.CodeScanning.GetSARIF(ctx, owner, repo, sarifID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SARIF status: %w", err)
	}

	return &UploadStatus{
		Status:      ProcessingStatus(upload.GetProcessingStatus()),
		AnalysesURL: upload.GetAnalysesURL(),
	}, nil
}

// WaitForProcessing polls the upload status until processing is complete or the context is canceled.
//
// pollInterval specifies how long to wait between status checks.
// Returns the final status when processing is complete or failed.
func WaitForProcessing(ctx context.Context, gh *github.Client, owner, repo, sarifID string, pollInterval time.Duration) (*UploadStatus, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := GetUploadStatus(ctx, gh, owner, repo, sarifID)
			if err != nil {
				return nil, err
			}

			if status.Status != StatusPending {
				return status, nil
			}
		}
	}
}

// UploadAndWait uploads a SARIF file and waits for processing to complete.
//
// This is a convenience function that combines UploadFile and WaitForProcessing.
// pollInterval specifies how long to wait between status checks (default: 5s).
func UploadAndWait(ctx context.Context, gh *github.Client, owner, repo, filePath string, opts UploadOptions, pollInterval time.Duration) (*UploadStatus, error) {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	result, err := UploadFile(ctx, gh, owner, repo, filePath, opts)
	if err != nil {
		return nil, err
	}

	return WaitForProcessing(ctx, gh, owner, repo, result.SarifID, pollInterval)
}

// gzipCompress compresses data using gzip.
func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// gzipDecompress decompresses gzip data.
func gzipDecompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}
