package chart

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// StackedBarChart renders data as a stacked vertical bar chart.
type StackedBarChart struct {
	ChartType  ChartType  `json:"type"`
	Metadata   Metadata   `json:"metadata"`
	Dimensions Dimensions `json:"dimensions"`
	XAxis      Axis       `json:"x_axis"`
	YAxis      Axis       `json:"y_axis"`
	Series     []Series   `json:"series"`
	Legend     bool       `json:"legend"`
	theme      Theme
}

// Stacked bar layout constants
const (
	StackedBarLegendHeight = 30
)

// DefaultSeriesColors provides distinct colors for series.
var DefaultSeriesColors = []string{
	"#2ea043", // green - feat/Added
	"#da3633", // red - fix/Fixed
	"#58a6ff", // blue - refactor/Changed
	"#8b949e", // gray - chore/Internal
	"#d29922", // yellow - docs/Documentation
	"#a371f7", // purple - test/Tests
	"#f78166", // orange - build/Build
	"#3fb950", // light green - perf/Performance
	"#db61a2", // pink - ci/Infrastructure
	"#79c0ff", // light blue - deps/Dependencies
}

// NewStackedBarChart creates a new stacked bar chart.
func NewStackedBarChart(title string, themeName string) *StackedBarChart {
	return &StackedBarChart{
		ChartType: TypeBar,
		Metadata: Metadata{
			Title:     title,
			Generated: time.Now().UTC(),
			Theme:     themeName,
		},
		Dimensions: Dimensions{
			Width:  600,
			Height: 250,
		},
		Series: []Series{},
		Legend: true,
		theme:  GetTheme(themeName),
	}
}

// SetXLabels sets the X-axis labels.
func (s *StackedBarChart) SetXLabels(labels []string) *StackedBarChart {
	s.XAxis.Labels = labels
	return s
}

// AddSeries adds a data series with automatic color assignment.
func (s *StackedBarChart) AddSeries(name string, data []float64) *StackedBarChart {
	color := DefaultSeriesColors[len(s.Series)%len(DefaultSeriesColors)]
	s.Series = append(s.Series, Series{
		Name:  name,
		Data:  data,
		Color: color,
	})
	return s
}

// AddSeriesWithColor adds a data series with a specific color.
func (s *StackedBarChart) AddSeriesWithColor(name string, data []float64, color string) *StackedBarChart {
	s.Series = append(s.Series, Series{
		Name:  name,
		Data:  data,
		Color: color,
	})
	return s
}

// SetDimensions sets custom dimensions.
func (s *StackedBarChart) SetDimensions(width, height int) *StackedBarChart {
	s.Dimensions.Width = width
	s.Dimensions.Height = height
	return s
}

// ShowLegend enables or disables the legend.
func (s *StackedBarChart) ShowLegend(show bool) *StackedBarChart {
	s.Legend = show
	return s
}

// Type returns the chart type.
func (s *StackedBarChart) Type() ChartType {
	return TypeBar
}

// ToJSON returns the chart as JSON.
func (s *StackedBarChart) ToJSON() ([]byte, error) {
	return marshalChartJSON(s)
}

