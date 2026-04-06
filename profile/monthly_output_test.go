package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grokify/gogithub/graphql"
)

func TestWriteMonthlyFile(t *testing.T) {
	dir := t.TempDir()

	stats := MonthlyStats{
		Year:                 2024,
		Month:                3,
		MonthName:            "March",
		Commits:              50,
		Issues:               10,
		PRs:                  5,
		Reviews:              15,
		Releases:             3,
		Additions:            1000,
		Deletions:            300,
		NetAdditions:         700,
		RepoCountContributed: 2,
		RepoCountCreated:     1,
	}

	meta := QueryMetadata{
		Username:        "testuser",
		From:            "2024-03-01",
		To:              "2024-03-31",
		Visibility:      "public",
		IncludeReleases: true,
		GeneratedAt:     time.Now().UTC(),
	}

	path, err := WriteMonthlyFile(dir, "testuser", stats, meta)
	if err != nil {
		t.Fatalf("WriteMonthlyFile failed: %v", err)
	}

	expectedFilename := "testuser_github_2024-03.json"
	if filepath.Base(path) != expectedFilename {
		t.Errorf("filename = %q, want %q", filepath.Base(path), expectedFilename)
	}

	// Verify file contents
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var output MonthlyOutputFile
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if output.Username != "testuser" {
		t.Errorf("Username = %q, want %q", output.Username, "testuser")
	}
	if output.Year != 2024 {
		t.Errorf("Year = %d, want 2024", output.Year)
	}
	if output.Month != 3 {
		t.Errorf("Month = %d, want 3", output.Month)
	}
	if output.Stats.Commits != 50 {
		t.Errorf("Stats.Commits = %d, want 50", output.Stats.Commits)
	}
	if output.Stats.NetAdditions != 700 {
		t.Errorf("Stats.NetAdditions = %d, want 700", output.Stats.NetAdditions)
	}
	// Verify metadata
	if output.Metadata.Visibility != "public" {
		t.Errorf("Metadata.Visibility = %q, want %q", output.Metadata.Visibility, "public")
	}
}

func TestWriteMonthlyFiles(t *testing.T) {
	dir := t.TempDir()

	profile := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{
					Year:          2024,
					Month:         time.January,
					Commits:       100,
					Additions:     500,
					Deletions:     200,
					CommitsByRepo: map[string]int{"repo1": 100},
				},
				{
					Year:          2024,
					Month:         time.February,
					Commits:       50,
					Additions:     200,
					Deletions:     100,
					CommitsByRepo: map[string]int{"repo1": 30, "repo2": 20},
				},
			},
		},
	}

	opts := &Options{
		Visibility:      graphql.VisibilityPublic,
		IncludeReleases: false,
	}

	written, err := WriteMonthlyFiles(dir, profile, opts)
	if err != nil {
		t.Fatalf("WriteMonthlyFiles failed: %v", err)
	}

	if len(written) != 2 {
		t.Fatalf("WriteMonthlyFiles wrote %d files, want 2", len(written))
	}

	// Check files exist
	for _, path := range written {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", path)
		}
	}
}

func TestWriteMonthlyFilesNoActivity(t *testing.T) {
	dir := t.TempDir()

	profile := &UserProfile{
		Username: "testuser",
		Activity: nil,
	}

	_, err := WriteMonthlyFiles(dir, profile, nil)
	if err == nil {
		t.Error("WriteMonthlyFiles should fail with no activity data")
	}
}

func TestWriteMonthlyMultiFile(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "monthly.json")

	profile := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{
					Year:          2024,
					Month:         time.January,
					Commits:       100,
					Additions:     500,
					Deletions:     200,
					CommitsByRepo: map[string]int{"repo1": 100},
				},
			},
		},
	}

	opts := &Options{
		Visibility:      graphql.VisibilityAll,
		IncludeReleases: true,
		ReleaseOrgs:     []string{"grokify"},
	}

	err := WriteMonthlyMultiFile(outputPath, profile, opts)
	if err != nil {
		t.Fatalf("WriteMonthlyMultiFile failed: %v", err)
	}

	// Verify file contents
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var output MonthlyOutputMulti
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if output.Username != "testuser" {
		t.Errorf("Username = %q, want %q", output.Username, "testuser")
	}
	if len(output.Months) != 1 {
		t.Fatalf("Months length = %d, want 1", len(output.Months))
	}
	if output.Months[0].Commits != 100 {
		t.Errorf("Months[0].Commits = %d, want 100", output.Months[0].Commits)
	}
	// Verify metadata
	if output.Metadata.Visibility != "all" {
		t.Errorf("Metadata.Visibility = %q, want %q", output.Metadata.Visibility, "all")
	}
	if !output.Metadata.IncludeReleases {
		t.Error("Metadata.IncludeReleases should be true")
	}
	if len(output.Metadata.ReleaseOrgs) != 1 || output.Metadata.ReleaseOrgs[0] != "grokify" {
		t.Errorf("Metadata.ReleaseOrgs = %v, want [grokify]", output.Metadata.ReleaseOrgs)
	}
}

