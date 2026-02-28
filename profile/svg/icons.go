package svg

import "fmt"

// IconType represents the type of icon to render.
type IconType string

const (
	IconStar     IconType = "star"
	IconCommit   IconType = "commit"
	IconPR       IconType = "pr"
	IconIssue    IconType = "issue"
	IconCode     IconType = "code"
	IconRepo     IconType = "repo"
	IconReview   IconType = "review"
	IconCalendar IconType = "calendar"
	IconStreak   IconType = "streak"
)

// iconPaths contains SVG path data for each icon type.
// Icons are designed for a 16x16 viewBox.
var iconPaths = map[IconType]string{
	// Star icon (GitHub star)
	IconStar: "M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.416-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z",

	// Commit icon (git commit)
	IconCommit: "M11.93 8.5a4.002 4.002 0 01-7.86 0H.75a.75.75 0 010-1.5h3.32a4.002 4.002 0 017.86 0h3.32a.75.75 0 010 1.5h-3.32zm-1.43-.75a2.5 2.5 0 10-5 0 2.5 2.5 0 005 0z",

	// Pull request icon
	IconPR: "M7.177 3.073L9.573.677A.25.25 0 0110 .854v4.792a.25.25 0 01-.427.177L7.177 3.427a.25.25 0 010-.354zM3.75 2.5a.75.75 0 100 1.5.75.75 0 000-1.5zm-2.25.75a2.25 2.25 0 113 2.122v5.256a2.251 2.251 0 11-1.5 0V5.372A2.25 2.25 0 011.5 3.25zM11 2.5h-1V4h1a1 1 0 011 1v5.628a2.251 2.251 0 101.5 0V5A2.5 2.5 0 0011 2.5zm1 10.25a.75.75 0 111.5 0 .75.75 0 01-1.5 0zM3.75 12a.75.75 0 100 1.5.75.75 0 000-1.5z",

	// Issue icon (circle with dot)
	IconIssue: "M8 9.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3z M8 0a8 8 0 100 16A8 8 0 008 0zM1.5 8a6.5 6.5 0 1113 0 6.5 6.5 0 01-13 0z",

	// Code icon (angle brackets)
	IconCode: "M4.72 3.22a.75.75 0 011.06 1.06L2.56 7.5l3.22 3.22a.75.75 0 11-1.06 1.06l-3.75-3.75a.75.75 0 010-1.06l3.75-3.75zm6.56 0a.75.75 0 10-1.06 1.06L13.44 7.5l-3.22 3.22a.75.75 0 101.06 1.06l3.75-3.75a.75.75 0 000-1.06l-3.75-3.75z",

	// Repository icon
	IconRepo: "M2 2.5A2.5 2.5 0 014.5 0h8.75a.75.75 0 01.75.75v12.5a.75.75 0 01-.75.75h-2.5a.75.75 0 110-1.5h1.75v-2h-8a1 1 0 00-.714 1.7.75.75 0 01-1.072 1.05A2.495 2.495 0 012 11.5v-9zm10.5-1V9h-8c-.356 0-.694.074-1 .208V2.5a1 1 0 011-1h8zM5 12.25v3.25a.25.25 0 00.4.2l1.45-1.087a.25.25 0 01.3 0L8.6 15.7a.25.25 0 00.4-.2v-3.25a.25.25 0 00-.25-.25h-3.5a.25.25 0 00-.25.25z",

	// Review icon (eye)
	IconReview: "M8 2c1.981 0 3.671.992 4.933 2.078 1.27 1.091 2.187 2.345 2.637 3.023a1.62 1.62 0 010 1.798c-.45.678-1.367 1.932-2.637 3.023C11.67 13.008 9.981 14 8 14c-1.981 0-3.671-.992-4.933-2.078C1.797 10.831.88 9.577.43 8.899a1.62 1.62 0 010-1.798c.45-.678 1.367-1.932 2.637-3.023C4.33 2.992 6.019 2 8 2zM1.679 7.932a.12.12 0 000 .136c.411.622 1.241 1.75 2.366 2.717C5.176 11.758 6.527 12.5 8 12.5c1.473 0 2.824-.742 3.955-1.715 1.124-.967 1.954-2.096 2.366-2.717a.12.12 0 000-.136c-.412-.621-1.242-1.75-2.366-2.717C10.824 4.242 9.473 3.5 8 3.5c-1.473 0-2.824.742-3.955 1.715-1.124.967-1.954 2.096-2.366 2.717zM8 10a2 2 0 100-4 2 2 0 000 4z",

	// Calendar icon
	IconCalendar: "M4.75 0a.75.75 0 01.75.75V2h5V.75a.75.75 0 011.5 0V2h1.25c.966 0 1.75.784 1.75 1.75v10.5A1.75 1.75 0 0113.25 16H2.75A1.75 1.75 0 011 14.25V3.75C1 2.784 1.784 2 2.75 2H4V.75A.75.75 0 014.75 0zm0 3.5h-.5a.25.25 0 00-.25.25V5h8V3.75a.25.25 0 00-.25-.25H4.75zm-2.25 3v7.75c0 .138.112.25.25.25h10.5a.25.25 0 00.25-.25V6.5z",

	// Streak/fire icon
	IconStreak: "M7.998 14.5c2.832 0 5-1.98 5-4.5 0-1.463-.68-2.19-1.879-3.383l-.036-.037c-1.013-1.008-2.3-2.29-2.834-4.434a.217.217 0 00-.36-.1l-.007.007c-.777.818-1.318 1.775-1.612 2.678-.297.907-.419 1.756-.439 2.343l-.005.168c.002.334.005.56-.012.745-.024.267-.081.423-.173.534-.066.077-.136.131-.214.174l-.039.02c-.12.057-.26.094-.468.108l-.096.004c-.26 0-.616-.149-.987-.367a3.96 3.96 0 01-.663-.494l-.07-.068a5.031 5.031 0 00-.707-.602c-.652-.455-1.2-.699-1.485-.699a.217.217 0 00-.192.316c.497.89 1.17 1.622 1.974 2.29a7.005 7.005 0 002.81 1.505l.057.015c.406.096.82.141 1.242.141zm3.25-4.5c0 1.658-1.457 3-3.25 3-1.792 0-3.25-1.342-3.25-3 0-.854.336-1.494.822-2.038l.089-.095c.414-.428.91-.82 1.378-1.348a.22.22 0 01.362.055c.347.713.858 1.295 1.41 1.83.552.536 1.152 1.03 1.59 1.649.228.323.35.645.35.947z",
}

// RenderIcon returns an SVG group element containing the icon.
// The icon is rendered at the specified position with the given size and color.
func RenderIcon(iconType IconType, x, y, size float64, color string) string {
	path, ok := iconPaths[iconType]
	if !ok {
		return ""
	}

	return fmt.Sprintf(
		`<g transform="translate(%g, %g)"><svg width="%g" height="%g" viewBox="0 0 16 16"><path fill="%s" d="%s"/></svg></g>`,
		x, y, size, size, color, path,
	)
}

// RenderIconInline returns just the path element for embedding in an existing SVG.
func RenderIconInline(iconType IconType, color string) string {
	path, ok := iconPaths[iconType]
	if !ok {
		return ""
	}
	return fmt.Sprintf(`<path fill="%s" d="%s"/>`, color, path)
}
