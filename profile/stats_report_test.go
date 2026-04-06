package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMonthToQuarter(t *testing.T) {
	tests := []struct {
		month    int
		expected int
	}{
		{1, 1}, {2, 1}, {3, 1},
		{4, 2}, {5, 2}, {6, 2},
		{7, 3}, {8, 3}, {9, 3},
		{10, 4}, {11, 4}, {12, 4},
	}

	for _, tt := range tests {
		got := monthToQuarter(tt.month)
		if got != tt.expected {
			t.Errorf("monthToQuarter(%d) = %d, want %d", tt.month, got, tt.expected)
		}
	}
}

func TestAggregateStatsAdd(t *testing.T) {
	a := AggregateStats{
		Commits:              10,
		Issues:               5,
		PRs:                  3,
		Reviews:              2,
		Releases:             1,
		Additions:            100,
		Deletions:            50,
		NetAdditions:         50,
		RepoCountContributed: 5,
		RepoCountCreated:     1,
	}

	b := AggregateStats{
		Commits:              20,
		Issues:               10,
		PRs:                  6,
		Reviews:              4,
		Releases:             2,
		Additions:            200,
		Deletions:            100,
		NetAdditions:         100,
		RepoCountContributed: 10,
		RepoCountCreated:     2,
	}

	a.Add(b)

	if a.Commits != 30 {
		t.Errorf("Commits = %d, want 30", a.Commits)
	}
	if a.Issues != 15 {
		t.Errorf("Issues = %d, want 15", a.Issues)
	}
	if a.PRs != 9 {
		t.Errorf("PRs = %d, want 9", a.PRs)
	}
	if a.Reviews != 6 {
		t.Errorf("Reviews = %d, want 6", a.Reviews)
	}
	if a.Releases != 3 {
		t.Errorf("Releases = %d, want 3", a.Releases)
	}
	if a.Additions != 300 {
		t.Errorf("Additions = %d, want 300", a.Additions)
	}
	if a.Deletions != 150 {
		t.Errorf("Deletions = %d, want 150", a.Deletions)
	}
	if a.NetAdditions != 150 {
		t.Errorf("NetAdditions = %d, want 150", a.NetAdditions)
	}
	if a.RepoCountContributed != 15 {
		t.Errorf("RepoCountContributed = %d, want 15", a.RepoCountContributed)
	}
	if a.RepoCountCreated != 3 {
		t.Errorf("RepoCountCreated = %d, want 3", a.RepoCountCreated)
	}
}

