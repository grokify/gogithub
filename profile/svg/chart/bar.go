package chart

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// BarChart renders data as a vertical bar chart.
type BarChart struct {
	ChartType  ChartType  `json:"type"`
	Metadata   Metadata   `json:"metadata"`
	Dimensions Dimensions `json:"dimensions"`
	XAxis      Axis       `json:"x_axis"`
	YAxis      Axis       `json:"y_axis"`
	Series     []Series   `json:"series"`
	theme      Theme
}

// Bar chart layout constants
const (
	BarPaddingX     = 50
	BarPaddingY     = 30
	BarLabelHeight  = 20
	BarGap          = 4
	BarBorderRadius = 4
)

// NewBarChart creates a new bar chart.
func NewBarChart(title string, themeName string) *BarChart {
	return &BarChart{
		ChartType: TypeBar,
		Metadata: Metadata{
			Title:     title,
			Generated: time.Now().UTC(),
			Theme:     themeName,
		},
		Dimensions: Dimensions{
			Width:  500,
			Height: 200,
		},
		Series: []Series{},
		theme:  GetTheme(themeName),
	}
}

// SetXLabels sets the X-axis labels.
func (b *BarChart) SetXLabels(labels []string) *BarChart {
	b.XAxis.Labels = labels
	return b
}

// SetYLabel sets the Y-axis label.
func (b *BarChart) SetYLabel(label string) *BarChart {
	b.YAxis.Label = label
	return b
}

// AddSeries adds a data series.
func (b *BarChart) AddSeries(name string, data []float64) *BarChart {
	b.Series = append(b.Series, Series{
		Name: name,
		Data: data,
	})
	return b
}

// AddSeriesWithColor adds a data series with a custom color.
func (b *BarChart) AddSeriesWithColor(name string, data []float64, color string) *BarChart {
	b.Series = append(b.Series, Series{
		Name:  name,
		Data:  data,
		Color: color,
	})
	return b
}

// SetDimensions sets custom dimensions.
func (b *BarChart) SetDimensions(width, height int) *BarChart {
	b.Dimensions.Width = width
	b.Dimensions.Height = height
	return b
}

// Type returns the chart type.
func (b *BarChart) Type() ChartType {
	return TypeBar
}

// ToJSON returns the chart as JSON.
func (b *BarChart) ToJSON() ([]byte, error) {
	return marshalChartJSON(b)
}

