package readme

import (
	"strings"
	"testing"
	"time"

	"github.com/grokify/gogithub/profile"
)

func TestGenerateHeatmapNil(t *testing.T) {
	result := GenerateHeatmap(nil)
	if result != "" {
		t.Errorf("GenerateHeatmap(nil) = %q, want empty string", result)
	}
}

func TestGenerateHeatmapEmpty(t *testing.T) {
	cal := &profile.ContributionCalendar{}
	result := GenerateHeatmap(cal)
	if result != "" {
		t.Errorf("GenerateHeatmap(empty) = %q, want empty string", result)
	}
}

func TestGenerateHeatmapBasic(t *testing.T) {
	// Create a simple calendar with some contribution data
	cal := createTestCalendar(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 52)

	result := GenerateHeatmap(cal)

	// Should have Mon, Wed, Fri labels
	if !strings.Contains(result, "Mon") {
		t.Error("Heatmap missing 'Mon' label")
	}
	if !strings.Contains(result, "Wed") {
		t.Error("Heatmap missing 'Wed' label")
	}
	if !strings.Contains(result, "Fri") {
		t.Error("Heatmap missing 'Fri' label")
	}

	// Should contain block characters
	hasBlocks := strings.ContainsAny(result, "░▒▓█")
	if !hasBlocks {
		t.Error("Heatmap missing block characters")
	}

	// Should have multiple lines
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("Heatmap should have at least 3 lines (header + 3 weekday rows), got %d", len(lines))
	}
}

func TestGenerateCompactHeatmap(t *testing.T) {
	cal := createTestCalendar(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 10)

	result := GenerateCompactHeatmap(cal)

	// Should not be empty
	if result == "" {
		t.Error("GenerateCompactHeatmap returned empty string for valid calendar")
	}

	// Should contain block characters
	hasBlocks := strings.ContainsAny(result, "░▒▓█")
	if !hasBlocks {
		t.Error("Compact heatmap missing block characters")
	}

	// Should be a single line (no newlines)
	if strings.Contains(result, "\n") {
		t.Error("Compact heatmap should be a single line")
	}
}

func TestHeatmapChars(t *testing.T) {
	// Verify all levels have mappings
	levels := []profile.ContributionLevel{
		profile.LevelNone,
		profile.LevelLow,
		profile.LevelMedium,
		profile.LevelHigh,
		profile.LevelMaximum,
	}

	for _, level := range levels {
		char, ok := HeatmapChars[level]
		if !ok {
			t.Errorf("HeatmapChars missing mapping for level %d", level)
		}
		if char == 0 {
			t.Errorf("HeatmapChars[%d] is zero rune", level)
		}
	}
}

func TestGenerateWeekColumns(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)

	dayLevels := map[string]profile.ContributionLevel{
		"2024-01-01": profile.LevelLow,
		"2024-01-02": profile.LevelMedium,
		"2024-01-08": profile.LevelHigh,
	}

	weeks := generateWeekColumns(start, end, dayLevels)

	// Should have at least 2 weeks for a 14-day period
	if len(weeks) < 2 {
		t.Errorf("generateWeekColumns returned %d weeks, expected at least 2", len(weeks))
	}
}

func TestMonthLabels(t *testing.T) {
	if len(monthLabels) != 12 {
		t.Errorf("monthLabels has %d entries, expected 12", len(monthLabels))
	}

	expected := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for i, want := range expected {
		if monthLabels[i] != want {
			t.Errorf("monthLabels[%d] = %q, want %q", i, monthLabels[i], want)
		}
	}
}

// createTestCalendar creates a test calendar with the specified number of weeks.
func createTestCalendar(startDate time.Time, numWeeks int) *profile.ContributionCalendar {
	var days []profile.CalendarDay

	// Start from the Sunday of the start date's week
	weekStart := startDate.AddDate(0, 0, -int(startDate.Weekday()))

	totalContributions := 0
	for w := 0; w < numWeeks; w++ {
		for d := 0; d < 7; d++ {
			date := weekStart.AddDate(0, 0, w*7+d)
			// Alternate contribution levels for variety
			count := (w + d) % 5
			totalContributions += count
			days = append(days, profile.CalendarDay{
				Date:              date,
				Weekday:           date.Weekday(),
				ContributionCount: count,
				Level:             profile.CalculateLevel(count),
			})
		}
	}

	return profile.NewCalendarFromDays(days)
}
