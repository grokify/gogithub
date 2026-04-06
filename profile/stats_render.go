package profile

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

// RenderFormat specifies the output format for rendering.
type RenderFormat string

const (
	RenderFormatMarkdown RenderFormat = "markdown"
	RenderFormatHTML     RenderFormat = "html"
	RenderFormatText     RenderFormat = "text"
)

// RenderOptions configures report rendering.
type RenderOptions struct {
	Format           RenderFormat
	Title            string
	ShowMonthDetails bool
	ShowDataSource   bool
	DataSourceURL    string
	RawDataFiles     []string
	RegenerateCmd    string
}

// DefaultRenderOptions returns default render options.
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		Format:           RenderFormatMarkdown,
		ShowMonthDetails: true,
		ShowDataSource:   true,
	}
}

// RenderToMarkdown renders a StatsReport to Markdown format.
func RenderToMarkdown(report *StatsReport, opts RenderOptions) (string, error) {
	if opts.Title == "" {
		opts.Title = fmt.Sprintf("GitHub Statistics - %s", report.Metadata.Username)
	}

	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", opts.Title))

	// Visibility info
	sb.WriteString(fmt.Sprintf("%s repository contribution statistics.\n\n", capitalize(report.Metadata.Visibility)))

	// Render each year (most recent first for readability)
	for i := len(report.Years) - 1; i >= 0; i-- {
		year := report.Years[i]
		renderYearMarkdown(&sb, year, opts)
	}

	// Data source section
	if opts.ShowDataSource {
		renderDataSourceMarkdown(&sb, report, opts)
	}

	return sb.String(), nil
}

// renderYearMarkdown renders a single year's stats to the string builder.
func renderYearMarkdown(sb *strings.Builder, year YearStats, opts RenderOptions) {
	// Render quarters (most recent first)
	for i := len(year.Quarters) - 1; i >= 0; i-- {
		q := year.Quarters[i]
		renderQuarterMarkdown(sb, q, opts)
	}
}

// renderQuarterMarkdown renders a single quarter's stats to the string builder.
func renderQuarterMarkdown(sb *strings.Builder, q QuarterStats, opts RenderOptions) {
	sb.WriteString(fmt.Sprintf("## %s Summary\n\n", q.Label))

	// Summary table
	sb.WriteString("| Metric | Total |\n")
	sb.WriteString("|--------|------:|\n")
	sb.WriteString(fmt.Sprintf("| Commits | %s |\n", formatNumber(q.Stats.Commits)))
	sb.WriteString(fmt.Sprintf("| Releases | %s |\n", formatNumber(q.Stats.Releases)))
	sb.WriteString(fmt.Sprintf("| Additions | %s |\n", formatNumber(q.Stats.Additions)))
	sb.WriteString(fmt.Sprintf("| Deletions | %s |\n", formatNumber(q.Stats.Deletions)))
	sb.WriteString(fmt.Sprintf("| Net Additions | %s |\n", formatNumber(q.Stats.NetAdditions)))

	// Calculate total unique repos (sum of monthly, note: not truly unique across months)
	totalRepos := 0
	for _, m := range q.Months {
		totalRepos += m.Stats.RepoCountContributed
	}
	sb.WriteString(fmt.Sprintf("| Repos Contributed | %s (unique monthly) |\n", formatNumber(totalRepos)))
	sb.WriteString("\n")

	// Monthly breakdown tables
	sb.WriteString("## Monthly Breakdown\n\n")

	// Commits table
	sb.WriteString("### Commits\n\n")
	sb.WriteString("| Month | Commits | Repos |\n")
	sb.WriteString("|-------|--------:|------:|\n")
	for _, m := range q.Months {
		sb.WriteString(fmt.Sprintf("| %s %d | %s | %s |\n",
			m.MonthName, m.Year,
			formatNumber(m.Stats.Commits),
			formatNumber(m.Stats.RepoCountContributed)))
	}
	sb.WriteString(fmt.Sprintf("| **%s Total** | **%s** | |\n\n", q.Label, formatNumber(q.Stats.Commits)))

	// Code changes table
	sb.WriteString("### Code Changes\n\n")
	sb.WriteString("| Month | Additions | Deletions | Net |\n")
	sb.WriteString("|-------|----------:|----------:|----:|\n")
	for _, m := range q.Months {
		sb.WriteString(fmt.Sprintf("| %s %d | %s | %s | %s |\n",
			m.MonthName, m.Year,
			formatNumber(m.Stats.Additions),
			formatNumber(m.Stats.Deletions),
			formatSignedNumber(m.Stats.NetAdditions)))
	}
	sb.WriteString(fmt.Sprintf("| **%s Total** | **%s** | **%s** | **%s** |\n\n",
		q.Label,
		formatNumber(q.Stats.Additions),
		formatNumber(q.Stats.Deletions),
		formatSignedNumber(q.Stats.NetAdditions)))

	// Releases table
	sb.WriteString("### Releases\n\n")
	sb.WriteString("| Month | Releases |\n")
	sb.WriteString("|-------|-------:|\n")
	for _, m := range q.Months {
		sb.WriteString(fmt.Sprintf("| %s %d | %s |\n",
			m.MonthName, m.Year,
			formatNumber(m.Stats.Releases)))
	}
	sb.WriteString(fmt.Sprintf("| **%s Total** | **%s** |\n\n", q.Label, formatNumber(q.Stats.Releases)))

	// Monthly details
	if opts.ShowMonthDetails {
		sb.WriteString("## Monthly Details\n\n")
		// Most recent first
		for i := len(q.Months) - 1; i >= 0; i-- {
			m := q.Months[i]
			sb.WriteString(fmt.Sprintf("### %s %d\n\n", m.MonthName, m.Year))
			sb.WriteString(fmt.Sprintf("- **Commits:** %s across %s repositories\n",
				formatNumber(m.Stats.Commits),
				formatNumber(m.Stats.RepoCountContributed)))
			sb.WriteString(fmt.Sprintf("- **Releases:** %s\n", formatNumber(m.Stats.Releases)))
			sb.WriteString(fmt.Sprintf("- **Lines changed:** %s / %s (net %s)\n\n",
				formatSignedNumber(m.Stats.Additions),
				formatSignedNumber(-m.Stats.Deletions),
				formatSignedNumber(m.Stats.NetAdditions)))
		}
	}
}

