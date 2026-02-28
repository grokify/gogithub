package svg

import (
	"fmt"

	"github.com/grokify/gogithub/profile"
	"github.com/grokify/gogithub/profile/svg/chart"
	"github.com/grokify/mogo/strconv/strconvutil"
)

// ProfileStatsTable creates a table chart from a UserProfile.
func ProfileStatsTable(p *profile.UserProfile, themeName, title string) *chart.TableChart {
	if title == "" {
		title = fmt.Sprintf("%s's GitHub Stats", p.Username)
	}

	table := chart.NewTableChart(title, themeName).
		AddRow("commit", "Total Commits", strconvutil.Commify(int64(p.TotalCommits))).
		AddRow("pr", "Pull Requests", strconvutil.Commify(int64(p.TotalPRs))).
		AddRow("issue", "Issues", strconvutil.Commify(int64(p.TotalIssues))).
		AddRow("review", "Code Reviews", strconvutil.Commify(int64(p.TotalReviews))).
		AddRow("repo", "Repos Contributed To", strconvutil.Commify(int64(p.ReposContributedTo)))

	// Add code stats if available
	if p.TotalAdditions > 0 || p.TotalDeletions > 0 {
		net := p.TotalAdditions - p.TotalDeletions
		sign := "+"
		if net < 0 {
			sign = ""
		}
		table.AddRow("code", "Lines Changed", fmt.Sprintf("+%s / -%s (%s%s)",
			strconvutil.Commify(int64(p.TotalAdditions)),
			strconvutil.Commify(int64(p.TotalDeletions)),
			sign,
			strconvutil.Commify(int64(net)),
		))
	}

	return table
}

// ProfileStatsTableSVG generates an SVG stats table from a profile.
func ProfileStatsTableSVG(p *profile.UserProfile, themeName, title string) string {
	return ProfileStatsTable(p, themeName, title).Render()
}

// ProfileStatsTableJSON generates a JSON IR for the stats table.
func ProfileStatsTableJSON(p *profile.UserProfile, themeName, title string) ([]byte, error) {
	return ProfileStatsTable(p, themeName, title).ToJSON()
}

// MonthlyLinesBarChart creates a bar chart showing net lines by month.
func MonthlyLinesBarChart(p *profile.UserProfile, themeName, title string) *chart.BarChart {
	if title == "" {
		title = fmt.Sprintf("%s's Net Lines by Month", p.Username)
	}

	bar := chart.NewBarChart(title, themeName)

	if p.Activity == nil || len(p.Activity.Months) == 0 {
		return bar
	}

	var labels []string
	var data []float64

	for _, m := range p.Activity.Months {
		// Short month name
		monthName := m.Month.String()[:3]
		labels = append(labels, monthName)
		data = append(data, float64(m.Additions-m.Deletions))
	}

	bar.SetXLabels(labels).AddSeries("Net Lines", data)

	return bar
}

// MonthlyLinesBarChartSVG generates an SVG bar chart from a profile.
func MonthlyLinesBarChartSVG(p *profile.UserProfile, themeName, title string) string {
	return MonthlyLinesBarChart(p, themeName, title).Render()
}

// MonthlyLinesBarChartJSON generates a JSON IR for the monthly lines chart.
func MonthlyLinesBarChartJSON(p *profile.UserProfile, themeName, title string) ([]byte, error) {
	return MonthlyLinesBarChart(p, themeName, title).ToJSON()
}

// MonthlyAdditionsDeleteionsBarChart creates a multi-series bar chart.
func MonthlyAdditionsDeletionsBarChart(p *profile.UserProfile, themeName, title string) *chart.BarChart {
	if title == "" {
		title = fmt.Sprintf("%s's Code Changes by Month", p.Username)
	}

	bar := chart.NewBarChart(title, themeName)

	if p.Activity == nil || len(p.Activity.Months) == 0 {
		return bar
	}

	var labels []string
	var additions, deletions []float64

	for _, m := range p.Activity.Months {
		monthName := m.Month.String()[:3]
		labels = append(labels, monthName)
		additions = append(additions, float64(m.Additions))
		deletions = append(deletions, float64(-m.Deletions)) // Negative for visual effect
	}

	bar.SetXLabels(labels).
		AddSeriesWithColor("Additions", additions, "#2ea043").
		AddSeriesWithColor("Deletions", deletions, "#da3633")

	return bar
}

