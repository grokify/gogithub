package svg

import (
	"fmt"
	"strings"

	"github.com/grokify/gogithub/profile"
	"github.com/grokify/mogo/strconv/strconvutil"
)

// StatsCard generates an SVG stats card from a UserProfile.
type StatsCard struct {
	*Card
	stats []StatRow
}

// NewStatsCard creates a new stats card from a UserProfile.
func NewStatsCard(p *profile.UserProfile, theme Theme, title string) *StatsCard {
	if title == "" {
		title = fmt.Sprintf("%s's GitHub Stats", p.Username)
	}

	card := NewCard(title, theme)
	sc := &StatsCard{
		Card:  card,
		stats: buildStats(p),
	}

	// Calculate height based on number of stats
	sc.SetHeight(CalculateHeight(len(sc.stats)))

	return sc
}

// buildStats creates the stat rows from a profile.
func buildStats(p *profile.UserProfile) []StatRow {
	rows := []StatRow{
		{
			Icon:  IconCommit,
			Label: "Total Commits",
			Value: strconvutil.Commify(int64(p.TotalCommits)),
		},
		{
			Icon:  IconPR,
			Label: "Pull Requests",
			Value: strconvutil.Commify(int64(p.TotalPRs)),
		},
		{
			Icon:  IconIssue,
			Label: "Issues",
			Value: strconvutil.Commify(int64(p.TotalIssues)),
		},
		{
			Icon:  IconReview,
			Label: "Code Reviews",
			Value: strconvutil.Commify(int64(p.TotalReviews)),
		},
		{
			Icon:  IconRepo,
			Label: "Repos Contributed To",
			Value: strconvutil.Commify(int64(p.ReposContributedTo)),
		},
	}

	// Add code stats if available
	if p.TotalAdditions > 0 || p.TotalDeletions > 0 {
		net := p.TotalAdditions - p.TotalDeletions
		sign := "+"
		if net < 0 {
			sign = ""
		}
		rows = append(rows, StatRow{
			Icon:  IconCode,
			Label: "Lines Changed",
			Value: fmt.Sprintf("+%s / -%s (%s%s)",
				strconvutil.Commify(int64(p.TotalAdditions)),
				strconvutil.Commify(int64(p.TotalDeletions)),
				sign,
				strconvutil.Commify(int64(net)),
			),
		})
	}

	return rows
}

// Render generates the complete SVG string.
func (sc *StatsCard) Render() string {
	var sb strings.Builder

	// XML declaration (optional but recommended for standalone SVG files)
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")

	// SVG header
	sb.WriteString(sc.RenderHeader())
	sb.WriteString("\n")

	// Title
	sb.WriteString(sc.RenderTitle())
	sb.WriteString("\n")

	// Styles
	sb.WriteString(sc.RenderStyles())
	sb.WriteString("\n")

	// Background
	sb.WriteString(sc.RenderBackground())
	sb.WriteString("\n")

	// Title text
	sb.WriteString(sc.RenderTitleText())
	sb.WriteString("\n")

	// Stats (starting below the title)
	sb.WriteString(sc.RenderStatRows(sc.stats, 55))
	sb.WriteString("\n")

	// Footer
	sb.WriteString(sc.RenderFooter())
	sb.WriteString("\n")

	return sb.String()
}

// RenderBytes returns the SVG as a byte slice.
func (sc *StatsCard) RenderBytes() []byte {
	return []byte(sc.Render())
}

// StatsCardOptions configures the stats card generation.
type StatsCardOptions struct {
	Theme        string
	Title        string
	HideBorder   bool
	HideTitle    bool
	CustomStats  []StatRow // Additional custom stats to include
	ExcludeStats []string  // Stat labels to exclude
	Width        float64   // Custom width (0 = default)
	BgColor      string    // Override background color
	TitleColor   string    // Override title color
	TextColor    string    // Override text color
	IconColor    string    // Override icon color
}

// NewStatsCardWithOptions creates a stats card with custom options.
func NewStatsCardWithOptions(p *profile.UserProfile, opts StatsCardOptions) *StatsCard {
	theme := GetTheme(opts.Theme)

	// Apply color overrides
	if opts.BgColor != "" {
		theme.BgColor = opts.BgColor
	}
	if opts.TitleColor != "" {
		theme.TitleColor = opts.TitleColor
	}
	if opts.TextColor != "" {
		theme.TextColor = opts.TextColor
	}
	if opts.IconColor != "" {
		theme.IconColor = opts.IconColor
	}

	title := opts.Title
	if title == "" {
		title = fmt.Sprintf("%s's GitHub Stats", p.Username)
	}

	card := NewCard(title, theme)

	// Apply width override
	if opts.Width > 0 {
		card.Width = opts.Width
	}

	// Build stats
	stats := buildStats(p)

	// Filter excluded stats
	if len(opts.ExcludeStats) > 0 {
		exclude := make(map[string]bool)
		for _, label := range opts.ExcludeStats {
			exclude[label] = true
		}
		filtered := make([]StatRow, 0, len(stats))
		for _, s := range stats {
			if !exclude[s.Label] {
				filtered = append(filtered, s)
			}
		}
		stats = filtered
	}

	// Add custom stats
	stats = append(stats, opts.CustomStats...)

	sc := &StatsCard{
		Card:  card,
		stats: stats,
	}

	// Calculate height
	sc.SetHeight(CalculateHeight(len(sc.stats)))

	return sc
}

// GenerateSVG is a convenience function to generate an SVG from a profile.
// Deprecated: Use ProfileStatsTableSVG for the new generic chart API.
func GenerateSVG(p *profile.UserProfile, themeName, title string) string {
	theme := GetTheme(themeName)
	sc := NewStatsCard(p, theme, title)
	return sc.Render()
}

// GenerateSVGBytes is a convenience function to generate SVG bytes from a profile.
// Deprecated: Use ProfileStatsTable(p, theme, title).RenderBytes() for the new generic chart API.
func GenerateSVGBytes(p *profile.UserProfile, themeName, title string) []byte {
	return []byte(GenerateSVG(p, themeName, title))
}

// GenerateMonthlyLinesJSON creates the JSON IR for a monthly lines bar chart.
// Deprecated: Use MonthlyLinesBarChartJSON for the new generic chart API.
func GenerateMonthlyLinesJSON(p *profile.UserProfile, themeName string) ([]byte, error) {
	return MonthlyLinesBarChartJSON(p, themeName, "")
}

// GenerateMonthlyLinesSVG creates the SVG for a monthly lines bar chart.
// Deprecated: Use MonthlyLinesBarChartSVG for the new generic chart API.
func GenerateMonthlyLinesSVG(p *profile.UserProfile, themeName, title string) string {
	return MonthlyLinesBarChartSVG(p, themeName, title)
}
