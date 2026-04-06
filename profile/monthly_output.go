package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/grokify/gogithub/graphql"
)

// QueryMetadata captures the parameters used to generate the output.
// This enables reproducibility and consistent generation of additional data.
type QueryMetadata struct {
	// Query parameters
	Username    string   `json:"username"`
	From        string   `json:"from"`                  // YYYY-MM-DD format
	To          string   `json:"to"`                    // YYYY-MM-DD format
	Visibility  string   `json:"visibility"`            // all, public, private
	ReleaseOrgs []string `json:"releaseOrgs,omitempty"` // orgs to filter releases

	// Feature flags
	IncludeReleases bool `json:"includeReleases"`

	// Generation info
	GeneratedAt time.Time `json:"generatedAt"`
	Command     string    `json:"command,omitempty"` // CLI command used (optional)
}

// NewQueryMetadata creates QueryMetadata from Options and profile data.
func NewQueryMetadata(username string, from, to time.Time, opts *Options) QueryMetadata {
	visibility := "all"
	if opts != nil {
		switch opts.Visibility {
		case graphql.VisibilityPublic:
			visibility = "public"
		case graphql.VisibilityPrivate:
			visibility = "private"
		default:
			visibility = "all"
		}
	}

	meta := QueryMetadata{
		Username:    username,
		From:        from.Format("2006-01-02"),
		To:          to.Format("2006-01-02"),
		Visibility:  visibility,
		GeneratedAt: time.Now().UTC(),
	}

	if opts != nil {
		meta.IncludeReleases = opts.IncludeReleases
		meta.ReleaseOrgs = opts.ReleaseOrgs
	}

	return meta
}

// MonthlyOutputFile is the structure for a single monthly output file.
type MonthlyOutputFile struct {
	Metadata  QueryMetadata `json:"metadata"`
	Username  string        `json:"username"`
	Year      int           `json:"year"`
	Month     int           `json:"month"`
	MonthName string        `json:"monthName"`
	Stats     MonthlyStats  `json:"stats"`
}

// MonthlyOutputMulti is the structure for a combined monthly output file.
type MonthlyOutputMulti struct {
	Metadata QueryMetadata  `json:"metadata"`
	Username string         `json:"username"`
	Months   []MonthlyStats `json:"months"`
}

// WriteMonthlyFile writes a single month's stats to a file.
// Filename format: {username}_github_{YYYY-MM}.json
func WriteMonthlyFile(dir string, username string, stats MonthlyStats, meta QueryMetadata) (string, error) {
	filename := fmt.Sprintf("%s_github_%04d-%02d.json", username, stats.Year, stats.Month)
	fp := filepath.Join(dir, filename)

	output := MonthlyOutputFile{
		Metadata:  meta,
		Username:  username,
		Year:      stats.Year,
		Month:     stats.Month,
		MonthName: stats.MonthName,
		Stats:     stats,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal monthly JSON: %w", err)
	}

	if err := os.WriteFile(fp, append(data, '\n'), 0600); err != nil {
		return "", fmt.Errorf("write monthly file: %w", err)
	}

	return fp, nil
}

// WriteMonthlyFiles writes individual files for each month in the profile.
// Returns a list of written file paths.
func WriteMonthlyFiles(dir string, p *UserProfile, opts *Options) ([]string, error) {
	if p.Activity == nil || len(p.Activity.Months) == 0 {
		return nil, errors.New("no monthly activity data available")
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	var written []string
	for _, month := range p.Activity.Months {
		stats := month.ToMonthlyStats()
		// Create metadata for this specific month
		monthStart := time.Date(month.Year, month.Month, 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, -1) // Last day of month
		meta := NewQueryMetadata(p.Username, monthStart, monthEnd, opts)
		path, err := WriteMonthlyFile(dir, p.Username, stats, meta)
		if err != nil {
			return written, err
		}
		written = append(written, path)
	}

	return written, nil
}

// WriteMonthlyMultiFile writes all months to a single file with merge support.
// If the file exists, new months are merged with existing data (overwriting duplicates).
// Months are sorted in descending order (newest first).
func WriteMonthlyMultiFile(fp string, p *UserProfile, opts *Options) error {
	if p.Activity == nil || len(p.Activity.Months) == 0 {
		return errors.New("no monthly activity data available")
	}

	newMonths := p.Activity.ToMonthlyStats()

	// Try to read existing file
	var existing MonthlyOutputMulti
	existingData, err := os.ReadFile(fp)
	if err == nil {
		if err := json.Unmarshal(existingData, &existing); err != nil {
			return fmt.Errorf("parse existing monthly file: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing monthly file: %w", err)
	}

	// Merge: create map of existing months, then update/add new months
	monthMap := make(map[string]MonthlyStats)
	for _, m := range existing.Months {
		key := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
		monthMap[key] = m
	}
	for _, m := range newMonths {
		key := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
		monthMap[key] = m // Overwrites existing or adds new
	}

	// Convert map back to slice
	merged := make([]MonthlyStats, 0, len(monthMap))
	for _, m := range monthMap {
		merged = append(merged, m)
	}

	// Sort descending by year, then month (newest first)
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Year != merged[j].Year {
			return merged[i].Year > merged[j].Year
		}
		return merged[i].Month > merged[j].Month
	})

	// Create metadata from profile date range
	meta := NewQueryMetadata(p.Username, p.From, p.To, opts)

	// Build output
	output := MonthlyOutputMulti{
		Metadata: meta,
		Username: p.Username,
		Months:   merged,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal monthly JSON: %w", err)
	}

	if err := os.WriteFile(fp, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write monthly file: %w", err)
	}

	return nil
}

// GetMonthlyFileCount returns the count of merged months for reporting.
func GetMonthlyFileCount(filepath string) (int, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return 0, err
	}

	var output MonthlyOutputMulti
	if err := json.Unmarshal(data, &output); err != nil {
		return 0, err
	}

	return len(output.Months), nil
}