// renderDataSourceMarkdown renders the data source section.
func renderDataSourceMarkdown(sb *strings.Builder, report *StatsReport, opts RenderOptions) {
	sb.WriteString("---\n\n")
	sb.WriteString("## Data Source\n\n")

	url := opts.DataSourceURL
	if url == "" {
		url = "https://github.com/grokify/gogithub"
	}
	sb.WriteString(fmt.Sprintf("Statistics generated from GitHub API using [gogithub](%s).\n\n", url))

	sb.WriteString("**Query parameters:**\n\n")
	sb.WriteString(fmt.Sprintf("- Visibility: %s\n", report.Metadata.Visibility))
	sb.WriteString("- Include releases: yes\n\n")

	if len(opts.RawDataFiles) > 0 {
		sb.WriteString("**Raw data files:**\n\n")
		for _, f := range opts.RawDataFiles {
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", f, f))
		}
		sb.WriteString("\n")
	}

	if opts.RegenerateCmd != "" {
		sb.WriteString("**Regenerate command:**\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString(opts.RegenerateCmd)
		sb.WriteString("\n```\n")
	}
}

// RenderToFile renders a StatsReport to a file.
func RenderToFile(path string, report *StatsReport, opts RenderOptions) error {
	var content string
	var err error

	switch opts.Format {
	case RenderFormatMarkdown:
		content, err = RenderToMarkdown(report, opts)
	case RenderFormatHTML:
		content, err = RenderToHTML(report, opts)
	case RenderFormatText:
		content, err = RenderToText(report, opts)
	default:
		content, err = RenderToMarkdown(report, opts)
	}

	if err != nil {
		return fmt.Errorf("render report: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// RenderToHTML renders a StatsReport to HTML format.
func RenderToHTML(report *StatsReport, opts RenderOptions) (string, error) {
	if opts.Title == "" {
		opts.Title = fmt.Sprintf("GitHub Statistics - %s", report.Metadata.Username)
	}

	const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 900px; margin: 0 auto; padding: 20px; }
        h1 { border-bottom: 2px solid #333; padding-bottom: 10px; }
        h2 { color: #333; margin-top: 30px; }
        h3 { color: #555; }
        table { border-collapse: collapse; width: 100%; margin: 15px 0; }
        th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
        th { background-color: #f5f5f5; }
        td:not(:first-child) { text-align: right; }
        .summary { background-color: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .month-detail { margin: 15px 0; padding: 10px; background-color: #fafafa; border-left: 3px solid #007acc; }
        .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; font-size: 0.9em; }
        code { background-color: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        pre { background-color: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Visibility | capitalize}} repository contribution statistics.</p>
    {{range .Years}}
    {{range .Quarters}}
    <h2>{{.Label}} Summary</h2>
    <div class="summary">
        <table>
            <tr><th>Metric</th><th>Total</th></tr>
            <tr><td>Commits</td><td>{{.Stats.Commits | number}}</td></tr>
            <tr><td>Releases</td><td>{{.Stats.Releases | number}}</td></tr>
            <tr><td>Additions</td><td>{{.Stats.Additions | number}}</td></tr>
            <tr><td>Deletions</td><td>{{.Stats.Deletions | number}}</td></tr>
            <tr><td>Net Additions</td><td>{{.Stats.NetAdditions | number}}</td></tr>
        </table>
    </div>

    <h3>Commits</h3>
    <table>
        <tr><th>Month</th><th>Commits</th><th>Repos</th></tr>
        {{range .Months}}
        <tr><td>{{.MonthName}} {{.Year}}</td><td>{{.Stats.Commits | number}}</td><td>{{.Stats.RepoCountContributed | number}}</td></tr>
        {{end}}
    </table>

    <h3>Code Changes</h3>
    <table>
        <tr><th>Month</th><th>Additions</th><th>Deletions</th><th>Net</th></tr>
        {{range .Months}}
        <tr><td>{{.MonthName}} {{.Year}}</td><td>{{.Stats.Additions | number}}</td><td>{{.Stats.Deletions | number}}</td><td>{{.Stats.NetAdditions | signed}}</td></tr>
        {{end}}
    </table>

    <h3>Releases</h3>
    <table>
        <tr><th>Month</th><th>Releases</th></tr>
        {{range .Months}}
        <tr><td>{{.MonthName}} {{.Year}}</td><td>{{.Stats.Releases | number}}</td></tr>
        {{end}}
    </table>
    {{end}}
    {{end}}

    <div class="footer">
        <p>Generated: {{.GeneratedAt}}</p>
        <p>Data range: {{.DataRange.From}} to {{.DataRange.To}}</p>
    </div>
</body>
</html>`

	funcMap := template.FuncMap{
		"number":     formatNumber,
		"signed":     formatSignedNumber,
		"capitalize": capitalize,
	}

	tmpl, err := template.New("html").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		Title       string
		Visibility  string
		Years       []YearStats
		GeneratedAt string
		DataRange   DateRange
	}{
		Title:       opts.Title,
		Visibility:  report.Metadata.Visibility,
		Years:       report.Years,
		GeneratedAt: report.Metadata.GeneratedAt.Format(time.RFC3339),
		DataRange:   report.Metadata.DataRange,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return sb.String(), nil
}

// RenderToText renders a StatsReport to plain text format.
func RenderToText(report *StatsReport, opts RenderOptions) (string, error) {
	if opts.Title == "" {
		opts.Title = fmt.Sprintf("GitHub Statistics - %s", report.Metadata.Username)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", opts.Title))
	sb.WriteString(strings.Repeat("=", len(opts.Title)))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%s repository contribution statistics.\n\n", capitalize(report.Metadata.Visibility)))

	for i := len(report.Years) - 1; i >= 0; i-- {
		year := report.Years[i]
		for j := len(year.Quarters) - 1; j >= 0; j-- {
			q := year.Quarters[j]
			sb.WriteString(fmt.Sprintf("%s Summary\n", q.Label))
			sb.WriteString(strings.Repeat("-", len(q.Label)+8))
			sb.WriteString("\n\n")

			sb.WriteString(fmt.Sprintf("  Commits:       %s\n", formatNumber(q.Stats.Commits)))
			sb.WriteString(fmt.Sprintf("  Releases:      %s\n", formatNumber(q.Stats.Releases)))
			sb.WriteString(fmt.Sprintf("  Additions:     %s\n", formatNumber(q.Stats.Additions)))
			sb.WriteString(fmt.Sprintf("  Deletions:     %s\n", formatNumber(q.Stats.Deletions)))
			sb.WriteString(fmt.Sprintf("  Net Additions: %s\n\n", formatNumber(q.Stats.NetAdditions)))

			sb.WriteString("Monthly Breakdown:\n\n")
			for _, m := range q.Months {
				sb.WriteString(fmt.Sprintf("  %s %d:\n", m.MonthName, m.Year))
				sb.WriteString(fmt.Sprintf("    Commits: %s across %s repos\n",
					formatNumber(m.Stats.Commits),
					formatNumber(m.Stats.RepoCountContributed)))
				sb.WriteString(fmt.Sprintf("    Releases: %s\n", formatNumber(m.Stats.Releases)))
				sb.WriteString(fmt.Sprintf("    Lines: %s / %s (net %s)\n\n",
					formatSignedNumber(m.Stats.Additions),
					formatSignedNumber(-m.Stats.Deletions),
					formatSignedNumber(m.Stats.NetAdditions)))
			}
		}
	}

	return sb.String(), nil
}

// formatNumber formats an integer with thousand separators.
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Format with commas
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// formatSignedNumber formats an integer with sign and thousand separators.
func formatSignedNumber(n int) string {
	if n >= 0 {
		return "+" + formatNumber(n)
	}
	return formatNumber(n)
}

// capitalize returns the string with the first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
