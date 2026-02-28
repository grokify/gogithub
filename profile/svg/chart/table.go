package chart

import (
	"fmt"
	"strings"
	"time"
)

// TableChart renders data as a styled table/card with icon-label-value rows.
type TableChart struct {
	ChartType  ChartType  `json:"type"`
	Metadata   Metadata   `json:"metadata"`
	Dimensions Dimensions `json:"dimensions"`
	Rows       []TableRow `json:"rows"`
	IconPaths  []IconPath `json:"icon_paths,omitempty"` // Custom icon definitions
	theme      Theme
}

// TableRow represents a single row in the table.
type TableRow struct {
	Icon  string `json:"icon"` // Icon name or custom path
	Label string `json:"label"`
	Value string `json:"value"`
	Color string `json:"color,omitempty"` // Optional row-specific color
}

// IconPath defines a custom SVG icon.
type IconPath struct {
	Name string `json:"name"`
	Path string `json:"path"` // SVG path data
}

// Table layout constants
const (
	TablePaddingX     = 25
	TablePaddingY     = 35
	TableRowHeight    = 25
	TableIconSize     = 16
	TableBorderRadius = 4.5
)

// Built-in icon paths (16x16 viewBox)
var builtinIcons = map[string]string{
	"commit":   "M11.93 8.5a4.002 4.002 0 01-7.86 0H.75a.75.75 0 010-1.5h3.32a4.002 4.002 0 017.86 0h3.32a.75.75 0 010 1.5h-3.32zm-1.43-.75a2.5 2.5 0 10-5 0 2.5 2.5 0 005 0z",
	"pr":       "M7.177 3.073L9.573.677A.25.25 0 0110 .854v4.792a.25.25 0 01-.427.177L7.177 3.427a.25.25 0 010-.354zM3.75 2.5a.75.75 0 100 1.5.75.75 0 000-1.5zm-2.25.75a2.25 2.25 0 113 2.122v5.256a2.251 2.251 0 11-1.5 0V5.372A2.25 2.25 0 011.5 3.25zM11 2.5h-1V4h1a1 1 0 011 1v5.628a2.251 2.251 0 101.5 0V5A2.5 2.5 0 0011 2.5zm1 10.25a.75.75 0 111.5 0 .75.75 0 01-1.5 0zM3.75 12a.75.75 0 100 1.5.75.75 0 000-1.5z",
	"issue":    "M8 9.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3z M8 0a8 8 0 100 16A8 8 0 008 0zM1.5 8a6.5 6.5 0 1113 0 6.5 6.5 0 01-13 0z",
	"code":     "M4.72 3.22a.75.75 0 011.06 1.06L2.56 7.5l3.22 3.22a.75.75 0 11-1.06 1.06l-3.75-3.75a.75.75 0 010-1.06l3.75-3.75zm6.56 0a.75.75 0 10-1.06 1.06L13.44 7.5l-3.22 3.22a.75.75 0 101.06 1.06l3.75-3.75a.75.75 0 000-1.06l-3.75-3.75z",
	"repo":     "M2 2.5A2.5 2.5 0 014.5 0h8.75a.75.75 0 01.75.75v12.5a.75.75 0 01-.75.75h-2.5a.75.75 0 110-1.5h1.75v-2h-8a1 1 0 00-.714 1.7.75.75 0 01-1.072 1.05A2.495 2.495 0 012 11.5v-9zm10.5-1V9h-8c-.356 0-.694.074-1 .208V2.5a1 1 0 011-1h8zM5 12.25v3.25a.25.25 0 00.4.2l1.45-1.087a.25.25 0 01.3 0L8.6 15.7a.25.25 0 00.4-.2v-3.25a.25.25 0 00-.25-.25h-3.5a.25.25 0 00-.25.25z",
	"review":   "M8 2c1.981 0 3.671.992 4.933 2.078 1.27 1.091 2.187 2.345 2.637 3.023a1.62 1.62 0 010 1.798c-.45.678-1.367 1.932-2.637 3.023C11.67 13.008 9.981 14 8 14c-1.981 0-3.671-.992-4.933-2.078C1.797 10.831.88 9.577.43 8.899a1.62 1.62 0 010-1.798c.45-.678 1.367-1.932 2.637-3.023C4.33 2.992 6.019 2 8 2zM1.679 7.932a.12.12 0 000 .136c.411.622 1.241 1.75 2.366 2.717C5.176 11.758 6.527 12.5 8 12.5c1.473 0 2.824-.742 3.955-1.715 1.124-.967 1.954-2.096 2.366-2.717a.12.12 0 000-.136c-.412-.621-1.242-1.75-2.366-2.717C10.824 4.242 9.473 3.5 8 3.5c-1.473 0-2.824.742-3.955 1.715-1.124.967-1.954 2.096-2.366 2.717zM8 10a2 2 0 100-4 2 2 0 000 4z",
	"star":     "M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.416-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z",
	"calendar": "M4.75 0a.75.75 0 01.75.75V2h5V.75a.75.75 0 011.5 0V2h1.25c.966 0 1.75.784 1.75 1.75v10.5A1.75 1.75 0 0113.25 16H2.75A1.75 1.75 0 011 14.25V3.75C1 2.784 1.784 2 2.75 2H4V.75A.75.75 0 014.75 0zm0 3.5h-.5a.25.25 0 00-.25.25V5h8V3.75a.25.25 0 00-.25-.25H4.75zm-2.25 3v7.75c0 .138.112.25.25.25h10.5a.25.25 0 00.25-.25V6.5z",
	"streak":   "M7.998 14.5c2.832 0 5-1.98 5-4.5 0-1.463-.68-2.19-1.879-3.383l-.036-.037c-1.013-1.008-2.3-2.29-2.834-4.434-.066-.268-.32-.146-.36-.1l-.007.007c-.777.818-1.318 1.775-1.612 2.678-.297.907-.419 1.756-.439 2.343l-.005.168c.002.334.005.56-.012.745-.024.267-.081.423-.173.534-.066.077-.136.131-.214.174l-.039.02c-.12.057-.26.094-.468.108l-.096.004c-.26 0-.616-.149-.987-.367a3.96 3.96 0 01-.663-.494l-.07-.068a5.031 5.031 0 00-.707-.602c-.652-.455-1.2-.699-1.485-.699-.128 0-.192.092-.192.316.497.89 1.17 1.622 1.974 2.29a7.005 7.005 0 002.81 1.505l.057.015c.406.096.82.141 1.242.141z",
}

