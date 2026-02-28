package readme

import (
	"strings"
	"testing"
	"time"

	"github.com/grokify/gogithub/profile"
)

func TestNewGenerator(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	if g == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if g.Template == nil {
		t.Fatal("NewGenerator() Template is nil")
	}
}

func TestGenerateBasic(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	p := &profile.UserProfile{
		Username:           "testuser",
		From:               time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                 time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		TotalCommits:       500,
		TotalPRs:           50,
		TotalIssues:        25,
		TotalReviews:       100,
		TotalAdditions:     50000,
		TotalDeletions:     20000,
		ReposContributedTo: 10,
		RepoStats: []profile.RepoContribution{
			{FullName: "testuser/repo1", Commits: 100, Additions: 10000, Deletions: 5000},
			{FullName: "testuser/repo2", Commits: 80, Additions: 8000, Deletions: 3000},
		},
	}

	cfg := &Config{
		Greeting:      "Hi there",
		Bio:           "Software engineer",
		ShowStats:     true,
		ShowTopRepos:  true,
		TopReposCount: 5,
	}

	readme, err := g.Generate(p, cfg)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check for expected content
	checks := []string{
		"# Hi there",
		"Software engineer",
		"## GitHub Stats",
		"| Commits | 500 |",
		"| Pull Requests | 50 |",
		"## Top Repositories",
		"testuser/repo1",
		"testuser/repo2",
	}

	for _, check := range checks {
		if !strings.Contains(readme, check) {
			t.Errorf("Generate() output missing %q", check)
		}
	}
}

func TestGenerateWithOrganizations(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	p := &profile.UserProfile{
		Username: "testuser",
	}

	cfg := &Config{
		Organizations: []Organization{
			{Name: "MyOrg", URL: "https://github.com/myorg", Description: "My organization"},
		},
	}

	readme, err := g.Generate(p, cfg)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(readme, "## Other Projects") {
		t.Error("Generate() output missing 'Other Projects' section")
	}
	if !strings.Contains(readme, "[MyOrg](https://github.com/myorg)") {
		t.Error("Generate() output missing organization link")
	}
}

func TestGenerateWithLinks(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	p := &profile.UserProfile{
		Username: "testuser",
	}

	cfg := &Config{
		Blog:     &Link{Text: "Blog", URL: "https://example.com/blog"},
		LinkedIn: &Link{Text: "LinkedIn", URL: "https://linkedin.com/in/testuser"},
	}

	readme, err := g.Generate(p, cfg)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(readme, "## Connect") {
		t.Error("Generate() output missing 'Connect' section")
	}
	if !strings.Contains(readme, "[Blog](https://example.com/blog)") {
		t.Error("Generate() output missing blog link")
	}
	if !strings.Contains(readme, "[LinkedIn](https://linkedin.com/in/testuser)") {
		t.Error("Generate() output missing LinkedIn link")
	}
}

func TestGenerateWithExternalStats(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	p := &profile.UserProfile{
		Username: "testuser",
	}

	cfg := &Config{
		ExternalStats: []ExternalStat{
			{Platform: "StackOverflow", Label: "Reputation", Value: "15.2k", URL: "https://stackoverflow.com/users/123"},
		},
	}

	readme, err := g.Generate(p, cfg)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(readme, "## Stats") {
		t.Error("Generate() output missing 'Stats' section")
	}
	if !strings.Contains(readme, "StackOverflow") {
		t.Error("Generate() output missing platform name")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.ShowStats {
		t.Error("DefaultConfig() ShowStats should be true")
	}
	if !cfg.ShowTopRepos {
		t.Error("DefaultConfig() ShowTopRepos should be true")
	}
	if !cfg.ShowHeatmap {
		t.Error("DefaultConfig() ShowHeatmap should be true")
	}
	if cfg.TopReposCount != 5 {
		t.Errorf("DefaultConfig() TopReposCount = %d, want 5", cfg.TopReposCount)
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{100, "100"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{10000, "10.0k"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
	}

	for _, tt := range tests {
		got := formatNumber(tt.input)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatChange(t *testing.T) {
	got := formatChange(1500, 500)
	want := "+1.5k / -500"
	if got != want {
		t.Errorf("formatChange(1500, 500) = %q, want %q", got, want)
	}
}

func TestRepoURL(t *testing.T) {
	got := repoURL("owner/repo")
	want := "https://github.com/owner/repo"
	if got != want {
		t.Errorf("repoURL(%q) = %q, want %q", "owner/repo", got, want)
	}
}

func TestFormatDateRange(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	got := formatDateRange(from, to)
	want := "Jan 1, 2024 to Dec 31, 2024"
	if got != want {
		t.Errorf("formatDateRange() = %q, want %q", got, want)
	}
}

func TestHasLinks(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{"no links", &Config{}, false},
		{"blog only", &Config{Blog: &Link{Text: "Blog", URL: "https://example.com"}}, true},
		{"website only", &Config{Website: &Link{Text: "Site", URL: "https://example.com"}}, true},
		{"linkedin only", &Config{LinkedIn: &Link{Text: "LI", URL: "https://linkedin.com"}}, true},
		{"twitter only", &Config{Twitter: &Link{Text: "X", URL: "https://x.com"}}, true},
		{"multiple links", &Config{
			Blog:    &Link{Text: "Blog", URL: "https://example.com"},
			Twitter: &Link{Text: "X", URL: "https://x.com"},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasLinks(tt.cfg)
			if got != tt.want {
				t.Errorf("hasLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnectLinks(t *testing.T) {
	cfg := &Config{
		Blog:    &Link{Text: "Blog", URL: "https://example.com/blog"},
		Twitter: &Link{Text: "Twitter", URL: "https://twitter.com/test"},
	}

	got := connectLinks(cfg)

	if !strings.Contains(got, "[Blog](https://example.com/blog)") {
		t.Error("connectLinks() missing blog link")
	}
	if !strings.Contains(got, "[Twitter](https://twitter.com/test)") {
		t.Error("connectLinks() missing twitter link")
	}
	if !strings.Contains(got, " | ") {
		t.Error("connectLinks() missing separator")
	}
}