// Render generates the SVG string.
func (b *BarChart) Render() string {
	var sb strings.Builder

	width := b.Dimensions.Width
	height := b.Dimensions.Height
	plotWidth := float64(width) - 2*BarPaddingX
	plotHeight := float64(height) - 2*BarPaddingY - BarLabelHeight

	// XML declaration
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")

	// SVG header
	fmt.Fprintf(&sb, `<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		width, height, width, height)
	sb.WriteString("\n")

	// Title element
	fmt.Fprintf(&sb, `  <title>%s</title>`, escapeXML(b.Metadata.Title))
	sb.WriteString("\n")

	// Styles
	fmt.Fprintf(&sb, `  <style>
    .chart-title { font: 600 14px 'Segoe UI', Ubuntu, sans-serif; fill: %s; }
    .axis-label { font: 400 10px 'Segoe UI', Ubuntu, sans-serif; fill: %s; }
    .bar-positive { fill: %s; }
    .bar-negative { fill: %s; }
    .grid-line { stroke: %s; stroke-width: 0.5; stroke-dasharray: 2,2; }
    .axis-line { stroke: %s; stroke-width: 1; }
  </style>`,
		b.theme.TitleColor,
		b.theme.TextColor,
		b.theme.PositiveColor,
		b.theme.NegativeColor,
		b.theme.GridColor,
		b.theme.TextColor,
	)
	sb.WriteString("\n")

	// Background
	fmt.Fprintf(&sb, `  <rect x="0" y="0" width="%d" height="%d" fill="%s" rx="%d"/>`,
		width, height, b.theme.BackgroundColor, BarBorderRadius)
	sb.WriteString("\n")

	// Border
	fmt.Fprintf(&sb, `  <rect x="0.5" y="0.5" width="%d" height="%d" fill="none" stroke="%s" rx="%d"/>`,
		width-1, height-1, b.theme.BorderColor, BarBorderRadius)
	sb.WriteString("\n")

	// Title text
	fmt.Fprintf(&sb, `  <text class="chart-title" x="%d" y="20" text-anchor="middle">%s</text>`,
		width/2, escapeXML(b.Metadata.Title))
	sb.WriteString("\n")

	// Get all data points and calculate range
	var allData []float64
	for _, s := range b.Series {
		allData = append(allData, s.Data...)
	}

	if len(allData) == 0 {
		fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%d" text-anchor="middle">No data available</text>`,
			width/2, height/2)
		sb.WriteString("\n")
		sb.WriteString(`</svg>`)
		sb.WriteString("\n")
		return sb.String()
	}

	// Calculate min/max
	minVal, maxVal := allData[0], allData[0]
	for _, v := range allData {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Determine scale
	hasNegative := minVal < 0
	maxAbs := math.Max(math.Abs(maxVal), math.Abs(minVal))
	if maxAbs == 0 {
		maxAbs = 1
	}

	var zeroY, scale float64
	if hasNegative {
		zeroY = float64(BarPaddingY) + plotHeight/2
		scale = (plotHeight / 2) / maxAbs
	} else {
		zeroY = float64(BarPaddingY) + plotHeight
		scale = plotHeight / maxAbs
	}

	// Draw grid lines
	gridLines := 4
	for i := 0; i <= gridLines; i++ {
		y := float64(BarPaddingY) + float64(i)*(plotHeight/float64(gridLines))
		fmt.Fprintf(&sb, `  <line class="grid-line" x1="%d" y1="%g" x2="%d" y2="%g"/>`,
			BarPaddingX, y, width-BarPaddingX, y)
		sb.WriteString("\n")
	}

	// Draw zero/axis line
	fmt.Fprintf(&sb, `  <line class="axis-line" x1="%d" y1="%g" x2="%d" y2="%g"/>`,
		BarPaddingX, zeroY, width-BarPaddingX, zeroY)
	sb.WriteString("\n")

	// Determine number of bars
	numBars := len(b.XAxis.Labels)
	if numBars == 0 && len(b.Series) > 0 {
		numBars = len(b.Series[0].Data)
	}

	// Calculate bar width
	barWidth := (plotWidth - float64(numBars-1)*BarGap) / float64(numBars)
	if barWidth < 5 {
		barWidth = 5
	}

	// Draw bars (for first series only for now)
	if len(b.Series) > 0 {
		series := b.Series[0]
		for i, val := range series.Data {
			if i >= numBars {
				break
			}

			x := float64(BarPaddingX) + float64(i)*(barWidth+BarGap)
			barHeight := math.Abs(val) * scale

			var barY float64
			var barClass string

			if val >= 0 {
				barY = zeroY - barHeight
				barClass = "bar-positive"
			} else {
				barY = zeroY
				barClass = "bar-negative"
			}

			// Custom color from series
			if series.Color != "" {
				if val >= 0 {
					fmt.Fprintf(&sb, `  <rect fill="%s" x="%g" y="%g" width="%g" height="%g" rx="2"/>`,
						series.Color, x, barY, barWidth, barHeight)
				} else {
					fmt.Fprintf(&sb, `  <rect fill="%s" x="%g" y="%g" width="%g" height="%g" rx="2"/>`,
						b.theme.NegativeColor, x, barY, barWidth, barHeight)
				}
			} else if barHeight > 0 {
				fmt.Fprintf(&sb, `  <rect class="%s" x="%g" y="%g" width="%g" height="%g" rx="2"/>`,
					barClass, x, barY, barWidth, barHeight)
			}
			sb.WriteString("\n")

			// Draw x-axis label
			if i < len(b.XAxis.Labels) {
				labelX := x + barWidth/2
				labelY := float64(height) - 8
				fmt.Fprintf(&sb, `  <text class="axis-label" x="%g" y="%g" text-anchor="middle">%s</text>`,
					labelX, labelY, escapeXML(b.XAxis.Labels[i]))
				sb.WriteString("\n")
			}
		}
	}

	// Y-axis labels
	fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%g" text-anchor="end">+%s</text>`,
		BarPaddingX-5, float64(BarPaddingY)+10, formatCompact(maxAbs))
	sb.WriteString("\n")

	if hasNegative {
		fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%g" text-anchor="end">-%s</text>`,
			BarPaddingX-5, float64(BarPaddingY)+plotHeight-5, formatCompact(maxAbs))
		sb.WriteString("\n")
	}

	fmt.Fprintf(&sb, `  <text class="axis-label" x="%d" y="%g" text-anchor="end">0</text>`,
		BarPaddingX-5, zeroY+4)
	sb.WriteString("\n")

	// Footer
	sb.WriteString(`</svg>`)
	sb.WriteString("\n")

	return sb.String()
}

// RenderBytes returns the SVG as bytes.
func (b *BarChart) RenderBytes() []byte {
	return []byte(b.Render())
}

// formatCompact formats a number in compact form.
func formatCompact(n float64) string {
	absN := math.Abs(n)
	if absN < 1000 {
		return fmt.Sprintf("%.0f", absN)
	}
	if absN < 1000000 {
		return fmt.Sprintf("%.1fk", absN/1000)
	}
	return fmt.Sprintf("%.1fM", absN/1000000)
}
