package graphql

import (
	"testing"
	"time"
)

func TestVisibilityConstants(t *testing.T) {
	// Verify constant values are distinct
	if VisibilityAll == VisibilityPublic {
		t.Error("VisibilityAll should not equal VisibilityPublic")
	}
	if VisibilityAll == VisibilityPrivate {
		t.Error("VisibilityAll should not equal VisibilityPrivate")
	}
	if VisibilityPublic == VisibilityPrivate {
		t.Error("VisibilityPublic should not equal VisibilityPrivate")
	}

	// Verify order (iota starts at 0)
	if VisibilityAll != 0 {
		t.Errorf("VisibilityAll = %d, want 0", VisibilityAll)
	}
	if VisibilityPublic != 1 {
		t.Errorf("VisibilityPublic = %d, want 1", VisibilityPublic)
	}
	if VisibilityPrivate != 2 {
		t.Errorf("VisibilityPrivate = %d, want 2", VisibilityPrivate)
	}
}

func TestMonthlyCommitStatsYearMonth(t *testing.T) {
	tests := []struct {
		name string
		mcs  MonthlyCommitStats
		want string
	}{
		{
			name: "January 2024",
			mcs:  MonthlyCommitStats{Year: 2024, Month: time.January, Commits: 10, Additions: 100, Deletions: 50},
			want: "2024-01",
		},
		{
			name: "December 2023",
			mcs:  MonthlyCommitStats{Year: 2023, Month: time.December, Commits: 5, Additions: 50, Deletions: 25},
			want: "2023-12",
		},
		{
			name: "June 2025",
			mcs:  MonthlyCommitStats{Year: 2025, Month: time.June},
			want: "2025-06",
		},
		{
			name: "October 2024",
			mcs:  MonthlyCommitStats{Year: 2024, Month: time.October, Commits: 30, Additions: 300, Deletions: 150},
			want: "2024-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mcs.YearMonth()
			if got != tt.want {
				t.Errorf("YearMonth() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommitStatsStruct(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	stats := CommitStats{
		Username:     "testuser",
		From:         from,
		To:           to,
		Visibility:   VisibilityPublic,
		TotalCommits: 100,
		Additions:    5000,
		Deletions:    2000,
		ByMonth: []MonthlyCommitStats{
			{Year: 2024, Month: time.January, Commits: 50, Additions: 2500, Deletions: 1000},
			{Year: 2024, Month: time.February, Commits: 50, Additions: 2500, Deletions: 1000},
		},
		ByRepo: []RepoCommitStats{
			{Owner: "owner1", Name: "repo1", IsPrivate: false, Commits: 60, Additions: 3000, Deletions: 1200},
			{Owner: "owner2", Name: "repo2", IsPrivate: false, Commits: 40, Additions: 2000, Deletions: 800},
		},
	}

	if stats.Username != "testuser" {
		t.Errorf("Username = %q, want %q", stats.Username, "testuser")
	}
	if stats.Visibility != VisibilityPublic {
		t.Errorf("Visibility = %d, want %d", stats.Visibility, VisibilityPublic)
	}
	if stats.TotalCommits != 100 {
		t.Errorf("TotalCommits = %d, want %d", stats.TotalCommits, 100)
	}
	if stats.Additions != 5000 {
		t.Errorf("Additions = %d, want %d", stats.Additions, 5000)
	}
	if stats.Deletions != 2000 {
		t.Errorf("Deletions = %d, want %d", stats.Deletions, 2000)
	}
	if len(stats.ByMonth) != 2 {
		t.Errorf("ByMonth len = %d, want %d", len(stats.ByMonth), 2)
	}
	if len(stats.ByRepo) != 2 {
		t.Errorf("ByRepo len = %d, want %d", len(stats.ByRepo), 2)
	}
}

func TestMonthlyCommitStatsStruct(t *testing.T) {
	mcs := MonthlyCommitStats{
		Year:      2024,
		Month:     time.March,
		Commits:   42,
		Additions: 1000,
		Deletions: 500,
	}

	if mcs.Year != 2024 {
		t.Errorf("Year = %d, want %d", mcs.Year, 2024)
	}
	if mcs.Month != time.March {
		t.Errorf("Month = %v, want %v", mcs.Month, time.March)
	}
	if mcs.Commits != 42 {
		t.Errorf("Commits = %d, want %d", mcs.Commits, 42)
	}
	if mcs.Additions != 1000 {
		t.Errorf("Additions = %d, want %d", mcs.Additions, 1000)
	}
	if mcs.Deletions != 500 {
		t.Errorf("Deletions = %d, want %d", mcs.Deletions, 500)
	}
}

func TestRepoCommitStatsStruct(t *testing.T) {
	rcs := RepoCommitStats{
		Owner:     "myorg",
		Name:      "myrepo",
		IsPrivate: true,
		Commits:   50,
		Additions: 2500,
		Deletions: 1000,
	}

	if rcs.Owner != "myorg" {
		t.Errorf("Owner = %q, want %q", rcs.Owner, "myorg")
	}
	if rcs.Name != "myrepo" {
		t.Errorf("Name = %q, want %q", rcs.Name, "myrepo")
	}
	if !rcs.IsPrivate {
		t.Error("IsPrivate = false, want true")
	}
	if rcs.Commits != 50 {
		t.Errorf("Commits = %d, want %d", rcs.Commits, 50)
	}
	if rcs.Additions != 2500 {
		t.Errorf("Additions = %d, want %d", rcs.Additions, 2500)
	}
	if rcs.Deletions != 1000 {
		t.Errorf("Deletions = %d, want %d", rcs.Deletions, 1000)
	}
}

func TestRepoCommitStatsPublic(t *testing.T) {
	rcs := RepoCommitStats{
		Owner:     "publicorg",
		Name:      "publicrepo",
		IsPrivate: false,
		Commits:   25,
		Additions: 1250,
		Deletions: 500,
	}

	if rcs.IsPrivate {
		t.Error("IsPrivate = true, want false")
	}
}

func TestMonthlyMapToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]*MonthlyCommitStats
		wantLen  int
		wantKeys []string
	}{
		{
			name:     "empty map",
			input:    map[string]*MonthlyCommitStats{},
			wantLen:  0,
			wantKeys: []string{},
		},
		{
			name: "single month",
			input: map[string]*MonthlyCommitStats{
				"2024-01": {Year: 2024, Month: time.January, Commits: 10, Additions: 100, Deletions: 50},
			},
			wantLen:  1,
			wantKeys: []string{"2024-01"},
		},
		{
			name: "multiple months sorted",
			input: map[string]*MonthlyCommitStats{
				"2024-03": {Year: 2024, Month: time.March, Commits: 30, Additions: 300, Deletions: 150},
				"2024-01": {Year: 2024, Month: time.January, Commits: 10, Additions: 100, Deletions: 50},
				"2024-02": {Year: 2024, Month: time.February, Commits: 20, Additions: 200, Deletions: 100},
			},
			wantLen:  3,
			wantKeys: []string{"2024-01", "2024-02", "2024-03"},
		},
		{
			name: "multiple years sorted",
			input: map[string]*MonthlyCommitStats{
				"2025-01": {Year: 2025, Month: time.January, Commits: 50, Additions: 500, Deletions: 250},
				"2024-01": {Year: 2024, Month: time.January, Commits: 10, Additions: 100, Deletions: 50},
				"2024-12": {Year: 2024, Month: time.December, Commits: 40, Additions: 400, Deletions: 200},
			},
			wantLen:  3,
			wantKeys: []string{"2024-01", "2024-12", "2025-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monthlyMapToSlice(tt.input)

			if len(result) != tt.wantLen {
				t.Errorf("monthlyMapToSlice() len = %d, want %d", len(result), tt.wantLen)
			}

			for i, wantKey := range tt.wantKeys {
				if i >= len(result) {
					break
				}
				gotKey := result[i].YearMonth()
				if gotKey != wantKey {
					t.Errorf("monthlyMapToSlice()[%d].YearMonth() = %q, want %q", i, gotKey, wantKey)
				}
			}
		})
	}
}

