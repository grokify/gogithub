package svg

import (
	"strings"
	"testing"

	"github.com/grokify/gogithub/profile"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"default", "default"},
		{"dark", "dark"},
		{"radical", "radical"},
		{"tokyonight", "tokyonight"},
		{"gruvbox", "gruvbox"},
		{"nonexistent", "default"}, // Should fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.name)
			if theme.Name != tt.expected {
				t.Errorf("GetTheme(%q) = %q, want %q", tt.name, theme.Name, tt.expected)
			}
		})
	}
}

func TestThemeColors(t *testing.T) {
	theme := GetTheme("default")

	if theme.TitleColor == "" {
		t.Error("default theme TitleColor is empty")
	}
	if theme.TextColor == "" {
		t.Error("default theme TextColor is empty")
	}
	if theme.IconColor == "" {
		t.Error("default theme IconColor is empty")
	}
	if theme.BgColor == "" {
		t.Error("default theme BgColor is empty")
	}
	if theme.BorderColor == "" {
		t.Error("default theme BorderColor is empty")
	}
}

func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	if len(names) < 5 {
		t.Errorf("expected at least 5 themes, got %d", len(names))
	}

	// Verify expected themes exist
	expectedThemes := []string{"default", "dark", "radical", "tokyonight", "gruvbox"}
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, expected := range expectedThemes {
		if !nameSet[expected] {
			t.Errorf("expected theme %q not found in ThemeNames()", expected)
		}
	}
}

func TestRenderIcon(t *testing.T) {
	svg := RenderIcon(IconCommit, 10, 20, 16, "#4c71f2")

	if svg == "" {
		t.Error("RenderIcon returned empty string")
	}
	if !strings.Contains(svg, "translate(10, 20)") {
		t.Error("RenderIcon missing transform")
	}
	if !strings.Contains(svg, "#4c71f2") {
		t.Error("RenderIcon missing color")
	}
	if !strings.Contains(svg, "<path") {
		t.Error("RenderIcon missing path element")
	}
}

func TestRenderIconUnknown(t *testing.T) {
	svg := RenderIcon("unknown", 0, 0, 16, "#000")
	if svg != "" {
		t.Errorf("RenderIcon for unknown type should return empty, got %q", svg)
	}
}

func TestNewCard(t *testing.T) {
	theme := GetTheme("default")
	card := NewCard("Test Title", theme)

	if card.Width != float64(DefaultWidth) {
		t.Errorf("card width = %g, want %d", card.Width, DefaultWidth)
	}
	if card.Title != "Test Title" {
		t.Errorf("card title = %q, want %q", card.Title, "Test Title")
	}
	if card.Theme.Name != "default" {
		t.Errorf("card theme = %q, want %q", card.Theme.Name, "default")
	}
}

func TestCardRenderHeader(t *testing.T) {
	theme := GetTheme("default")
	card := NewCard("Test", theme)
	header := card.RenderHeader()

	if !strings.HasPrefix(header, "<svg") {
		t.Error("RenderHeader should start with <svg")
	}
	if !strings.Contains(header, "xmlns=") {
		t.Error("RenderHeader missing xmlns")
	}
}

func TestCardRenderStyles(t *testing.T) {
	theme := GetTheme("dark")
	card := NewCard("Test", theme)
	styles := card.RenderStyles()

	if !strings.Contains(styles, "<style>") {
		t.Error("RenderStyles missing <style> tag")
	}
	if !strings.Contains(styles, theme.TitleColor) {
		t.Error("RenderStyles missing title color")
	}
	if !strings.Contains(styles, theme.TextColor) {
		t.Error("RenderStyles missing text color")
	}
}

func TestCardRenderBackground(t *testing.T) {
	theme := GetTheme("default")
	card := NewCard("Test", theme)
	bg := card.RenderBackground()

	if !strings.Contains(bg, "<rect") {
		t.Error("RenderBackground missing <rect>")
	}
	if !strings.Contains(bg, theme.BgColor) {
		t.Error("RenderBackground missing background color")
	}
	if !strings.Contains(bg, theme.BorderColor) {
		t.Error("RenderBackground missing border color")
	}
}