func TestWriteMonthlyMultiFileMerge(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "monthly.json")

	// Write first month
	profile1 := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{
					Year:          2024,
					Month:         time.January,
					Commits:       100,
					Additions:     500,
					Deletions:     200,
					CommitsByRepo: map[string]int{"repo1": 100},
				},
			},
		},
	}

	err := WriteMonthlyMultiFile(outputPath, profile1, nil)
	if err != nil {
		t.Fatalf("first WriteMonthlyMultiFile failed: %v", err)
	}

	// Write second month (should merge)
	profile2 := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{
					Year:          2024,
					Month:         time.February,
					Commits:       50,
					Additions:     200,
					Deletions:     100,
					CommitsByRepo: map[string]int{"repo1": 30, "repo2": 20},
				},
			},
		},
	}

	err = WriteMonthlyMultiFile(outputPath, profile2, nil)
	if err != nil {
		t.Fatalf("second WriteMonthlyMultiFile failed: %v", err)
	}

	// Verify merged contents
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var output MonthlyOutputMulti
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(output.Months) != 2 {
		t.Fatalf("Months length = %d, want 2", len(output.Months))
	}

	// Should be sorted descending (Feb first)
	if output.Months[0].Month != 2 {
		t.Errorf("Months[0].Month = %d, want 2 (February)", output.Months[0].Month)
	}
	if output.Months[1].Month != 1 {
		t.Errorf("Months[1].Month = %d, want 1 (January)", output.Months[1].Month)
	}
}

func TestWriteMonthlyMultiFileOverwrite(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "monthly.json")

	// Write January with 100 commits
	profile1 := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{Year: 2024, Month: time.January, Commits: 100, CommitsByRepo: map[string]int{"repo1": 100}},
			},
		},
	}

	err := WriteMonthlyMultiFile(outputPath, profile1, nil)
	if err != nil {
		t.Fatalf("first WriteMonthlyMultiFile failed: %v", err)
	}

	// Write January again with updated data (150 commits)
	profile2 := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{Year: 2024, Month: time.January, Commits: 150, CommitsByRepo: map[string]int{"repo1": 150}},
			},
		},
	}

	err = WriteMonthlyMultiFile(outputPath, profile2, nil)
	if err != nil {
		t.Fatalf("second WriteMonthlyMultiFile failed: %v", err)
	}

	// Verify overwritten contents
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var output MonthlyOutputMulti
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(output.Months) != 1 {
		t.Fatalf("Months length = %d, want 1", len(output.Months))
	}
	if output.Months[0].Commits != 150 {
		t.Errorf("Months[0].Commits = %d, want 150 (should be overwritten)", output.Months[0].Commits)
	}
}

func TestGetMonthlyFileCount(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "monthly.json")

	profile := &UserProfile{
		Username: "testuser",
		From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
		Activity: &ActivityTimeline{
			Months: []MonthlyActivity{
				{Year: 2024, Month: time.January, Commits: 100, CommitsByRepo: map[string]int{"repo1": 100}},
				{Year: 2024, Month: time.February, Commits: 50, CommitsByRepo: map[string]int{"repo1": 50}},
				{Year: 2024, Month: time.March, Commits: 75, CommitsByRepo: map[string]int{"repo1": 75}},
			},
		},
	}

	err := WriteMonthlyMultiFile(outputPath, profile, nil)
	if err != nil {
		t.Fatalf("WriteMonthlyMultiFile failed: %v", err)
	}

	count, err := GetMonthlyFileCount(outputPath)
	if err != nil {
		t.Fatalf("GetMonthlyFileCount failed: %v", err)
	}

	if count != 3 {
		t.Errorf("GetMonthlyFileCount = %d, want 3", count)
	}
}

func TestNewQueryMetadata(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	opts := &Options{
		Visibility:      graphql.VisibilityPublic,
		IncludeReleases: true,
		ReleaseOrgs:     []string{"grokify", "plexusone"},
	}

	meta := NewQueryMetadata("testuser", from, to, opts)

	if meta.Username != "testuser" {
		t.Errorf("Username = %q, want %q", meta.Username, "testuser")
	}
	if meta.From != "2024-01-01" {
		t.Errorf("From = %q, want %q", meta.From, "2024-01-01")
	}
	if meta.To != "2024-03-31" {
		t.Errorf("To = %q, want %q", meta.To, "2024-03-31")
	}
	if meta.Visibility != "public" {
		t.Errorf("Visibility = %q, want %q", meta.Visibility, "public")
	}
	if !meta.IncludeReleases {
		t.Error("IncludeReleases should be true")
	}
	if len(meta.ReleaseOrgs) != 2 {
		t.Errorf("ReleaseOrgs length = %d, want 2", len(meta.ReleaseOrgs))
	}
}

func TestNewQueryMetadataNilOpts(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	meta := NewQueryMetadata("testuser", from, to, nil)

	if meta.Visibility != "all" {
		t.Errorf("Visibility = %q, want %q (default)", meta.Visibility, "all")
	}
	if meta.IncludeReleases {
		t.Error("IncludeReleases should be false (default)")
	}
}
