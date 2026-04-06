package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// StatsReport is the top-level structure for aggregated statistics.
// It organizes data hierarchically: years -> quarters -> months.
type StatsReport struct {
	Metadata ReportMetadata `json:"metadata"`
	Years    []YearStats    `json:"years"`
}

// ReportMetadata contains information about the report generation.
type ReportMetadata struct {
	Username    string    `json:"username"`
	Visibility  string    `json:"visibility"`
	GeneratedAt time.Time `json:"generatedAt"`
	DataRange   DateRange `json:"dataRange"`
}

// DateRange represents a date range.
type DateRange struct {
	From string `json:"from"` // YYYY-MM-DD
	To   string `json:"to"`   // YYYY-MM-DD
}

// YearStats contains statistics for a single year.
type YearStats struct {
	Year     int            `json:"year"`
	Stats    AggregateStats `json:"stats"`
	Quarters []QuarterStats `json:"quarters"`
}

// QuarterStats contains statistics for a single quarter.
type QuarterStats struct {
	Quarter int            `json:"quarter"` // 1-4
	Year    int            `json:"year"`
	Label   string         `json:"label"` // e.g., "Q1 2026"
	Stats   AggregateStats `json:"stats"`
	Months  []MonthStats   `json:"months"`
}

// MonthStats contains statistics for a single month within a report.
type MonthStats struct {
	Year      int            `json:"year"`
	Month     int            `json:"month"`
	MonthName string         `json:"monthName"`
	Stats     AggregateStats `json:"stats"`
}

// AggregateStats contains the actual statistics that can be summed.
type AggregateStats struct {
	Commits              int `json:"commits"`
	Issues               int `json:"issues"`
	PRs                  int `json:"prs"`
	Reviews              int `json:"reviews"`
	Releases             int `json:"releases"`
	Additions            int `json:"additions"`
	Deletions            int `json:"deletions"`
	NetAdditions         int `json:"netAdditions"`
	RepoCountContributed int `json:"repoCountContributed"`
	RepoCountCreated     int `json:"repoCountCreated"`
}

// Add adds another AggregateStats to this one (for rollups).
func (a *AggregateStats) Add(b AggregateStats) {
	a.Commits += b.Commits
	a.Issues += b.Issues
	a.PRs += b.PRs
	a.Reviews += b.Reviews
	a.Releases += b.Releases
	a.Additions += b.Additions
	a.Deletions += b.Deletions
	a.NetAdditions += b.NetAdditions
	a.RepoCountContributed += b.RepoCountContributed
	a.RepoCountCreated += b.RepoCountCreated
}

// LoadMonthlyFiles loads all monthly JSON files from a directory.
func LoadMonthlyFiles(dir string) ([]MonthlyOutputFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var files []MonthlyOutputFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		// Skip non-monthly files (e.g., report.json)
		if !strings.Contains(entry.Name(), "_github_") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", entry.Name(), err)
		}

		var file MonthlyOutputFile
		if err := json.Unmarshal(data, &file); err != nil {
			return nil, fmt.Errorf("parse file %s: %w", entry.Name(), err)
		}

		files = append(files, file)
	}

	// Sort by year, then month
	sort.Slice(files, func(i, j int) bool {
		if files[i].Year != files[j].Year {
			return files[i].Year < files[j].Year
		}
		return files[i].Month < files[j].Month
	})

	return files, nil
}

// BuildStatsReport builds a StatsReport from monthly files.
func BuildStatsReport(files []MonthlyOutputFile) (*StatsReport, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no monthly files provided")
	}

	// Extract metadata from first file
	first := files[0]
	last := files[len(files)-1]

	report := &StatsReport{
		Metadata: ReportMetadata{
			Username:    first.Username,
			Visibility:  first.Metadata.Visibility,
			GeneratedAt: time.Now().UTC(),
			DataRange: DateRange{
				From: first.Metadata.From,
				To:   last.Metadata.To,
			},
		},
		Years: []YearStats{},
	}

	// Group files by year
	yearMap := make(map[int][]MonthlyOutputFile)
	for _, f := range files {
		yearMap[f.Year] = append(yearMap[f.Year], f)
	}

	// Get sorted years
	var years []int
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Ints(years)

	// Build year stats
	for _, year := range years {
		yearFiles := yearMap[year]
		yearStats := buildYearStats(year, yearFiles)
		report.Years = append(report.Years, yearStats)
	}

	return report, nil
}

