package readme

// DefaultTemplate is the default Go template for README generation.
// It supports all Config options and gracefully handles missing data.
const DefaultTemplate = `{{- /* Greeting */ -}}
{{- if .Config.Greeting }}
# {{ .Config.Greeting }}
{{ end }}
{{- /* Bio */ -}}
{{- if .Config.Bio }}
{{ .Config.Bio }}
{{ end }}
{{- /* Current work and learning */ -}}
{{- if or .Config.CurrentWork .Config.Learning }}

{{- if .Config.CurrentWork }}
- {{ .Config.CurrentWork }}
{{- end }}
{{- if .Config.Learning }}
- {{ .Config.Learning }}
{{- end }}
{{ end }}
{{- /* Contribution Heatmap */ -}}
{{- if and .Config.ShowHeatmap .Heatmap }}

## Contribution Activity

` + "```" + `
{{ .Heatmap }}` + "```" + `
{{ end }}
{{- /* GitHub Stats Table */ -}}
{{- if .Config.ShowStats }}

## GitHub Stats

_{{ formatDateRange .Profile.From .Profile.To }}_

| Metric | Value |
|--------|-------|
| Commits | {{ formatNumber .Profile.TotalCommits }} |
| Pull Requests | {{ formatNumber .Profile.TotalPRs }} |
| Issues | {{ formatNumber .Profile.TotalIssues }} |
| Code Reviews | {{ formatNumber .Profile.TotalReviews }} |
| Lines Added | +{{ formatNumber .Profile.TotalAdditions }} |
| Lines Deleted | -{{ formatNumber .Profile.TotalDeletions }} |
{{- if .Profile.Calendar }}
| Longest Streak | {{ .Profile.Calendar.LongestStreak }} days |
{{- end }}
{{ end }}
{{- /* Top Repositories */ -}}
{{- if and .Config.ShowTopRepos (gt (len .TopRepos) 0) }}

## Top Repositories

| Repository | Commits | Lines Changed |
|------------|---------|---------------|
{{- range .TopRepos }}
| [{{ .FullName }}]({{ repoURL .FullName }}) | {{ formatNumber .Commits }} | {{ formatChange .Additions .Deletions }} |
{{- end }}
{{ end }}
{{- /* Organizations */ -}}
{{- if gt (len .Config.Organizations) 0 }}

## Other Projects

{{- range .Config.Organizations }}
- [{{ .Name }}]({{ .URL }}){{ if .Description }} - {{ .Description }}{{ end }}
{{- end }}
{{ end }}
{{- /* Connect Links */ -}}
{{- if hasLinks .Config }}

## Connect

{{ connectLinks .Config }}
{{ end }}
{{- /* External Stats */ -}}
{{- if gt (len .Config.ExternalStats) 0 }}

## Stats

| Platform | Metric | Value |
|----------|--------|-------|
{{- range .Config.ExternalStats }}
| {{ .Platform }} | {{ .Label }} | [{{ .Value }}]({{ .URL }}) |
{{- end }}
{{ end }}
`
