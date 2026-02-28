package chart

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewTableChart(t *testing.T) {
	table := NewTableChart("Test Stats", "default")

	if table.Type() != TypeTable {
		t.Errorf("Type() = %v, want %v", table.Type(), TypeTable)
	}

	if table.Metadata.Title != "Test Stats" {
		t.Errorf("Title = %q, want %q", table.Metadata.Title, "Test Stats")
	}
}

func TestTableChartAddRow(t *testing.T) {
	table := NewTableChart("Test", "default").
		AddRow("commit", "Commits", "100").
		AddRow("pr", "PRs", "50")

	if len(table.Rows) != 2 {
		t.Errorf("Rows count = %d, want 2", len(table.Rows))
	}

	if table.Rows[0].Label != "Commits" {
		t.Errorf("First row label = %q, want %q", table.Rows[0].Label, "Commits")
	}
}

func TestTableChartRender(t *testing.T) {
	table := NewTableChart("GitHub Stats", "dark").
		AddRow("commit", "Commits", "1,234").
		AddRow("pr", "Pull Requests", "56")

	svg := table.Render()

	// Verify SVG structure
	if !strings.HasPrefix(svg, "<?xml") {
		t.Error("SVG should start with XML declaration")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg> tag")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("SVG missing closing </svg>")
	}

	// Verify content
	if !strings.Contains(svg, "GitHub Stats") {
		t.Error("SVG missing title")
	}
	if !strings.Contains(svg, "1,234") {
		t.Error("SVG missing value")
	}

	// Verify dark theme colors
	if !strings.Contains(svg, "#0d1117") {
		t.Error("SVG should use dark theme background")
	}
}

func TestTableChartToJSON(t *testing.T) {
	table := NewTableChart("Test", "default").
		AddRow("commit", "Commits", "100")

	jsonBytes, err := table.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Errorf("JSON unmarshal error: %v", err)
	}

	if parsed["type"] != "table" {
		t.Errorf("JSON type = %v, want table", parsed["type"])
	}
}

func TestNewBarChart(t *testing.T) {
	bar := NewBarChart("Monthly Activity", "default")

	if bar.Type() != TypeBar {
		t.Errorf("Type() = %v, want %v", bar.Type(), TypeBar)
	}
}

func TestBarChartAddSeries(t *testing.T) {
	bar := NewBarChart("Test", "default").
		SetXLabels([]string{"Jan", "Feb", "Mar"}).
		AddSeries("Net Lines", []float64{1000, -500, 2000})

	if len(bar.Series) != 1 {
		t.Errorf("Series count = %d, want 1", len(bar.Series))
	}

	if len(bar.XAxis.Labels) != 3 {
		t.Errorf("X labels count = %d, want 3", len(bar.XAxis.Labels))
	}
}

func TestBarChartRender(t *testing.T) {
	bar := NewBarChart("Monthly Lines", "default").
		SetXLabels([]string{"Jan", "Feb", "Mar"}).
		AddSeries("Net", []float64{5000, -2000, 3000})

	svg := bar.Render()

	// Verify SVG structure
	if !strings.HasPrefix(svg, "<?xml") {
		t.Error("SVG should start with XML declaration")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg> tag")
	}

	// Verify content
	if !strings.Contains(svg, "Monthly Lines") {
		t.Error("SVG missing title")
	}
	if !strings.Contains(svg, "Jan") {
		t.Error("SVG missing x-axis label")
	}

	// Verify bars
	if !strings.Contains(svg, "bar-positive") {
		t.Error("SVG should have positive bars")
	}
	if !strings.Contains(svg, "bar-negative") {
		t.Error("SVG should have negative bars")
	}
}

func TestBarChartEmptyData(t *testing.T) {
	bar := NewBarChart("Empty", "default")
	svg := bar.Render()

	if !strings.Contains(svg, "No data available") {
		t.Error("Empty chart should show 'No data available'")
	}
}

func TestBarChartToJSON(t *testing.T) {
	bar := NewBarChart("Test", "dark").
		SetXLabels([]string{"Jan"}).
		AddSeries("Data", []float64{100})

	jsonBytes, err := bar.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Errorf("JSON unmarshal error: %v", err)
	}

	if parsed["type"] != "bar" {
		t.Errorf("JSON type = %v, want bar", parsed["type"])
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"default", "default"},
		{"dark", "dark"},
		{"nonexistent", "default"},
	}

	for _, tt := range tests {
		theme := GetTheme(tt.name)
		if theme.Name != tt.expected {
			t.Errorf("GetTheme(%q).Name = %q, want %q", tt.name, theme.Name, tt.expected)
		}
	}
}

