// Package profile provides aggregated GitHub user profile statistics.
package profile

import (
	"sort"
	"time"
)

// ContributionCalendar represents the GitHub contribution calendar grid.
// This mirrors the visual contribution graph shown on GitHub user profiles.
type ContributionCalendar struct {
	TotalContributions int
	Weeks              []CalendarWeek
}

// CalendarWeek represents a single week in the contribution calendar.
type CalendarWeek struct {
	StartDate time.Time      // Sunday of this week
	Days      [7]CalendarDay // Sunday (0) through Saturday (6)
}

// CalendarDay represents a single day in the contribution calendar.
type CalendarDay struct {
	Date              time.Time
	Weekday           time.Weekday
	ContributionCount int
	Level             ContributionLevel // Intensity level for coloring (0-4)
}

// ContributionLevel represents the intensity of contributions for visual display.
type ContributionLevel int

const (
	LevelNone    ContributionLevel = 0 // No contributions
	LevelLow     ContributionLevel = 1 // 1-3 contributions
	LevelMedium  ContributionLevel = 2 // 4-6 contributions
	LevelHigh    ContributionLevel = 3 // 7-9 contributions
	LevelMaximum ContributionLevel = 4 // 10+ contributions
)

// CalculateLevel determines the contribution level based on count.
// These thresholds approximate GitHub's visual intensity levels.
func CalculateLevel(count int) ContributionLevel {
	switch {
	case count == 0:
		return LevelNone
	case count <= 3:
		return LevelLow
	case count <= 6:
		return LevelMedium
	case count <= 9:
		return LevelHigh
	default:
		return LevelMaximum
	}
}

// GetDateRange returns the first and last dates in the calendar.
func (c *ContributionCalendar) GetDateRange() (first, last time.Time) {
	if len(c.Weeks) == 0 {
		return time.Time{}, time.Time{}
	}

	firstWeek := c.Weeks[0]
	lastWeek := c.Weeks[len(c.Weeks)-1]

	// Find first non-zero day
	for _, day := range firstWeek.Days {
		if !day.Date.IsZero() {
			first = day.Date
			break
		}
	}

	// Find last non-zero day (iterate backwards)
	for i := 6; i >= 0; i-- {
		if !lastWeek.Days[i].Date.IsZero() {
			last = lastWeek.Days[i].Date
			break
		}
	}

	return first, last
}

// GetDay returns the contribution data for a specific date.
// Returns nil if the date is not in the calendar.
func (c *ContributionCalendar) GetDay(date time.Time) *CalendarDay {
	date = normalizeDate(date)

	for i := range c.Weeks {
		for j := range c.Weeks[i].Days {
			if sameDay(c.Weeks[i].Days[j].Date, date) {
				return &c.Weeks[i].Days[j]
			}
		}
	}
	return nil
}

// GetWeek returns the week containing the given date.
// Returns nil if the date is not in the calendar.
func (c *ContributionCalendar) GetWeek(date time.Time) *CalendarWeek {
	date = normalizeDate(date)

	for i := range c.Weeks {
		for _, day := range c.Weeks[i].Days {
			if sameDay(day.Date, date) {
				return &c.Weeks[i]
			}
		}
	}
	return nil
}

// TotalForWeek returns the total contributions for a specific week.
func (w *CalendarWeek) TotalForWeek() int {
	total := 0
	for _, day := range w.Days {
		total += day.ContributionCount
	}
	return total
}

// DaysWithContributions returns the number of days with at least one contribution.
func (c *ContributionCalendar) DaysWithContributions() int {
	count := 0
	for _, week := range c.Weeks {
		for _, day := range week.Days {
			if day.ContributionCount > 0 {
				count++
			}
		}
	}
	return count
}

// LongestStreak returns the longest consecutive streak of days with contributions.
func (c *ContributionCalendar) LongestStreak() int {
	// Flatten all days into a sorted slice
	var days []CalendarDay
	for _, week := range c.Weeks {
		for _, day := range week.Days {
			if !day.Date.IsZero() {
				days = append(days, day)
			}
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.Before(days[j].Date)
	})

	maxStreak, currentStreak := 0, 0
	var lastDate time.Time

	for _, day := range days {
		if day.ContributionCount > 0 {
			if !lastDate.IsZero() && day.Date.Sub(lastDate).Hours() <= 25 {
				// Consecutive day (allowing for timezone differences)
				currentStreak++
			} else {
				currentStreak = 1
			}
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
			lastDate = day.Date
		} else {
			currentStreak = 0
			lastDate = time.Time{}
		}
	}

	return maxStreak
}

// CurrentStreak returns the current ongoing streak ending today (or most recent day).
func (c *ContributionCalendar) CurrentStreak() int {
	// Flatten and sort days
	var days []CalendarDay
	for _, week := range c.Weeks {
		for _, day := range week.Days {
			if !day.Date.IsZero() {
				days = append(days, day)
			}
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.After(days[j].Date) // Reverse chronological
	})

	streak := 0
	var lastDate time.Time

	for _, day := range days {
		if day.ContributionCount > 0 {
			if lastDate.IsZero() || lastDate.Sub(day.Date).Hours() <= 25 {
				streak++
				lastDate = day.Date
			} else {
				break
			}
		} else if streak > 0 {
			// Hit a zero day after starting a streak
			break
		}
	}

	return streak
}

// normalizeDate returns the date with time set to midnight UTC.
func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// sameDay checks if two times are on the same calendar day.
func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// NewCalendarFromDays creates a ContributionCalendar from a slice of day data.
// Days should be provided in chronological order.
func NewCalendarFromDays(days []CalendarDay) *ContributionCalendar {
	if len(days) == 0 {
		return &ContributionCalendar{}
	}

	cal := &ContributionCalendar{}

	// Group days into weeks
	weekMap := make(map[string]*CalendarWeek)

	for _, day := range days {
		cal.TotalContributions += day.ContributionCount

		// Find the Sunday of this day's week
		sunday := day.Date.AddDate(0, 0, -int(day.Date.Weekday()))
		key := sunday.Format("2006-01-02")

		week, exists := weekMap[key]
		if !exists {
			week = &CalendarWeek{StartDate: normalizeDate(sunday)}
			weekMap[key] = week
		}

		dayIdx := int(day.Date.Weekday())
		day.Level = CalculateLevel(day.ContributionCount)
		week.Days[dayIdx] = day
	}

	// Convert map to sorted slice
	for _, week := range weekMap {
		cal.Weeks = append(cal.Weeks, *week)
	}

	sort.Slice(cal.Weeks, func(i, j int) bool {
		return cal.Weeks[i].StartDate.Before(cal.Weeks[j].StartDate)
	})

	return cal
}
