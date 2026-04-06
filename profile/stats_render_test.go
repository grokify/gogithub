package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{999, "999"},
		{1000, "1,000"},
		{1234, "1,234"},
		{10000, "10,000"},
		{100000, "100,000"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
		{-1, "-1"},
		{-1000, "-1,000"},
		{-1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.expected {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestFormatSignedNumber(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "+0"},
		{1, "+1"},
		{1000, "+1,000"},
		{-1, "-1"},
		{-1000, "-1,000"},
	}

	for _, tt := range tests {
		got := formatSignedNumber(tt.n)
		if got != tt.expected {
			t.Errorf("formatSignedNumber(%d) = %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		s        string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"public", "Public"},
		{"PRIVATE", "PRIVATE"},
	}

	for _, tt := range tests {
		got := capitalize(tt.s)
		if got != tt.expected {
			t.Errorf("capitalize(%q) = %q, want %q", tt.s, got, tt.expected)
		}
	}
}

func TestRenderToMarkdown(t *testing.T) {
	report := createTestReport()

	opts := RenderOptions{
		Title:            "Test Report",
		ShowMonthDetails: true,
		ShowDataSource:   true,
		DataSourceURL:    "https://example.com",
		RawDataFiles:     []string{"test_2026-01.json"},
		RegenerateCmd:    "gogithub profile --user test",
	}

	md, err := RenderToMarkdown(report, opts)
	if err != nil {
		t.Fatalf("RenderToMarkdown() error = %v", err)
	}

	// Check that key content is present
	expectedStrings := []string{
		"# Test Report",
		"Public repository contribution statistics.",
		"## Q1 2026 Summary",
		"| Commits | 450 |",
		"| Releases | 45 |",
		"### Commits",
		"| January 2026 | 100 |",
		"| February 2026 | 150 |",
		"| March 2026 | 200 |",
		"### Code Changes",
		"### Releases",
		"## Monthly Details",
		"### March 2026",
		"## Data Source",
		"[gogithub](https://example.com)",
		"test_2026-01.json",
		"gogithub profile --user test",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(md, expected) {
			t.Errorf("RenderToMarkdown() missing expected string: %q", expected)
		}
	}
}

func TestRenderToMarkdownDefaultTitle(t *testing.T) {
	report := createTestReport()

	opts := RenderOptions{}

	md, err := RenderToMarkdown(report, opts)
	if err != nil {
		t.Fatalf("RenderToMarkdown() error = %v", err)
	}

	expected := "# GitHub Statistics - testuser"
	if !strings.Contains(md, expected) {
		t.Errorf("RenderToMarkdown() missing default title: %q", expected)
	}
}

func TestRenderToHTML(t *testing.T) {
	report := createTestReport()

	opts := RenderOptions{
		Title: "HTML Test Report",
	}

	html, err := RenderToHTML(report, opts)
	if err != nil {
		t.Fatalf("RenderToHTML() error = %v", err)
	}

	expectedStrings := []string{
		"<!DOCTYPE html>",
		"<title>HTML Test Report</title>",
		"<h1>HTML Test Report</h1>",
		"<h2>Q1 2026 Summary</h2>",
		"<td>Commits</td>",
		"<td>450</td>",
		"January 2026",
		"February 2026",
		"March 2026",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(html, expected) {
			t.Errorf("RenderToHTML() missing expected string: %q", expected)
		}
	}
}

func TestRenderToText(t *testing.T) {
	report := createTestReport()

	opts := RenderOptions{
		Title: "Text Test Report",
	}

	text, err := RenderToText(report, opts)
	if err != nil {
		t.Fatalf("RenderToText() error = %v", err)
	}

	expectedStrings := []string{
		"Text Test Report",
		"================", // 16 chars matches title length
		"Q1 2026 Summary",
		"Commits:",
		"Releases:",
		"Additions:",
		"Deletions:",
		"Net Additions:",
		"Monthly Breakdown:",
		"January 2026:",
		"February 2026:",
		"March 2026:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(text, expected) {
			t.Errorf("RenderToText() missing expected string: %q", expected)
		}
	}
}

func TestRenderToFile(t *testing.T) {
	tmpDir := t.TempDir()
	report := createTestReport()

	tests := []struct {
		format   RenderFormat
		filename string
		contains string
	}{
		{RenderFormatMarkdown, "report.md", "# GitHub Statistics"},
		{RenderFormatHTML, "report.html", "<!DOCTYPE html>"},
		{RenderFormatText, "report.txt", "Q1 2026 Summary"},
	}

	for _, tt := range tests {
		path := filepath.Join(tmpDir, tt.filename)
		opts := RenderOptions{Format: tt.format}

		if err := RenderToFile(path, report, opts); err != nil {
			t.Errorf("RenderToFile(%q) error = %v", tt.filename, err)
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("os.ReadFile(%q) error = %v", path, err)
			continue
		}

		if !strings.Contains(string(content), tt.contains) {
			t.Errorf("RenderToFile(%q) missing expected content: %q", tt.filename, tt.contains)
		}
	}
}

func TestRenderToFileDefaultFormat(t *testing.T) {
	tmpDir := t.TempDir()
	report := createTestReport()
	path := filepath.Join(tmpDir, "report.md")

	// No format specified should default to Markdown
	opts := RenderOptions{}

	if err := RenderToFile(path, report, opts); err != nil {
		t.Fatalf("RenderToFile() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	if !strings.Contains(string(content), "# GitHub Statistics") {
		t.Error("RenderToFile() with no format should produce Markdown")
	}
}

func TestRenderOptionsDefault(t *testing.T) {
	opts := DefaultRenderOptions()

	if opts.Format != RenderFormatMarkdown {
		t.Errorf("Format = %q, want %q", opts.Format, RenderFormatMarkdown)
	}
	if !opts.ShowMonthDetails {
		t.Error("ShowMonthDetails should be true by default")
	}
	if !opts.ShowDataSource {
		t.Error("ShowDataSource should be true by default")
	}
}

// createTestReport creates a test report for rendering tests.
func createTestReport() *StatsReport {
	return &StatsReport{
		Metadata: ReportMetadata{
			Username:    "testuser",
			Visibility:  "public",
			GeneratedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			DataRange:   DateRange{From: "2026-01-01", To: "2026-03-31"},
		},
		Years: []YearStats{
			{
				Year: 2026,
				Stats: AggregateStats{
					Commits:      450,
					Releases:     45,
					Additions:    4500,
					Deletions:    2200,
					NetAdditions: 2300,
				},
				Quarters: []QuarterStats{
					{
						Quarter: 1,
						Year:    2026,
						Label:   "Q1 2026",
						Stats: AggregateStats{
							Commits:      450,
							Releases:     45,
							Additions:    4500,
							Deletions:    2200,
							NetAdditions: 2300,
						},
						Months: []MonthStats{
							{
								Year:      2026,
								Month:     1,
								MonthName: "January",
								Stats: AggregateStats{
									Commits:              100,
									Releases:             10,
									Additions:            1000,
									Deletions:            500,
									NetAdditions:         500,
									RepoCountContributed: 5,
								},
							},
							{
								Year:      2026,
								Month:     2,
								MonthName: "February",
								Stats: AggregateStats{
									Commits:              150,
									Releases:             15,
									Additions:            1500,
									Deletions:            700,
									NetAdditions:         800,
									RepoCountContributed: 8,
								},
							},
							{
								Year:      2026,
								Month:     3,
								MonthName: "March",
								Stats: AggregateStats{
									Commits:              200,
									Releases:             20,
									Additions:            2000,
									Deletions:            1000,
									NetAdditions:         1000,
									RepoCountContributed: 10,
								},
							},
						},
					},
				},
			},
		},
	}
}
