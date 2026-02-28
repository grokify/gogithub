// Package readme generates GitHub profile README files from user profile data.
package readme

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/grokify/gogithub/profile"
)

// Config holds static content and display options for README generation.
type Config struct {
	// Bio/greeting section
	Greeting    string `json:"greeting,omitempty"`     // e.g., "Hi there"
	Bio         string `json:"bio,omitempty"`          // Short bio/tagline
	CurrentWork string `json:"current_work,omitempty"` // "Currently working on..."
	Learning    string `json:"learning,omitempty"`     // "Currently learning..."

	// Links section
	Organizations []Organization `json:"organizations,omitempty"` // Other GitHub orgs to highlight
	Blog          *Link          `json:"blog,omitempty"`          // Blog URL
	Website       *Link          `json:"website,omitempty"`       // Personal website
	LinkedIn      *Link          `json:"linkedin,omitempty"`      // LinkedIn profile
	Twitter       *Link          `json:"twitter,omitempty"`       // Twitter/X profile

	// Display options
	ShowStats     bool `json:"show_stats"`      // Show contribution stats table
	ShowTopRepos  bool `json:"show_top_repos"`  // Show top repositories
	ShowLanguages bool `json:"show_languages"`  // Show language breakdown (if available)
	ShowHeatmap   bool `json:"show_heatmap"`    // Show contribution heatmap
	TopReposCount int  `json:"top_repos_count"` // Number of top repos to show (default: 5)

	// External stats placeholders (to be filled by structured-profile)
	ExternalStats []ExternalStat `json:"external_stats,omitempty"` // StackOverflow, blog posts, etc.
}

// Organization represents a GitHub organization to highlight.
type Organization struct {
	Name        string `json:"name"`        // Display name
	URL         string `json:"url"`         // GitHub URL
	Description string `json:"description"` // Brief description
}

// Link represents a hyperlink with display text and URL.
type Link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// ExternalStat represents a stat from an external platform.
type ExternalStat struct {
	Platform string `json:"platform"` // "stackoverflow", "blog", etc.
	Label    string `json:"label"`    // "Reputation", "Posts", etc.
	Value    string `json:"value"`    // "15.2k", "42", etc.
	URL      string `json:"url"`      // Link to profile/site
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ShowStats:     true,
		ShowTopRepos:  true,
		ShowHeatmap:   true,
		TopReposCount: 5,
	}
}

// Generator creates README markdown from profile data and config.
type Generator struct {
	Template *template.Template // Custom template (optional)
}

// NewGenerator creates a new README generator with the default template.
func NewGenerator() (*Generator, error) {
	tmpl, err := template.New("readme").Funcs(templateFuncs).Parse(DefaultTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse default template: %w", err)
	}
	return &Generator{Template: tmpl}, nil
}

// NewGeneratorWithTemplate creates a generator with a custom template.
func NewGeneratorWithTemplate(tmpl *template.Template) *Generator {
	return &Generator{Template: tmpl}
}

// TemplateData contains all data available to the README template.
type TemplateData struct {
	Profile  *profile.UserProfile
	Config   *Config
	Heatmap  string // Pre-generated ASCII heatmap
	TopRepos []profile.RepoContribution
}

// Generate creates README markdown from profile data and config.
func (g *Generator) Generate(p *profile.UserProfile, cfg *Config) (string, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Ensure TopReposCount has a default
	if cfg.TopReposCount <= 0 {
		cfg.TopReposCount = 5
	}

	// Prepare template data
	data := &TemplateData{
		Profile: p,
		Config:  cfg,
	}

	// Generate heatmap if enabled and calendar data exists
	if cfg.ShowHeatmap && p.Calendar != nil {
		data.Heatmap = GenerateHeatmap(p.Calendar)
	}

	// Get top repos if enabled
	if cfg.ShowTopRepos {
		data.TopRepos = p.TopReposByCommits(cfg.TopReposCount)
	}

	var buf bytes.Buffer
	if err := g.Template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// GenerateToFile writes README markdown to a file.
func (g *Generator) GenerateToFile(p *profile.UserProfile, cfg *Config, path string) error {
	content, err := g.Generate(p, cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0600)
}

// templateFuncs provides helper functions for templates.
var templateFuncs = template.FuncMap{
	"formatNumber":    formatNumber,
	"formatChange":    formatChange,
	"formatDateRange": formatDateRange,
	"repoURL":         repoURL,
	"hasLinks":        hasLinks,
	"connectLinks":    connectLinks,
}

// formatNumber formats a number with thousand separators.
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

// formatChange formats additions/deletions like "+1,234 / -567".
func formatChange(additions, deletions int) string {
	return fmt.Sprintf("+%s / -%s", formatNumber(additions), formatNumber(deletions))
}

// formatDateRange formats a date range for display.
func formatDateRange(from, to time.Time) string {
	fromStr := from.Format("Jan 2, 2006")
	toStr := to.Format("Jan 2, 2006")
	return fmt.Sprintf("%s to %s", fromStr, toStr)
}

// repoURL generates a GitHub URL for a repository.
func repoURL(fullName string) string {
	return fmt.Sprintf("https://github.com/%s", fullName)
}

// hasLinks returns true if any link is configured.
func hasLinks(cfg *Config) bool {
	return cfg.Blog != nil || cfg.Website != nil || cfg.LinkedIn != nil || cfg.Twitter != nil
}

// connectLinks generates a pipe-separated list of configured links.
func connectLinks(cfg *Config) string {
	var links []string
	if cfg.Blog != nil {
		links = append(links, fmt.Sprintf("[%s](%s)", cfg.Blog.Text, cfg.Blog.URL))
	}
	if cfg.Website != nil {
		links = append(links, fmt.Sprintf("[%s](%s)", cfg.Website.Text, cfg.Website.URL))
	}
	if cfg.LinkedIn != nil {
		links = append(links, fmt.Sprintf("[%s](%s)", cfg.LinkedIn.Text, cfg.LinkedIn.URL))
	}
	if cfg.Twitter != nil {
		links = append(links, fmt.Sprintf("[%s](%s)", cfg.Twitter.Text, cfg.Twitter.URL))
	}

	result := ""
	for i, link := range links {
		if i > 0 {
			result += " | "
		}
		result += link
	}
	return result
}
