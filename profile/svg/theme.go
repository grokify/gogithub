// Package svg provides SVG stats card generation for GitHub profiles.
package svg

// Theme defines the color scheme for an SVG stats card.
type Theme struct {
	Name        string
	TitleColor  string // hex color for title text
	TextColor   string // hex color for stat labels and values
	IconColor   string // hex color for stat icons
	BgColor     string // hex color for background
	BorderColor string // hex color for border
}

// themes contains all built-in themes.
var themes = map[string]Theme{
	"default": {
		Name:        "default",
		TitleColor:  "#2f80ed",
		TextColor:   "#434d58",
		IconColor:   "#4c71f2",
		BgColor:     "#fffefe",
		BorderColor: "#e4e2e2",
	},
	"dark": {
		Name:        "dark",
		TitleColor:  "#58a6ff",
		TextColor:   "#c9d1d9",
		IconColor:   "#58a6ff",
		BgColor:     "#0d1117",
		BorderColor: "#30363d",
	},
	"radical": {
		Name:        "radical",
		TitleColor:  "#fe428e",
		TextColor:   "#a9fef7",
		IconColor:   "#f8d847",
		BgColor:     "#141321",
		BorderColor: "#2a2533",
	},
	"tokyonight": {
		Name:        "tokyonight",
		TitleColor:  "#70a5fd",
		TextColor:   "#38bdae",
		IconColor:   "#bf91f3",
		BgColor:     "#1a1b27",
		BorderColor: "#252734",
	},
	"gruvbox": {
		Name:        "gruvbox",
		TitleColor:  "#fabd2f",
		TextColor:   "#ebdbb2",
		IconColor:   "#fe8019",
		BgColor:     "#282828",
		BorderColor: "#3c3836",
	},
	"dracula": {
		Name:        "dracula",
		TitleColor:  "#ff79c6",
		TextColor:   "#f8f8f2",
		IconColor:   "#bd93f9",
		BgColor:     "#282a36",
		BorderColor: "#44475a",
	},
	"nord": {
		Name:        "nord",
		TitleColor:  "#88c0d0",
		TextColor:   "#d8dee9",
		IconColor:   "#81a1c1",
		BgColor:     "#2e3440",
		BorderColor: "#3b4252",
	},
	"catppuccin": {
		Name:        "catppuccin",
		TitleColor:  "#cba6f7",
		TextColor:   "#cdd6f4",
		IconColor:   "#f5c2e7",
		BgColor:     "#1e1e2e",
		BorderColor: "#313244",
	},
}

// GetTheme returns a theme by name. If the theme is not found,
// it returns the default theme.
func GetTheme(name string) Theme {
	if theme, ok := themes[name]; ok {
		return theme
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

// DefaultTheme returns the default theme.
func DefaultTheme() Theme {
	return themes["default"]
}
