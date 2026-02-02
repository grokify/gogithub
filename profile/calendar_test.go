package profile

import (
	"testing"
	"time"
)

func TestCalculateLevel(t *testing.T) {
	tests := []struct {
		count    int
		expected ContributionLevel
	}{
		{0, LevelNone},
		{1, LevelLow},
		{2, LevelLow},
		{3, LevelLow},
		{4, LevelMedium},
		{5, LevelMedium},
		{6, LevelMedium},
		{7, LevelHigh},
		{8, LevelHigh},
		{9, LevelHigh},
		{10, LevelMaximum},
		{100, LevelMaximum},
	}

	for _, tt := range tests {
		got := CalculateLevel(tt.count)
		if got != tt.expected {
			t.Errorf("CalculateLevel(%d) = %d, want %d", tt.count, got, tt.expected)
		}
	}
}

func TestNewCalendarFromDays(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Weekday: time.Monday, ContributionCount: 5},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Weekday: time.Tuesday, ContributionCount: 3},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Weekday: time.Wednesday, ContributionCount: 0},
		{Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Weekday: time.Monday, ContributionCount: 10},
	}

	cal := NewCalendarFromDays(days)

	if cal.TotalContributions != 18 {
		t.Errorf("TotalContributions = %d, want 18", cal.TotalContributions)
	}

	// Should have 2 weeks (Dec 31 - Jan 6, Jan 7 - Jan 13)
	if len(cal.Weeks) != 2 {
		t.Errorf("len(Weeks) = %d, want 2", len(cal.Weeks))
	}
}

func TestNewCalendarFromDaysEmpty(t *testing.T) {
	cal := NewCalendarFromDays(nil)

	if cal.TotalContributions != 0 {
		t.Errorf("TotalContributions = %d, want 0", cal.TotalContributions)
	}

	if len(cal.Weeks) != 0 {
		t.Errorf("len(Weeks) = %d, want 0", len(cal.Weeks))
	}
}

func TestCalendarGetDay(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Weekday: time.Monday, ContributionCount: 5},
		{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Weekday: time.Tuesday, ContributionCount: 3},
	}

	cal := NewCalendarFromDays(days)

	// Test finding existing day
	day := cal.GetDay(time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)) // Different time, same day
	if day == nil {
		t.Fatal("GetDay returned nil for existing day")
	}
	if day.ContributionCount != 5 {
		t.Errorf("ContributionCount = %d, want 5", day.ContributionCount)
	}

	// Test non-existing day
	day = cal.GetDay(time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC))
	if day != nil {
		t.Error("GetDay should return nil for non-existing day")
	}
}

func TestCalendarGetWeek(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Weekday: time.Monday, ContributionCount: 5},
		{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Weekday: time.Tuesday, ContributionCount: 3},
	}

	cal := NewCalendarFromDays(days)

	week := cal.GetWeek(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if week == nil {
		t.Fatal("GetWeek returned nil for existing week")
	}

	total := week.TotalForWeek()
	if total != 8 {
		t.Errorf("TotalForWeek = %d, want 8", total)
	}
}

func TestCalendarGetDateRange(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), Weekday: time.Friday, ContributionCount: 1},
		{Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Weekday: time.Wednesday, ContributionCount: 2},
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Weekday: time.Monday, ContributionCount: 3},
	}

	cal := NewCalendarFromDays(days)

	first, last := cal.GetDateRange()

	if first.Day() != 5 || first.Month() != time.January {
		t.Errorf("first = %v, want Jan 5", first)
	}

	if last.Day() != 15 || last.Month() != time.January {
		t.Errorf("last = %v, want Jan 15", last)
	}
}

func TestCalendarGetDateRangeEmpty(t *testing.T) {
	cal := &ContributionCalendar{}

	first, last := cal.GetDateRange()

	if !first.IsZero() {
		t.Errorf("first should be zero, got %v", first)
	}
	if !last.IsZero() {
		t.Errorf("last should be zero, got %v", last)
	}
}

func TestCalendarDaysWithContributions(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 5},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 0},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 3},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), ContributionCount: 0},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
	}

	cal := NewCalendarFromDays(days)

	count := cal.DaysWithContributions()
	if count != 3 {
		t.Errorf("DaysWithContributions = %d, want 3", count)
	}
}

