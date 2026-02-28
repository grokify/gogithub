// Package chart provides generic SVG chart generation.
package chart

import (
	"encoding/json"
	"time"
)

// ChartType identifies the type of chart.
type ChartType string

const (
	TypeTable   ChartType = "table"
	TypeBar     ChartType = "bar"
	TypeLine    ChartType = "line"
	TypeHeatmap ChartType = "heatmap"
)

// Chart is the interface all chart types implement.
type Chart interface {
	Type() ChartType
	Render() string
	RenderBytes() []byte
	ToJSON() ([]byte, error)
}

// Metadata contains common chart metadata.
type Metadata struct {
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle,omitempty"`
	Generated time.Time `json:"generated"`
	Theme     string    `json:"theme"`
}

// Axis represents a chart axis configuration.
type Axis struct {
	Label  string   `json:"label,omitempty"`
	Labels []string `json:"labels,omitempty"` // For categorical axes
	Min    *float64 `json:"min,omitempty"`    // For numeric axes
	Max    *float64 `json:"max,omitempty"`
}

// Series represents a data series for bar/line charts.
type Series struct {
	Name  string    `json:"name"`
	Data  []float64 `json:"data"`
	Color string    `json:"color,omitempty"`
}

// Dimensions specifies chart dimensions.
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DefaultDimensions returns sensible defaults.
func DefaultDimensions() Dimensions {
	return Dimensions{Width: 400, Height: 200}
}

// Theme defines colors for chart rendering.
type Theme struct {
	Name            string `json:"name"`
	TitleColor      string `json:"title_color"`
	TextColor       string `json:"text_color"`
	BackgroundColor string `json:"background_color"`
	BorderColor     string `json:"border_color"`
	GridColor       string `json:"grid_color"`
	PositiveColor   string `json:"positive_color"`
	NegativeColor   string `json:"negative_color"`
	AccentColor     string `json:"accent_color"`
}

// themes contains built-in chart themes.
var themes = map[string]Theme{
	"default": {
		Name:            "default",
		TitleColor:      "#2f80ed",
		TextColor:       "#434d58",
		BackgroundColor: "#fffefe",
		BorderColor:     "#e4e2e2",
		GridColor:       "#e4e2e2",
		PositiveColor:   "#2ea043",
		NegativeColor:   "#da3633",
		AccentColor:     "#4c71f2",
	},
	"dark": {
		Name:            "dark",
		TitleColor:      "#58a6ff",
		TextColor:       "#c9d1d9",
		BackgroundColor: "#0d1117",
		BorderColor:     "#30363d",
		GridColor:       "#30363d",
		PositiveColor:   "#3fb950",
		NegativeColor:   "#f85149",
		AccentColor:     "#58a6ff",
	},
	"radical": {
		Name:            "radical",
		TitleColor:      "#fe428e",
		TextColor:       "#a9fef7",
		BackgroundColor: "#141321",
		BorderColor:     "#2a2533",
		GridColor:       "#2a2533",
		PositiveColor:   "#a9fef7",
		NegativeColor:   "#fe428e",
		AccentColor:     "#f8d847",
	},
	"tokyonight": {
		Name:            "tokyonight",
		TitleColor:      "#70a5fd",
		TextColor:       "#38bdae",
		BackgroundColor: "#1a1b27",
		BorderColor:     "#252734",
		GridColor:       "#252734",
		PositiveColor:   "#38bdae",
		NegativeColor:   "#f7768e",
		AccentColor:     "#bf91f3",
	},
	"gruvbox": {
		Name:            "gruvbox",
		TitleColor:      "#fabd2f",
		TextColor:       "#ebdbb2",
		BackgroundColor: "#282828",
		BorderColor:     "#3c3836",
		GridColor:       "#3c3836",
		PositiveColor:   "#b8bb26",
		NegativeColor:   "#fb4934",
		AccentColor:     "#fe8019",
	},
}

// GetTheme returns a theme by name, defaulting to "default".
func GetTheme(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}

// ThemeNames returns all available theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	return names
}

// marshalChartJSON marshals chart data to JSON with indentation.
func marshalChartJSON(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
