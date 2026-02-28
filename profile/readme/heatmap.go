package readme

import (
	"strings"
	"time"

	"github.com/grokify/gogithub/profile"
)

// HeatmapChars maps contribution levels to Unicode block characters.
var HeatmapChars = map[profile.ContributionLevel]rune{
	profile.LevelNone:    '░', // No contributions
	profile.LevelLow:     '▒', // Low
	profile.LevelMedium:  '▓', // Medium
	profile.LevelHigh:    '█', // High
	profile.LevelMaximum: '█', // Maximum (same as high)
}

// monthLabels are 3-character abbreviated month names.
var monthLabels = []string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// GenerateHeatmap creates an ASCII contribution calendar from calendar data.
// The output shows a grid with month labels across the top and weekday labels on the left.
func GenerateHeatmap(calendar *profile.ContributionCalendar) string {
	if calendar == nil || len(calendar.Weeks) == 0 {
		return ""
	}

	var sb strings.Builder

	// Build a map of date -> level for quick lookup
	dayLevels := make(map[string]profile.ContributionLevel)
	for _, week := range calendar.Weeks {
		for _, day := range week.Days {
			if !day.Date.IsZero() {
				key := day.Date.Format("2006-01-02")
				dayLevels[key] = day.Level
			}
		}
	}

	// Get date range
	firstDate, lastDate := calendar.GetDateRange()
	if firstDate.IsZero() || lastDate.IsZero() {
		return ""
	}

	// Generate week columns from first to last date
	weeks := generateWeekColumns(firstDate, lastDate, dayLevels)
	if len(weeks) == 0 {
		return ""
	}

	// Generate month header
	monthHeader := generateMonthHeader(weeks)
	sb.WriteString(monthHeader)
	sb.WriteString("\n")

	// Generate rows for Mon, Wed, Fri (indices 1, 3, 5 in Sunday-indexed weeks)
	weekdayRows := []struct {
		label   string
		weekday int
	}{
		{"Mon", 1},
		{"Wed", 3},
		{"Fri", 5},
	}

	for _, row := range weekdayRows {
		sb.WriteString(row.label)
		sb.WriteString("  ")
		for _, week := range weeks {
			if row.weekday < len(week) {
				sb.WriteRune(HeatmapChars[week[row.weekday]])
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// weekColumn represents contribution levels for each day of a week (Sun=0 to Sat=6).
type weekColumn [7]profile.ContributionLevel

// generateWeekColumns creates a slice of week columns for the date range.
func generateWeekColumns(start, end time.Time, dayLevels map[string]profile.ContributionLevel) []weekColumn {
	var weeks []weekColumn

	// Normalize to start of day
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	// Find the Sunday on or before start date
	weekStart := start.AddDate(0, 0, -int(start.Weekday()))

	for !weekStart.After(end) {
		var week weekColumn
		for i := 0; i < 7; i++ {
			day := weekStart.AddDate(0, 0, i)
			if !day.Before(start) && !day.After(end) {
				key := day.Format("2006-01-02")
				if level, ok := dayLevels[key]; ok {
					week[i] = level
				}
			}
		}
		weeks = append(weeks, week)
		weekStart = weekStart.AddDate(0, 0, 7)
	}

	return weeks
}

// generateMonthHeader creates the month label row for the heatmap.
// It positions month names above the first week of each month.
func generateMonthHeader(weeks []weekColumn) string {
	if len(weeks) == 0 {
		return ""
	}

	// We need to figure out which week each month starts on
	// For simplicity, use a header with evenly-spaced month labels

	// Calculate total width: "Mon  " prefix (5 chars) + 1 char per week
	headerLen := 5 + len(weeks)
	header := make([]rune, headerLen)
	for i := range header {
		header[i] = ' '
	}

	// For a simple approximation, if we have ~52 weeks (1 year),
	// place month labels every ~4.3 weeks
	if len(weeks) >= 12 {
		// Calculate month positions based on actual week count
		spacing := float64(len(weeks)) / 12.0
		for m := 0; m < 12; m++ {
			pos := 5 + int(float64(m)*spacing) // 5 for "Mon  " prefix
			label := monthLabels[m]
			// Write label if it fits
			if pos+len(label) <= headerLen {
				for i, r := range label {
					if pos+i < headerLen {
						header[pos+i] = r
					}
				}
			}
		}
	} else {
		// For shorter periods, just show "Month labels (year)" header
		return "     " + strings.Repeat(" ", len(weeks))
	}

	return string(header)
}

// GenerateCompactHeatmap creates a more compact single-row heatmap.
// This is useful for inline display in the README.
func GenerateCompactHeatmap(calendar *profile.ContributionCalendar) string {
	if calendar == nil || len(calendar.Weeks) == 0 {
		return ""
	}

	var sb strings.Builder

	for _, week := range calendar.Weeks {
		// Use the max level for the week as a summary
		maxLevel := profile.LevelNone
		for _, day := range week.Days {
			if !day.Date.IsZero() && day.Level > maxLevel {
				maxLevel = day.Level
			}
		}
		sb.WriteRune(HeatmapChars[maxLevel])
	}

	return sb.String()
}