func TestBuildStatsReport(t *testing.T) {
	files := []MonthlyOutputFile{
		{
			Metadata: QueryMetadata{
				Username:   "testuser",
				From:       "2026-01-01",
				To:         "2026-01-31",
				Visibility: "public",
			},
			Username:  "testuser",
			Year:      2026,
			Month:     1,
			MonthName: "January",
			Stats: MonthlyStats{
				Year:                 2026,
				Month:                1,
				MonthName:            "January",
				Commits:              100,
				Releases:             10,
				Additions:            1000,
				Deletions:            500,
				NetAdditions:         500,
				RepoCountContributed: 5,
			},
		},
		{
			Metadata: QueryMetadata{
				Username:   "testuser",
				From:       "2026-02-01",
				To:         "2026-02-28",
				Visibility: "public",
			},
			Username:  "testuser",
			Year:      2026,
			Month:     2,
			MonthName: "February",
			Stats: MonthlyStats{
				Year:                 2026,
				Month:                2,
				MonthName:            "February",
				Commits:              150,
				Releases:             15,
				Additions:            1500,
				Deletions:            700,
				NetAdditions:         800,
				RepoCountContributed: 8,
			},
		},
		{
			Metadata: QueryMetadata{
				Username:   "testuser",
				From:       "2026-03-01",
				To:         "2026-03-31",
				Visibility: "public",
			},
			Username:  "testuser",
			Year:      2026,
			Month:     3,
			MonthName: "March",
			Stats: MonthlyStats{
				Year:                 2026,
				Month:                3,
				MonthName:            "March",
				Commits:              200,
				Releases:             20,
				Additions:            2000,
				Deletions:            1000,
				NetAdditions:         1000,
				RepoCountContributed: 10,
			},
		},
	}

	report, err := BuildStatsReport(files)
	if err != nil {
		t.Fatalf("BuildStatsReport() error = %v", err)
	}

	// Check metadata
	if report.Metadata.Username != "testuser" {
		t.Errorf("Username = %q, want %q", report.Metadata.Username, "testuser")
	}
	if report.Metadata.Visibility != "public" {
		t.Errorf("Visibility = %q, want %q", report.Metadata.Visibility, "public")
	}
	if report.Metadata.DataRange.From != "2026-01-01" {
		t.Errorf("DataRange.From = %q, want %q", report.Metadata.DataRange.From, "2026-01-01")
	}
	if report.Metadata.DataRange.To != "2026-03-31" {
		t.Errorf("DataRange.To = %q, want %q", report.Metadata.DataRange.To, "2026-03-31")
	}

	// Check years
	if len(report.Years) != 1 {
		t.Fatalf("len(Years) = %d, want 1", len(report.Years))
	}

	year := report.Years[0]
	if year.Year != 2026 {
		t.Errorf("Year = %d, want 2026", year.Year)
	}

	// Check year stats (sum of all months)
	if year.Stats.Commits != 450 {
		t.Errorf("Year.Stats.Commits = %d, want 450", year.Stats.Commits)
	}
	if year.Stats.Releases != 45 {
		t.Errorf("Year.Stats.Releases = %d, want 45", year.Stats.Releases)
	}
	if year.Stats.Additions != 4500 {
		t.Errorf("Year.Stats.Additions = %d, want 4500", year.Stats.Additions)
	}
	if year.Stats.Deletions != 2200 {
		t.Errorf("Year.Stats.Deletions = %d, want 2200", year.Stats.Deletions)
	}
	if year.Stats.NetAdditions != 2300 {
		t.Errorf("Year.Stats.NetAdditions = %d, want 2300", year.Stats.NetAdditions)
	}

	// Check quarters
	if len(year.Quarters) != 1 {
		t.Fatalf("len(Quarters) = %d, want 1", len(year.Quarters))
	}

	quarter := year.Quarters[0]
	if quarter.Quarter != 1 {
		t.Errorf("Quarter = %d, want 1", quarter.Quarter)
	}
	if quarter.Label != "Q1 2026" {
		t.Errorf("Label = %q, want %q", quarter.Label, "Q1 2026")
	}

	// Check quarter stats
	if quarter.Stats.Commits != 450 {
		t.Errorf("Quarter.Stats.Commits = %d, want 450", quarter.Stats.Commits)
	}

	// Check months
	if len(quarter.Months) != 3 {
		t.Fatalf("len(Months) = %d, want 3", len(quarter.Months))
	}

	// Months should be sorted chronologically
	if quarter.Months[0].MonthName != "January" {
		t.Errorf("Months[0].MonthName = %q, want %q", quarter.Months[0].MonthName, "January")
	}
	if quarter.Months[1].MonthName != "February" {
		t.Errorf("Months[1].MonthName = %q, want %q", quarter.Months[1].MonthName, "February")
	}
	if quarter.Months[2].MonthName != "March" {
		t.Errorf("Months[2].MonthName = %q, want %q", quarter.Months[2].MonthName, "March")
	}
}

func TestBuildStatsReportEmpty(t *testing.T) {
	_, err := BuildStatsReport(nil)
	if err == nil {
		t.Error("BuildStatsReport(nil) should return error")
	}

	_, err = BuildStatsReport([]MonthlyOutputFile{})
	if err == nil {
		t.Error("BuildStatsReport([]) should return error")
	}
}