// buildYearStats builds YearStats from monthly files for a single year.
func buildYearStats(year int, files []MonthlyOutputFile) YearStats {
	ys := YearStats{
		Year:     year,
		Stats:    AggregateStats{},
		Quarters: []QuarterStats{},
	}

	// Group by quarter
	quarterMap := make(map[int][]MonthlyOutputFile)
	for _, f := range files {
		q := monthToQuarter(f.Month)
		quarterMap[q] = append(quarterMap[q], f)
	}

	// Get sorted quarters
	var quarters []int
	for q := range quarterMap {
		quarters = append(quarters, q)
	}
	sort.Ints(quarters)

	// Build quarter stats
	for _, q := range quarters {
		qFiles := quarterMap[q]
		qStats := buildQuarterStats(year, q, qFiles)
		ys.Quarters = append(ys.Quarters, qStats)
		ys.Stats.Add(qStats.Stats)
	}

	return ys
}

// buildQuarterStats builds QuarterStats from monthly files for a single quarter.
func buildQuarterStats(year, quarter int, files []MonthlyOutputFile) QuarterStats {
	qs := QuarterStats{
		Quarter: quarter,
		Year:    year,
		Label:   fmt.Sprintf("Q%d %d", quarter, year),
		Stats:   AggregateStats{},
		Months:  []MonthStats{},
	}

	// Sort files by month
	sort.Slice(files, func(i, j int) bool {
		return files[i].Month < files[j].Month
	})

	for _, f := range files {
		ms := MonthStats{
			Year:      f.Year,
			Month:     f.Month,
			MonthName: f.MonthName,
			Stats:     statsFromMonthlyStats(f.Stats),
		}
		qs.Months = append(qs.Months, ms)
		qs.Stats.Add(ms.Stats)
	}

	return qs
}

// statsFromMonthlyStats converts MonthlyStats to AggregateStats.
func statsFromMonthlyStats(ms MonthlyStats) AggregateStats {
	return AggregateStats{
		Commits:              ms.Commits,
		Issues:               ms.Issues,
		PRs:                  ms.PRs,
		Reviews:              ms.Reviews,
		Releases:             ms.Releases,
		Additions:            ms.Additions,
		Deletions:            ms.Deletions,
		NetAdditions:         ms.NetAdditions,
		RepoCountContributed: ms.RepoCountContributed,
		RepoCountCreated:     ms.RepoCountCreated,
	}
}

// monthToQuarter returns the quarter (1-4) for a given month (1-12).
func monthToQuarter(month int) int {
	return (month-1)/3 + 1
}

// WriteStatsReport writes a StatsReport to a JSON file.
func WriteStatsReport(path string, report *StatsReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}

// LoadStatsReport loads a StatsReport from a JSON file.
func LoadStatsReport(path string) (*StatsReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read report: %w", err)
	}

	var report StatsReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}

	return &report, nil
}

// GetLatestQuarter returns the most recent quarter in the report.
func (r *StatsReport) GetLatestQuarter() *QuarterStats {
	if len(r.Years) == 0 {
		return nil
	}
	lastYear := r.Years[len(r.Years)-1]
	if len(lastYear.Quarters) == 0 {
		return nil
	}
	return &lastYear.Quarters[len(lastYear.Quarters)-1]
}

// GetQuarter returns a specific quarter's stats.
func (r *StatsReport) GetQuarter(year, quarter int) *QuarterStats {
	for i := range r.Years {
		if r.Years[i].Year == year {
			for j := range r.Years[i].Quarters {
				if r.Years[i].Quarters[j].Quarter == quarter {
					return &r.Years[i].Quarters[j]
				}
			}
		}
	}
	return nil
}

// GetYear returns a specific year's stats.
func (r *StatsReport) GetYear(year int) *YearStats {
	for i := range r.Years {
		if r.Years[i].Year == year {
			return &r.Years[i]
		}
	}
	return nil
}

// TotalStats returns the total stats across all years.
func (r *StatsReport) TotalStats() AggregateStats {
	var total AggregateStats
	for _, y := range r.Years {
		total.Add(y.Stats)
	}
	return total
}