// NewTableChart creates a new table chart.
func NewTableChart(title string, themeName string) *TableChart {
	return &TableChart{
		ChartType: TypeTable,
		Metadata: Metadata{
			Title:     title,
			Generated: time.Now().UTC(),
			Theme:     themeName,
		},
		Dimensions: Dimensions{
			Width:  350,
			Height: 195, // Will be recalculated
		},
		Rows:  []TableRow{},
		theme: GetTheme(themeName),
	}
}

// AddRow adds a row to the table.
func (t *TableChart) AddRow(icon, label, value string) *TableChart {
	t.Rows = append(t.Rows, TableRow{
		Icon:  icon,
		Label: label,
		Value: value,
	})
	t.recalculateHeight()
	return t
}

// AddRowWithColor adds a row with a custom color.
func (t *TableChart) AddRowWithColor(icon, label, value, color string) *TableChart {
	t.Rows = append(t.Rows, TableRow{
		Icon:  icon,
		Label: label,
		Value: value,
		Color: color,
	})
	t.recalculateHeight()
	return t
}

// SetDimensions sets custom dimensions.
func (t *TableChart) SetDimensions(width, height int) *TableChart {
	t.Dimensions.Width = width
	t.Dimensions.Height = height
	return t
}

func (t *TableChart) recalculateHeight() {
	// Title area: ~45px, each row: 25px, bottom padding: 20px
	t.Dimensions.Height = 45 + len(t.Rows)*TableRowHeight + 20
}