func TestBuildStatsReportMultipleYears(t *testing.T) {
	files := []MonthlyOutputFile{
		{
			Metadata: QueryMetadata{Username: "testuser", From: "2025-12-01", To: "2025-12-31", Visibility: "public"},
			Username: "testuser", Year: 2025, Month: 12, MonthName: "December",
			Stats: MonthlyStats{Commits: 100, Releases: 10},
		},
		{
			Metadata: QueryMetadata{Username: "testuser", From: "2026-01-01", To: "2026-01-31", Visibility: "public"},
			Username: "testuser", Year: 2026, Month: 1, MonthName: "January",
			Stats: MonthlyStats{Commits: 150, Releases: 15},
		},
	}

	report, err := BuildStatsReport(files)
	if err != nil {
		t.Fatalf("BuildStatsReport() error = %v", err)
	}

	if len(report.Years) != 2 {
		t.Fatalf("len(Years) = %d, want 2", len(report.Years))
	}

	// Years should be sorted
	if report.Years[0].Year != 2025 {
		t.Errorf("Years[0].Year = %d, want 2025", report.Years[0].Year)
	}
	if report.Years[1].Year != 2026 {
		t.Errorf("Years[1].Year = %d, want 2026", report.Years[1].Year)
	}
}

func TestStatsReportGetters(t *testing.T) {
	report := &StatsReport{
		Metadata: ReportMetadata{Username: "testuser"},
		Years: []YearStats{
			{
				Year:  2025,
				Stats: AggregateStats{Commits: 100},
				Quarters: []QuarterStats{
					{Quarter: 4, Year: 2025, Label: "Q4 2025", Stats: AggregateStats{Commits: 100}},
				},
			},
			{
				Year:  2026,
				Stats: AggregateStats{Commits: 200},
				Quarters: []QuarterStats{
					{Quarter: 1, Year: 2026, Label: "Q1 2026", Stats: AggregateStats{Commits: 200}},
				},
			},
		},
	}

	// Test GetLatestQuarter
	latest := report.GetLatestQuarter()
	if latest == nil {
		t.Fatal("GetLatestQuarter() returned nil")
	}
	if latest.Label != "Q1 2026" {
		t.Errorf("GetLatestQuarter().Label = %q, want %q", latest.Label, "Q1 2026")
	}

	// Test GetQuarter
	q4 := report.GetQuarter(2025, 4)
	if q4 == nil {
		t.Fatal("GetQuarter(2025, 4) returned nil")
	}
	if q4.Label != "Q4 2025" {
		t.Errorf("GetQuarter(2025, 4).Label = %q, want %q", q4.Label, "Q4 2025")
	}

	// Test GetQuarter not found
	notFound := report.GetQuarter(2025, 1)
	if notFound != nil {
		t.Error("GetQuarter(2025, 1) should return nil")
	}

	// Test GetYear
	y2026 := report.GetYear(2026)
	if y2026 == nil {
		t.Fatal("GetYear(2026) returned nil")
	}
	if y2026.Stats.Commits != 200 {
		t.Errorf("GetYear(2026).Stats.Commits = %d, want 200", y2026.Stats.Commits)
	}

	// Test GetYear not found
	notFoundYear := report.GetYear(2024)
	if notFoundYear != nil {
		t.Error("GetYear(2024) should return nil")
	}

	// Test TotalStats
	total := report.TotalStats()
	if total.Commits != 300 {
		t.Errorf("TotalStats().Commits = %d, want 300", total.Commits)
	}
}

func TestLoadAndWriteStatsReport(t *testing.T) {
	tmpDir := t.TempDir()

	report := &StatsReport{
		Metadata: ReportMetadata{
			Username:    "testuser",
			Visibility:  "public",
			GeneratedAt: time.Now().UTC(),
			DataRange:   DateRange{From: "2026-01-01", To: "2026-03-31"},
		},
		Years: []YearStats{
			{
				Year:  2026,
				Stats: AggregateStats{Commits: 100, Releases: 10},
				Quarters: []QuarterStats{
					{
						Quarter: 1,
						Year:    2026,
						Label:   "Q1 2026",
						Stats:   AggregateStats{Commits: 100, Releases: 10},
						Months: []MonthStats{
							{Year: 2026, Month: 1, MonthName: "January", Stats: AggregateStats{Commits: 100, Releases: 10}},
						},
					},
				},
			},
		},
	}

	// Write report
	path := filepath.Join(tmpDir, "report.json")
	if err := WriteStatsReport(path, report); err != nil {
		t.Fatalf("WriteStatsReport() error = %v", err)
	}

	// Load report
	loaded, err := LoadStatsReport(path)
	if err != nil {
		t.Fatalf("LoadStatsReport() error = %v", err)
	}

	// Verify
	if loaded.Metadata.Username != report.Metadata.Username {
		t.Errorf("Username = %q, want %q", loaded.Metadata.Username, report.Metadata.Username)
	}
	if len(loaded.Years) != len(report.Years) {
		t.Errorf("len(Years) = %d, want %d", len(loaded.Years), len(report.Years))
	}
	if loaded.Years[0].Stats.Commits != report.Years[0].Stats.Commits {
		t.Errorf("Commits = %d, want %d", loaded.Years[0].Stats.Commits, report.Years[0].Stats.Commits)
	}
}

func TestLoadMonthlyFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test monthly files
	files := []MonthlyOutputFile{
		{
			Metadata: QueryMetadata{Username: "testuser", From: "2026-01-01", To: "2026-01-31", Visibility: "public"},
			Username: "testuser", Year: 2026, Month: 1, MonthName: "January",
			Stats: MonthlyStats{Commits: 100},
		},
		{
			Metadata: QueryMetadata{Username: "testuser", From: "2026-02-01", To: "2026-02-28", Visibility: "public"},
			Username: "testuser", Year: 2026, Month: 2, MonthName: "February",
			Stats: MonthlyStats{Commits: 150},
		},
	}

	for _, f := range files {
		filename := filepath.Join(tmpDir, "testuser_github_public_2026-"+padMonth(f.Month)+".json")
		data, err := json.MarshalIndent(f, "", "  ")
		if err != nil {
			t.Fatalf("json.MarshalIndent() error = %v", err)
		}
		if err := os.WriteFile(filename, data, 0600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}
	}

	// Also create a non-monthly file that should be skipped
	if err := os.WriteFile(filepath.Join(tmpDir, "report.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	// Load files
	loaded, err := LoadMonthlyFiles(tmpDir)
	if err != nil {
		t.Fatalf("LoadMonthlyFiles() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("len(loaded) = %d, want 2", len(loaded))
	}

	// Should be sorted chronologically
	if loaded[0].Month != 1 {
		t.Errorf("loaded[0].Month = %d, want 1", loaded[0].Month)
	}
	if loaded[1].Month != 2 {
		t.Errorf("loaded[1].Month = %d, want 2", loaded[1].Month)
	}
}

func padMonth(m int) string {
	return fmt.Sprintf("%02d", m)
}

func TestStatsFromMonthlyStats(t *testing.T) {
	ms := MonthlyStats{
		Year:                 2026,
		Month:                1,
		MonthName:            "January",
		Commits:              100,
		Issues:               5,
		PRs:                  10,
		Reviews:              3,
		Releases:             8,
		Additions:            1000,
		Deletions:            500,
		NetAdditions:         500,
		RepoCountContributed: 15,
		RepoCountCreated:     2,
	}

	as := statsFromMonthlyStats(ms)

	if as.Commits != 100 {
		t.Errorf("Commits = %d, want 100", as.Commits)
	}
	if as.Issues != 5 {
		t.Errorf("Issues = %d, want 5", as.Issues)
	}
	if as.PRs != 10 {
		t.Errorf("PRs = %d, want 10", as.PRs)
	}
	if as.Reviews != 3 {
		t.Errorf("Reviews = %d, want 3", as.Reviews)
	}
	if as.Releases != 8 {
		t.Errorf("Releases = %d, want 8", as.Releases)
	}
	if as.Additions != 1000 {
		t.Errorf("Additions = %d, want 1000", as.Additions)
	}
	if as.Deletions != 500 {
		t.Errorf("Deletions = %d, want 500", as.Deletions)
	}
	if as.NetAdditions != 500 {
		t.Errorf("NetAdditions = %d, want 500", as.NetAdditions)
	}
	if as.RepoCountContributed != 15 {
		t.Errorf("RepoCountContributed = %d, want 15", as.RepoCountContributed)
	}
	if as.RepoCountCreated != 2 {
		t.Errorf("RepoCountCreated = %d, want 2", as.RepoCountCreated)
	}
}