func TestCalculateHeight(t *testing.T) {
	tests := []struct {
		numRows   int
		minHeight float64
	}{
		{0, 65},
		{1, 90},
		{5, 190},
		{6, 215},
	}

	for _, tt := range tests {
		height := CalculateHeight(tt.numRows)
		if height < tt.minHeight {
			t.Errorf("CalculateHeight(%d) = %g, want >= %g", tt.numRows, height, tt.minHeight)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStatsCardRender(t *testing.T) {
	p := &profile.UserProfile{
		Username:           "testuser",
		TotalCommits:       1234,
		TotalPRs:           56,
		TotalIssues:        78,
		TotalReviews:       90,
		ReposContributedTo: 42,
		TotalAdditions:     50000,
		TotalDeletions:     25000,
	}

	theme := GetTheme("default")
	sc := NewStatsCard(p, theme, "")

	svg := sc.Render()

	// Verify SVG structure
	if !strings.HasPrefix(svg, "<?xml") {
		t.Error("SVG should start with XML declaration")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg> tag")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("SVG missing closing </svg> tag")
	}

	// Verify title
	if !strings.Contains(svg, "testuser&apos;s GitHub Stats") {
		t.Error("SVG missing user title")
	}

	// Verify stats are present
	if !strings.Contains(svg, "Total Commits") {
		t.Error("SVG missing commits stat")
	}
	if !strings.Contains(svg, "1,234") {
		t.Error("SVG missing formatted commit count")
	}
}

func TestStatsCardCustomTitle(t *testing.T) {
	p := &profile.UserProfile{
		Username:     "testuser",
		TotalCommits: 100,
	}

	theme := GetTheme("default")
	sc := NewStatsCard(p, theme, "Custom Title")

	svg := sc.Render()

	if !strings.Contains(svg, "Custom Title") {
		t.Error("SVG should contain custom title")
	}
}

func TestStatsCardWithOptions(t *testing.T) {
	p := &profile.UserProfile{
		Username:           "testuser",
		TotalCommits:       100,
		TotalPRs:           50,
		TotalIssues:        25,
		ReposContributedTo: 10,
	}

	opts := StatsCardOptions{
		Theme:        "dark",
		Title:        "Custom Stats",
		ExcludeStats: []string{"Issues"},
		Width:        400,
		BgColor:      "#111111",
	}

	sc := NewStatsCardWithOptions(p, opts)
	svg := sc.Render()

	// Verify custom theme colors (dark theme with override)
	if !strings.Contains(svg, "#111111") {
		t.Error("SVG should contain custom background color")
	}

	// Verify excluded stat is not present
	if strings.Contains(svg, ">Issues<") {
		t.Error("SVG should not contain excluded 'Issues' stat")
	}

	// Verify title
	if !strings.Contains(svg, "Custom Stats") {
		t.Error("SVG should contain custom title")
	}
}

func TestGenerateSVG(t *testing.T) {
	p := &profile.UserProfile{
		Username:     "testuser",
		TotalCommits: 100,
	}

	svg := GenerateSVG(p, "radical", "")
	theme := GetTheme("radical")

	// Verify theme is applied
	if !strings.Contains(svg, theme.TitleColor) {
		t.Errorf("GenerateSVG should apply radical theme title color %s", theme.TitleColor)
	}
}

func TestGenerateSVGBytes(t *testing.T) {
	p := &profile.UserProfile{
		Username:     "testuser",
		TotalCommits: 100,
	}

	bytes := GenerateSVGBytes(p, "default", "Test")

	if len(bytes) == 0 {
		t.Error("GenerateSVGBytes returned empty slice")
	}

	svg := string(bytes)
	if !strings.HasPrefix(svg, "<?xml") {
		t.Error("GenerateSVGBytes output should start with XML declaration")
	}
}
