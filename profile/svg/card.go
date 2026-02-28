package svg

import (
	"fmt"
	"strings"
)

const (
	// DefaultWidth is the default card width in pixels.
	DefaultWidth = 350

	// DefaultPaddingX is the horizontal padding.
	DefaultPaddingX = 25

	// DefaultPaddingY is the vertical padding for the title.
	DefaultPaddingY = 35

	// DefaultBorderRadius is the corner radius.
	DefaultBorderRadius = 4.5

	// DefaultStatRowHeight is the height of each stat row.
	DefaultStatRowHeight = 25

	// DefaultIconSize is the size of stat icons.
	DefaultIconSize = 16
)

// Card represents the base SVG card structure.
type Card struct {
	Width        float64
	Height       float64
	PaddingX     float64
	PaddingY     float64
	BorderRadius float64
	Theme        Theme
	Title        string
}

// NewCard creates a new card with default dimensions.
func NewCard(title string, theme Theme) *Card {
	return &Card{
		Width:        DefaultWidth,
		Height:       195, // Will be adjusted based on content
		PaddingX:     DefaultPaddingX,
		PaddingY:     DefaultPaddingY,
		BorderRadius: DefaultBorderRadius,
		Theme:        theme,
		Title:        title,
	}
}

// SetHeight adjusts the card height.
func (c *Card) SetHeight(height float64) {
	c.Height = height
}

// RenderHeader returns the SVG header with XML declaration and opening tag.
func (c *Card) RenderHeader() string {
	return fmt.Sprintf(
		`<svg width="%g" height="%g" viewBox="0 0 %g %g" xmlns="http://www.w3.org/2000/svg">`,
		c.Width, c.Height, c.Width, c.Height,
	)
}

// RenderTitle returns the SVG title element.
func (c *Card) RenderTitle() string {
	return fmt.Sprintf(`  <title>%s</title>`, escapeXML(c.Title))
}

// RenderStyles returns the CSS styles block.
func (c *Card) RenderStyles() string {
	return fmt.Sprintf(`  <style>
    .header { font: 600 18px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .stat { font: 600 14px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .stat-value { font: 700 14px 'Segoe UI', Ubuntu, 'Helvetica Neue', sans-serif; fill: %s; }
    .icon { fill: %s; }
  </style>`,
		c.Theme.TitleColor,
		c.Theme.TextColor,
		c.Theme.TextColor,
		c.Theme.IconColor,
	)
}

// RenderBackground returns the background rectangle.
func (c *Card) RenderBackground() string {
	return fmt.Sprintf(
		`  <rect x="0.5" y="0.5" rx="%g" width="%g" height="99%%" fill="%s" stroke="%s"/>`,
		c.BorderRadius, c.Width-1, c.Theme.BgColor, c.Theme.BorderColor,
	)
}

// RenderTitleText returns the title text element.
func (c *Card) RenderTitleText() string {
	return fmt.Sprintf(
		`  <g transform="translate(%g, %g)">
    <text class="header">%s</text>
  </g>`,
		c.PaddingX, c.PaddingY, escapeXML(c.Title),
	)
}

// RenderFooter returns the closing SVG tag.
func (c *Card) RenderFooter() string {
	return `</svg>`
}

// StatRow represents a single stat row with icon, label, and value.
type StatRow struct {
	Icon  IconType
	Label string
	Value string
}

// RenderStatRows renders multiple stat rows starting at the given Y position.
func (c *Card) RenderStatRows(rows []StatRow, startY float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`  <g transform="translate(%g, %g)">`, c.PaddingX, startY))
	sb.WriteString("\n")

	for i, row := range rows {
		y := float64(i) * DefaultStatRowHeight

		// Render icon
		sb.WriteString(fmt.Sprintf(`    <g transform="translate(0, %g)">`, y))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`      <svg width="%d" height="%d" viewBox="0 0 16 16" class="icon">`, DefaultIconSize, DefaultIconSize))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`        %s`, RenderIconInline(row.Icon, c.Theme.IconColor)))
		sb.WriteString("\n")
		sb.WriteString(`      </svg>`)
		sb.WriteString("\n")

		// Render label
		sb.WriteString(fmt.Sprintf(`      <text class="stat" x="25" y="12.5">%s:</text>`, escapeXML(row.Label)))
		sb.WriteString("\n")

		// Render value (right-aligned)
		valueX := c.Width - c.PaddingX*2 - 10
		sb.WriteString(fmt.Sprintf(`      <text class="stat-value" x="%g" y="12.5" text-anchor="end">%s</text>`, valueX, escapeXML(row.Value)))
		sb.WriteString("\n")
		sb.WriteString(`    </g>`)
		sb.WriteString("\n")
	}

	sb.WriteString(`  </g>`)
	return sb.String()
}

// CalculateHeight returns the required height for a card with the given number of stat rows.
func CalculateHeight(numRows int) float64 {
	// Title area: ~45px
	// Each row: 25px
	// Bottom padding: 20px
	return 45 + float64(numRows)*DefaultStatRowHeight + 20
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