func TestCalendarLongestStreak(t *testing.T) {
	// Create a streak: Jan 1-3 (3 days), gap, Jan 5-8 (4 days)
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 2},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), ContributionCount: 0}, // Gap
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), ContributionCount: 3},
		{Date: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), ContributionCount: 2},
		{Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
	}

	cal := NewCalendarFromDays(days)

	streak := cal.LongestStreak()
	if streak != 4 {
		t.Errorf("LongestStreak = %d, want 4", streak)
	}
}

func TestCalendarLongestStreakEmpty(t *testing.T) {
	cal := &ContributionCalendar{}

	streak := cal.LongestStreak()
	if streak != 0 {
		t.Errorf("LongestStreak = %d, want 0 for empty calendar", streak)
	}
}

func TestCalendarCurrentStreak(t *testing.T) {
	// Most recent days have contributions
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 0}, // Gap
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), ContributionCount: 2},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), ContributionCount: 1}, // Most recent
	}

	cal := NewCalendarFromDays(days)

	streak := cal.CurrentStreak()
	if streak != 3 {
		t.Errorf("CurrentStreak = %d, want 3", streak)
	}
}

func TestCalendarCurrentStreakBroken(t *testing.T) {
	// Most recent day has no contributions - streak finds most recent consecutive run
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 2},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 0}, // Gap day
	}

	cal := NewCalendarFromDays(days)

	// CurrentStreak finds the most recent consecutive streak, which is Jan 1-2 (2 days)
	streak := cal.CurrentStreak()
	if streak != 2 {
		t.Errorf("CurrentStreak = %d, want 2 (most recent consecutive run)", streak)
	}
}

func TestCalendarCurrentStreakWithGapInMiddle(t *testing.T) {
	// Gap in the middle breaks the streak
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 0}, // Gap
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), ContributionCount: 1},
	}

	cal := NewCalendarFromDays(days)

	// CurrentStreak is Jan 3-4 (2 days), since there's a gap before
	streak := cal.CurrentStreak()
	if streak != 2 {
		t.Errorf("CurrentStreak = %d, want 2", streak)
	}
}

func TestCalendarWeekTotalForWeek(t *testing.T) {
	week := CalendarWeek{
		StartDate: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		Days: [7]CalendarDay{
			{ContributionCount: 1},
			{ContributionCount: 2},
			{ContributionCount: 3},
			{ContributionCount: 0},
			{ContributionCount: 5},
			{ContributionCount: 0},
			{ContributionCount: 4},
		},
	}

	total := week.TotalForWeek()
	if total != 15 {
		t.Errorf("TotalForWeek = %d, want 15", total)
	}
}

func TestNormalizeDate(t *testing.T) {
	input := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.FixedZone("EST", -5*60*60))
	normalized := normalizeDate(input)

	if normalized.Hour() != 0 || normalized.Minute() != 0 || normalized.Second() != 0 {
		t.Errorf("normalizeDate should set time to midnight, got %v", normalized)
	}

	if normalized.Location() != time.UTC {
		t.Errorf("normalizeDate should set timezone to UTC, got %v", normalized.Location())
	}

	if normalized.Year() != 2024 || normalized.Month() != 6 || normalized.Day() != 15 {
		t.Errorf("normalizeDate should preserve date, got %v", normalized)
	}
}

func TestSameDay(t *testing.T) {
	tests := []struct {
		a, b     time.Time
		expected bool
	}{
		{
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
			true,
		},
		{
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			false,
		},
		{
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			false,
		},
	}

	for _, tt := range tests {
		got := sameDay(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("sameDay(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestCalendarDayLevel(t *testing.T) {
	days := []CalendarDay{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ContributionCount: 0},
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ContributionCount: 2},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), ContributionCount: 5},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), ContributionCount: 8},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), ContributionCount: 15},
	}

	cal := NewCalendarFromDays(days)

	// Verify levels were set correctly
	expectedLevels := []ContributionLevel{LevelNone, LevelLow, LevelMedium, LevelHigh, LevelMaximum}

	for i, week := range cal.Weeks {
		for j, day := range week.Days {
			if !day.Date.IsZero() {
				// Find corresponding expected level
				for k, d := range days {
					if sameDay(d.Date, day.Date) {
						if day.Level != expectedLevels[k] {
							t.Errorf("Week %d Day %d: Level = %d, want %d", i, j, day.Level, expectedLevels[k])
						}
						break
					}
				}
			}
		}
	}
}