// CommitTypesByMonthChart creates a stacked bar chart showing conventional commit types by month.
// This is the developer-focused view showing feat, fix, docs, refactor, etc.
func CommitTypesByMonthChart(data *chart.CommitTypeData, themeName, title string) *chart.StackedBarChart {
	if title == "" {
		title = fmt.Sprintf("%s's Commit Types by Month", data.Username)
	}

	sbc := chart.NewStackedBarChart(title, themeName).
		SetDimensions(700, 300)

	if len(data.Monthly) == 0 {
		return sbc
	}

	// Build x-axis labels
	var labels []string
	for _, m := range data.Monthly {
		labels = append(labels, m.MonthName)
	}
	sbc.SetXLabels(labels)

	// Add series for each CC type (in display order)
	for _, ccType := range chart.ConventionalCommitTypes {
		var seriesData []float64
		hasData := false

		for _, m := range data.Monthly {
			count := m.ByCCType[string(ccType)]
			seriesData = append(seriesData, float64(count))
			if count > 0 {
				hasData = true
			}
		}

		// Only add series that have data
		if hasData {
			color := chart.CCTypeColors[ccType]
			label := chart.CCTypeLabels[ccType]
			sbc.AddSeriesWithColor(label, seriesData, color)
		}
	}

	return sbc
}

// CommitTypesByMonthChartSVG generates an SVG for commit types by month.
func CommitTypesByMonthChartSVG(data *chart.CommitTypeData, themeName, title string) string {
	return CommitTypesByMonthChart(data, themeName, title).Render()
}

// CommitTypesByMonthChartJSON generates a JSON IR for commit types by month.
func CommitTypesByMonthChartJSON(data *chart.CommitTypeData, themeName, title string) ([]byte, error) {
	return CommitTypesByMonthChart(data, themeName, title).ToJSON()
}

// ChangelogCategoriesByMonthChart creates a stacked bar chart showing changelog categories by month.
// This is the stakeholder-focused view showing Added, Fixed, Changed, etc.
func ChangelogCategoriesByMonthChart(data *chart.CommitTypeData, themeName, title string) *chart.StackedBarChart {
	if title == "" {
		title = fmt.Sprintf("%s's Changes by Category", data.Username)
	}

	sbc := chart.NewStackedBarChart(title, themeName).
		SetDimensions(700, 300)

	if len(data.Monthly) == 0 {
		return sbc
	}

	// Build x-axis labels
	var labels []string
	for _, m := range data.Monthly {
		labels = append(labels, m.MonthName)
	}
	sbc.SetXLabels(labels)

	// Add series for each changelog category (in display order)
	for _, cat := range chart.ChangelogCategories {
		var seriesData []float64
		hasData := false

		for _, m := range data.Monthly {
			count := m.ByCLCat[string(cat)]
			seriesData = append(seriesData, float64(count))
			if count > 0 {
				hasData = true
			}
		}

		// Only add series that have data
		if hasData {
			color := chart.CLCategoryColors[cat]
			sbc.AddSeriesWithColor(string(cat), seriesData, color)
		}
	}

	return sbc
}

// ChangelogCategoriesByMonthChartSVG generates an SVG for changelog categories by month.
func ChangelogCategoriesByMonthChartSVG(data *chart.CommitTypeData, themeName, title string) string {
	return ChangelogCategoriesByMonthChart(data, themeName, title).Render()
}

// ChangelogCategoriesByMonthChartJSON generates a JSON IR for changelog categories by month.
func ChangelogCategoriesByMonthChartJSON(data *chart.CommitTypeData, themeName, title string) ([]byte, error) {
	return ChangelogCategoriesByMonthChart(data, themeName, title).ToJSON()
}