// Render generates the SVG string.
func (s *StackedBarChart) Render() string {
	var sb strings.Builder

	width := s.Dimensions.Width
	height := s.Dimensions.Height

	legendHeight := 0
	if s.Legend && len(s.Series) > 0 {
		legendHeight = StackedBarLegendHeight
	}

	plotWidth := float64(width) - 2*BarPaddingX
	plotHeight := float64(height) - 2*BarPaddingY - BarLabelHeight - float64(legendHeight)

	// XML declaration
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")

	// SVG header
	fmt.Fprintf(&sb, `<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		width, height, width, height)
	sb.WriteString("\n")

	// Title element
	fmt.Fprintf(&sb, `  <title>%s</title>`, escapeXML(s.Metadata.Title))
	sb.WriteString("\n")

	// Styles
	fmt.Fprintf(&sb, `  <style>
    .chart-title { font: 600 14px 'Segoe UI', Ubuntu, sans-serif; fill: %s; }
    .axis-label { font: 400 10px 'Segoe UI', Ubuntu, sans-serif; fill: %s; }
    .legend-text { font: 400 10px 'Segoe UI', Ubuntu, sans-serif; fill: %s; }
    .grid-line { stroke: %s; stroke-width: 0.5; stroke-dasharray: 2,2; }
    .axis-line { stroke: %s; stroke-width: 1; }
  </style>`,
		s.theme.TitleColor,
		s.theme.TextColor,
		s.theme.TextColor,
		s.theme.GridColor,
		s.theme.TextColor,
	)
	sb.WriteString("\n")

	// Background
	fmt.Fprintf(&sb, `  <rect x="0" y="0" width="%d" height="%d" fill="%s" rx="%d"/>`,
		width, height, s.theme.BackgroundColor, BarBorderRadius)
	sb.WriteString("\n")

	// Border
	fmt.Fprintf(&sb, `  <rect x="0.5" y="0.5" width="%d" height="%d" fill="none" stroke="%s" rx="%d"/>`,
		width-1, height-1, s.theme.BorderColor, BarBorderRadius)
	sb.WriteString("\n")

	// Title text
	fmt.Fprintf(&sb, `  <text class="chart-title" x="%d" y="20" text-anchor="middle">%s</text>`,
		width/2, escapeXML(s.Metadata.Title))
	sb.WriteString("\n")

	// Calculate stacked totals per x-value
	numBars := len(s.XAxis.Labels)
	if numBars == 0 && len(s.Series) > 0 && len(s.Series[0].Data) > 0 {
		numBars = len(s.Series[0].Data)
	}

	if numBars == 0 || len(s.Series) == 0 {
		fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%d" text-anchor="middle">No data available</text>`,
			width/2, height/2)
		sb.WriteString("\n")
		sb.WriteString(`</svg>`)
		sb.WriteString("\n")
		return sb.String()
	}

	// Calculate max stacked height
	maxTotal := 0.0
	for i := 0; i < numBars; i++ {
		total := 0.0
		for _, series := range s.Series {
			if i < len(series.Data) && series.Data[i] > 0 {
				total += series.Data[i]
			}
		}
		if total > maxTotal {
			maxTotal = total
		}
	}

	if maxTotal == 0 {
		maxTotal = 1
	}

	scale := plotHeight / maxTotal
	zeroY := float64(BarPaddingY) + plotHeight

	// Draw grid lines
	gridLines := 4
	for i := 0; i <= gridLines; i++ {
		y := float64(BarPaddingY) + float64(i)*(plotHeight/float64(gridLines))
		fmt.Fprintf(&sb, `  <line class="grid-line" x1="%d" y1="%g" x2="%d" y2="%g"/>`,
			BarPaddingX, y, width-BarPaddingX, y)
		sb.WriteString("\n")
	}

	// Draw axis line
	fmt.Fprintf(&sb, `  <line class="axis-line" x1="%d" y1="%g" x2="%d" y2="%g"/>`,
		BarPaddingX, zeroY, width-BarPaddingX, zeroY)
	sb.WriteString("\n")

	// Calculate bar width
	barWidth := (plotWidth - float64(numBars-1)*BarGap) / float64(numBars)
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 60 {
		barWidth = 60
	}

	// Draw stacked bars
	for i := 0; i < numBars; i++ {
		x := float64(BarPaddingX) + float64(i)*(barWidth+BarGap)
		currentY := zeroY

		// Stack series from bottom to top
		for _, series := range s.Series {
			if i >= len(series.Data) || series.Data[i] <= 0 {
				continue
			}

			barHeight := series.Data[i] * scale
			barY := currentY - barHeight

			fmt.Fprintf(&sb, `  <rect fill="%s" x="%g" y="%g" width="%g" height="%g"/>`,
				series.Color, x, barY, barWidth, barHeight)
			sb.WriteString("\n")

			currentY = barY
		}

		// Draw x-axis label
		if i < len(s.XAxis.Labels) {
			labelX := x + barWidth/2
			labelY := zeroY + 15
			fmt.Fprintf(&sb, `  <text class="axis-label" x="%g" y="%g" text-anchor="middle">%s</text>`,
				labelX, labelY, escapeXML(s.XAxis.Labels[i]))
			sb.WriteString("\n")
		}
	}

	// Y-axis labels
	fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%g" text-anchor="end">%s</text>`,
		BarPaddingX-5, float64(BarPaddingY)+10, formatCompact(maxTotal))
	sb.WriteString("\n")

	fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%g" text-anchor="end">0</text>`,
		BarPaddingX-5, zeroY+4)
	sb.WriteString("\n")

	// Draw legend
	if s.Legend && len(s.Series) > 0 {
		legendY := float64(height) - float64(legendHeight) + 15
		legendX := float64(BarPaddingX)
		itemWidth := (plotWidth) / float64(len(s.Series))
		if itemWidth > 100 {
			itemWidth = 100
		}

		for i, series := range s.Series {
			x := legendX + float64(i)*itemWidth

			// Color box
			fmt.Fprintf(&sb, `  <rect fill="%s" x="%g" y="%g" width="10" height="10" rx="2"/>`,
				series.Color, x, legendY-8)
			sb.WriteString("\n")

			// Label
			fmt.Fprintf(&sb, `  <text class="legend-text" x="%g" y="%g">%s</text>`,
				x+14, legendY, escapeXML(truncateString(series.Name, 10)))
			sb.WriteString("\n")
		}
	}

	// Footer
	sb.WriteString(`</svg>`)
	sb.WriteString("\n")

	return sb.String()
}

// RenderBytes returns the SVG as bytes.
func (s *StackedBarChart) RenderBytes() []byte {
	return []byte(s.Render())
}

// truncateString truncates a string to maxLen characters with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

// CalculatePercentages converts absolute values to percentages for 100% stacked charts.
func CalculatePercentages(series []Series) []Series {
	if len(series) == 0 {
		return series
	}

	numPoints := len(series[0].Data)
	result := make([]Series, len(series))

	for i := range series {
		result[i] = Series{
			Name:  series[i].Name,
			Color: series[i].Color,
			Data:  make([]float64, numPoints),
		}
	}

	for j := 0; j < numPoints; j++ {
		total := 0.0
		for i := range series {
			if j < len(series[i].Data) {
				total += math.Abs(series[i].Data[j])
			}
		}

		if total > 0 {
			for i := range series {
				if j < len(series[i].Data) {
					result[i].Data[j] = (math.Abs(series[i].Data[j]) / total) * 100
				}
			}
		}
	}

	return result
}