func TestFormatCompact(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}

	for _, tt := range tests {
		result := formatCompact(tt.input)
		if result != tt.expected {
			t.Errorf("formatCompact(%g) = %q, want %q", tt.input, result, tt.expected)
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
		{"it's", "it&apos;s"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewStackedBarChart(t *testing.T) {
	chart := NewStackedBarChart("Commit Types", "default")

	if chart.Type() != TypeBar {
		t.Errorf("Type() = %v, want %v", chart.Type(), TypeBar)
	}

	if chart.Metadata.Title != "Commit Types" {
		t.Errorf("Title = %q, want %q", chart.Metadata.Title, "Commit Types")
	}
}

func TestStackedBarChartAddSeries(t *testing.T) {
	chart := NewStackedBarChart("Test", "default").
		SetXLabels([]string{"Jan", "Feb", "Mar"}).
		AddSeries("feat", []float64{10, 5, 8}).
		AddSeries("fix", []float64{3, 7, 2})

	if len(chart.Series) != 2 {
		t.Errorf("Series count = %d, want 2", len(chart.Series))
	}

	if len(chart.XAxis.Labels) != 3 {
		t.Errorf("X labels count = %d, want 3", len(chart.XAxis.Labels))
	}

	// Verify automatic color assignment
	if chart.Series[0].Color == "" {
		t.Error("First series should have a color")
	}
	if chart.Series[0].Color == chart.Series[1].Color {
		t.Error("Series should have different colors")
	}
}

func TestStackedBarChartAddSeriesWithColor(t *testing.T) {
	chart := NewStackedBarChart("Test", "default").
		AddSeriesWithColor("feat", []float64{10, 5}, "#2ea043").
		AddSeriesWithColor("fix", []float64{3, 7}, "#da3633")

	if chart.Series[0].Color != "#2ea043" {
		t.Errorf("First series color = %q, want %q", chart.Series[0].Color, "#2ea043")
	}
	if chart.Series[1].Color != "#da3633" {
		t.Errorf("Second series color = %q, want %q", chart.Series[1].Color, "#da3633")
	}
}

func TestStackedBarChartRender(t *testing.T) {
	chart := NewStackedBarChart("Commit Types by Month", "default").
		SetXLabels([]string{"Jan", "Feb", "Mar"}).
		AddSeriesWithColor("Features", []float64{10, 5, 8}, "#2ea043").
		AddSeriesWithColor("Fixes", []float64{3, 7, 2}, "#da3633").
		AddSeriesWithColor("Docs", []float64{2, 1, 4}, "#d29922")

	svg := chart.Render()

	// Verify SVG structure
	if !strings.HasPrefix(svg, "<?xml") {
		t.Error("SVG should start with XML declaration")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("SVG missing <svg> tag")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("SVG missing closing </svg>")
	}

	// Verify content
	if !strings.Contains(svg, "Commit Types by Month") {
		t.Error("SVG missing title")
	}
	if !strings.Contains(svg, "Jan") {
		t.Error("SVG missing x-axis label")
	}

	// Verify series colors
	if !strings.Contains(svg, "#2ea043") {
		t.Error("SVG missing feat color")
	}
	if !strings.Contains(svg, "#da3633") {
		t.Error("SVG missing fix color")
	}

	// Verify legend
	if !strings.Contains(svg, "Features") {
		t.Error("SVG missing legend for Features")
	}
}

func TestStackedBarChartEmptyData(t *testing.T) {
	chart := NewStackedBarChart("Empty", "default")
	svg := chart.Render()

	if !strings.Contains(svg, "No data available") {
		t.Error("Empty chart should show 'No data available'")
	}
}

func TestStackedBarChartNoLegend(t *testing.T) {
	chart := NewStackedBarChart("Test", "default").
		SetXLabels([]string{"Jan"}).
		AddSeries("feat", []float64{10}).
		ShowLegend(false)

	svg := chart.Render()

	// Legend text should not appear in SVG body
	if strings.Contains(svg, "legend-text") && strings.Contains(svg, ">feat<") {
		t.Error("SVG should not have legend when ShowLegend(false)")
	}
}

func TestStackedBarChartToJSON(t *testing.T) {
	chart := NewStackedBarChart("Test", "dark").
		SetXLabels([]string{"Jan"}).
		AddSeries("feat", []float64{100})

	jsonBytes, err := chart.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Errorf("JSON unmarshal error: %v", err)
	}

	if parsed["type"] != "bar" {
		t.Errorf("JSON type = %v, want bar", parsed["type"])
	}
}

func TestCalculatePercentages(t *testing.T) {
	series := []Series{
		{Name: "A", Data: []float64{10, 20}, Color: "#aaa"},
		{Name: "B", Data: []float64{30, 80}, Color: "#bbb"},
	}

	result := CalculatePercentages(series)

	// First point: 10 / (10+30) = 25%, 30 / (10+30) = 75%
	if result[0].Data[0] != 25 {
		t.Errorf("A[0] = %g, want 25", result[0].Data[0])
	}
	if result[1].Data[0] != 75 {
		t.Errorf("B[0] = %g, want 75", result[1].Data[0])
	}

	// Second point: 20 / (20+80) = 20%, 80 / (20+80) = 80%
	if result[0].Data[1] != 20 {
		t.Errorf("A[1] = %g, want 20", result[0].Data[1])
	}
	if result[1].Data[1] != 80 {
		t.Errorf("B[1] = %g, want 80", result[1].Data[1])
	}
}

func TestCCToChangelogCategory(t *testing.T) {
	tests := []struct {
		ccType   ConventionalCommitType
		expected ChangelogCategory
	}{
		{CCFeat, CLAdded},
		{CCFix, CLFixed},
		{CCDocs, CLDocumentation},
		{CCRefactor, CLChanged},
		{CCPerf, CLPerformance},
		{CCTest, CLTests},
		{CCBuild, CLBuild},
		{CCCI, CLInfrastructure},
		{CCChore, CLInternal},
		{CCRevert, CLFixed},
		{CCSecurity, CLSecurity},
		{CCDeps, CLDependencies},
		{CCOther, CLOther},
	}

	for _, tt := range tests {
		result := CCToChangelogCategory[tt.ccType]
		if result != tt.expected {
			t.Errorf("CCToChangelogCategory[%s] = %s, want %s", tt.ccType, result, tt.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