func TestMonthlyMapToSlicePreservesData(t *testing.T) {
	input := map[string]*MonthlyCommitStats{
		"2024-01": {Year: 2024, Month: time.January, Commits: 10, Additions: 100, Deletions: 50},
		"2024-02": {Year: 2024, Month: time.February, Commits: 20, Additions: 200, Deletions: 100},
	}

	result := monthlyMapToSlice(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// Results are sorted, so January should be first
	if result[0].Commits != 10 {
		t.Errorf("result[0].Commits = %d, want 10", result[0].Commits)
	}
	if result[0].Additions != 100 {
		t.Errorf("result[0].Additions = %d, want 100", result[0].Additions)
	}
	if result[0].Deletions != 50 {
		t.Errorf("result[0].Deletions = %d, want 50", result[0].Deletions)
	}

	if result[1].Commits != 20 {
		t.Errorf("result[1].Commits = %d, want 20", result[1].Commits)
	}
	if result[1].Additions != 200 {
		t.Errorf("result[1].Additions = %d, want 200", result[1].Additions)
	}
	if result[1].Deletions != 100 {
		t.Errorf("result[1].Deletions = %d, want 100", result[1].Deletions)
	}
}

func TestRepoInfoStruct(t *testing.T) {
	ri := repoInfo{
		Owner:     "testowner",
		Name:      "testrepo",
		IsPrivate: true,
	}

	if ri.Owner != "testowner" {
		t.Errorf("Owner = %q, want %q", ri.Owner, "testowner")
	}
	if ri.Name != "testrepo" {
		t.Errorf("Name = %q, want %q", ri.Name, "testrepo")
	}
	if !ri.IsPrivate {
		t.Error("IsPrivate = false, want true")
	}
}
