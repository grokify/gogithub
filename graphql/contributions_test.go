package graphql

import (
	"testing"
	"time"
)

func TestMonthlyContributionYearMonth(t *testing.T) {
	tests := []struct {
		name string
		mc   MonthlyContribution
		want string
	}{
		{
			name: "January 2024",
			mc:   MonthlyContribution{Year: 2024, Month: time.January, Count: 10},
			want: "2024-01",
		},
		{
			name: "December 2023",
			mc:   MonthlyContribution{Year: 2023, Month: time.December, Count: 5},
			want: "2023-12",
		},
		{
			name: "June 2025",
			mc:   MonthlyContribution{Year: 2025, Month: time.June, Count: 0},
			want: "2025-06",
		},
		{
			name: "February leap year",
			mc:   MonthlyContribution{Year: 2024, Month: time.February, Count: 29},
			want: "2024-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mc.YearMonth()
			if got != tt.want {
				t.Errorf("YearMonth() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContributionStatsStruct(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	stats := ContributionStats{
		Username:                     "testuser",
		From:                         from,
		To:                           to,
		TotalCommitContributions:     100,
		TotalIssueContributions:      20,
		TotalPRContributions:         15,
		TotalPRReviewContributions:   30,
		TotalRepositoryContributions: 5,
		RestrictedContributions:      10,
		ContributionsByMonth: []MonthlyContribution{
			{Year: 2024, Month: time.January, Count: 50},
			{Year: 2024, Month: time.February, Count: 50},
		},
	}

	if stats.Username != "testuser" {
		t.Errorf("Username = %q, want %q", stats.Username, "testuser")
	}
	if stats.TotalCommitContributions != 100 {
		t.Errorf("TotalCommitContributions = %d, want %d", stats.TotalCommitContributions, 100)
	}
	if stats.TotalIssueContributions != 20 {
		t.Errorf("TotalIssueContributions = %d, want %d", stats.TotalIssueContributions, 20)
	}
	if stats.TotalPRContributions != 15 {
		t.Errorf("TotalPRContributions = %d, want %d", stats.TotalPRContributions, 15)
	}
	if stats.TotalPRReviewContributions != 30 {
		t.Errorf("TotalPRReviewContributions = %d, want %d", stats.TotalPRReviewContributions, 30)
	}
	if stats.TotalRepositoryContributions != 5 {
		t.Errorf("TotalRepositoryContributions = %d, want %d", stats.TotalRepositoryContributions, 5)
	}
	if stats.RestrictedContributions != 10 {
		t.Errorf("RestrictedContributions = %d, want %d", stats.RestrictedContributions, 10)
	}
	if len(stats.ContributionsByMonth) != 2 {
		t.Errorf("ContributionsByMonth len = %d, want %d", len(stats.ContributionsByMonth), 2)
	}
}

func TestMapToMonthlyContributions(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		wantLen  int
		wantKeys []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			wantLen:  0,
			wantKeys: []string{},
		},
		{
			name: "single month",
			input: map[string]int{
				"2024-01": 10,
			},
			wantLen:  1,
			wantKeys: []string{"2024-01"},
		},
		{
			name: "multiple months sorted",
			input: map[string]int{
				"2024-03": 30,
				"2024-01": 10,
				"2024-02": 20,
			},
			wantLen:  3,
			wantKeys: []string{"2024-01", "2024-02", "2024-03"},
		},
		{
			name: "multiple years sorted",
			input: map[string]int{
				"2025-01": 50,
				"2024-01": 10,
				"2024-12": 40,
			},
			wantLen:  3,
			wantKeys: []string{"2024-01", "2024-12", "2025-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapToMonthlyContributions(tt.input)

			if len(result) != tt.wantLen {
				t.Errorf("mapToMonthlyContributions() len = %d, want %d", len(result), tt.wantLen)
			}

			for i, wantKey := range tt.wantKeys {
				if i >= len(result) {
					break
				}
				gotKey := result[i].YearMonth()
				if gotKey != wantKey {
					t.Errorf("mapToMonthlyContributions()[%d].YearMonth() = %q, want %q", i, gotKey, wantKey)
				}
			}
		})
	}
}

func TestMapToMonthlyContributionsCount(t *testing.T) {
	input := map[string]int{
		"2024-01": 10,
		"2024-02": 20,
	}

	result := mapToMonthlyContributions(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// Results are sorted, so January should be first
	if result[0].Count != 10 {
		t.Errorf("result[0].Count = %d, want 10", result[0].Count)
	}
	if result[1].Count != 20 {
		t.Errorf("result[1].Count = %d, want 20", result[1].Count)
	}
}

func TestMapToMonthlyContributionsInvalidDate(t *testing.T) {
	// Invalid date format should be skipped
	input := map[string]int{
		"invalid": 10,
		"2024-01": 20,
	}

	result := mapToMonthlyContributions(input)

	if len(result) != 1 {
		t.Errorf("mapToMonthlyContributions() should skip invalid dates, got len = %d", len(result))
	}
	if result[0].YearMonth() != "2024-01" {
		t.Errorf("result[0].YearMonth() = %q, want %q", result[0].YearMonth(), "2024-01")
	}
}

func TestMonthlyContributionStruct(t *testing.T) {
	mc := MonthlyContribution{
		Year:  2024,
		Month: time.March,
		Count: 42,
	}

	if mc.Year != 2024 {
		t.Errorf("Year = %d, want %d", mc.Year, 2024)
	}
	if mc.Month != time.March {
		t.Errorf("Month = %v, want %v", mc.Month, time.March)
	}
	if mc.Count != 42 {
		t.Errorf("Count = %d, want %d", mc.Count, 42)
	}
}