// Type returns the chart type.
func (t *TableChart) Type() ChartType {
	return TypeTable
}

// ToJSON returns the chart as JSON.
func (t *TableChart) ToJSON() ([]byte, error) {
	return marshalChartJSON(t)
}

// Render generates the SVG string.
func (t *TableChart) Render() string {
	var sb strings.Builder

	width := t.Dimensions.Width
	height := t.Dimensions.Height

	// XML declaration
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")

	// SVG header
	fmt.Fprintf(&sb, `<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		width, height, width, height)
	sb.WriteString("\n")

	// Title
	fmt.Fprintf(&sb, `  <title>%s</title>`, escapeXML(t.Metadata.Title))
	sb.WriteString("\n")

	// Styles
	fmt.Fprintf(&sb, `  <style>
    .header { font: 600 18px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .stat { font: 600 14px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .stat-value { font: 700 14px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .icon { fill: %s; }
  </style>`,
		t.theme.TitleColor,
		t.theme.TextColor,
		t.theme.TextColor,
		t.theme.AccentColor,
	)
	sb.WriteString("\n")

	// Background
	fmt.Fprintf(&sb, `  <rect x="0.5" y="0.5" rx="%g" width="%d" height="99%%" fill="%s" stroke="%s"/>`,
		TableBorderRadius, width-1, t.theme.BackgroundColor, t.theme.BorderColor)
	sb.WriteString("\n")

	// Title text
	fmt.Fprintf(&sb, `  <g transform="translate(%d, %d)">
    <text class="header">%s</text>
  </g>`,
		TablePaddingX, TablePaddingY, escapeXML(t.Metadata.Title))
	sb.WriteString("\n")

	// Rows
	fmt.Fprintf(&sb, `  <g transform="translate(%d, 55)">`, TablePaddingX)
	sb.WriteString("\n")

	for i, row := range t.Rows {
		y := i * TableRowHeight

		// Row group
		fmt.Fprintf(&sb, `    <g transform="translate(0, %d)">`, y)
		sb.WriteString("\n")

		// Icon
		iconPath := t.getIconPath(row.Icon)
		if iconPath != "" {
			iconColor := t.theme.AccentColor
			if row.Color != "" {
				iconColor = row.Color
			}
			fmt.Fprintf(&sb, `      <svg width="%d" height="%d" viewBox="0 0 16 16" class="icon">`,
				TableIconSize, TableIconSize)
			sb.WriteString("\n")
			fmt.Fprintf(&sb, `        <path fill="%s" d="%s"/>`, iconColor, iconPath)
			sb.WriteString("\n")
			sb.WriteString(`      </svg>`)
			sb.WriteString("\n")
		}

		// Label
		fmt.Fprintf(&sb, `      <text class="stat" x="25" y="12.5">%s:</text>`, escapeXML(row.Label))
		sb.WriteString("\n")

		// Value (right-aligned)
		valueX := width - TablePaddingX*2 - 10
		fmt.Fprintf(&sb, `      <text class="stat-value" x="%d" y="12.5" text-anchor="end">%s</text>`,
			valueX, escapeXML(row.Value))
		sb.WriteString("\n")

		sb.WriteString(`    </g>`)
		sb.WriteString("\n")
	}

	sb.WriteString(`  </g>`)
	sb.WriteString("\n")

	// Footer
	sb.WriteString(`</svg>`)
	sb.WriteString("\n")

	return sb.String()
}

// RenderBytes returns the SVG as bytes.
func (t *TableChart) RenderBytes() []byte {
	return []byte(t.Render())
}

func (t *TableChart) getIconPath(name string) string {
	// Check custom icons first
	for _, icon := range t.IconPaths {
		if icon.Name == name {
			return icon.Path
		}
	}
	// Fall back to built-in
	return builtinIcons[name]
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
