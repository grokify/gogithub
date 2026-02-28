package chart

// ConventionalCommitType represents a conventional commit type.
type ConventionalCommitType string

// Conventional commit type constants.
const (
	CCFeat     ConventionalCommitType = "feat"
	CCFix      ConventionalCommitType = "fix"
	CCDocs     ConventionalCommitType = "docs"
	CCStyle    ConventionalCommitType = "style"
	CCRefactor ConventionalCommitType = "refactor"
	CCPerf     ConventionalCommitType = "perf"
	CCTest     ConventionalCommitType = "test"
	CCBuild    ConventionalCommitType = "build"
	CCCI       ConventionalCommitType = "ci"
	CCChore    ConventionalCommitType = "chore"
	CCRevert   ConventionalCommitType = "revert"
	CCSecurity ConventionalCommitType = "security"
	CCDeps     ConventionalCommitType = "deps"
	CCOther    ConventionalCommitType = "other"
)

// ConventionalCommitTypes lists all CC types in display order.
var ConventionalCommitTypes = []ConventionalCommitType{
	CCFeat, CCFix, CCRefactor, CCDocs, CCTest, CCChore, CCBuild, CCCI, CCPerf, CCDeps, CCSecurity, CCStyle, CCRevert, CCOther,
}

// CCTypeColors maps CC types to colors.
var CCTypeColors = map[ConventionalCommitType]string{
	CCFeat:     "#2ea043", // green
	CCFix:      "#da3633", // red
	CCRefactor: "#58a6ff", // blue
	CCDocs:     "#d29922", // yellow
	CCTest:     "#a371f7", // purple
	CCChore:    "#8b949e", // gray
	CCBuild:    "#f78166", // orange
	CCCI:       "#db61a2", // pink
	CCPerf:     "#3fb950", // light green
	CCDeps:     "#79c0ff", // light blue
	CCSecurity: "#f85149", // bright red
	CCStyle:    "#6e7681", // dark gray
	CCRevert:   "#ff7b72", // salmon
	CCOther:    "#484f58", // darker gray
}

// CCTypeLabels provides human-readable labels for CC types.
var CCTypeLabels = map[ConventionalCommitType]string{
	CCFeat:     "Features",
	CCFix:      "Fixes",
	CCRefactor: "Refactor",
	CCDocs:     "Docs",
	CCTest:     "Tests",
	CCChore:    "Chore",
	CCBuild:    "Build",
	CCCI:       "CI",
	CCPerf:     "Perf",
	CCDeps:     "Deps",
	CCSecurity: "Security",
	CCStyle:    "Style",
	CCRevert:   "Revert",
	CCOther:    "Other",
}

// ChangelogCategory represents a structured-changelog category.
type ChangelogCategory string

// Changelog category constants (matching structured-changelog).
const (
	CLAdded          ChangelogCategory = "Added"
	CLChanged        ChangelogCategory = "Changed"
	CLFixed          ChangelogCategory = "Fixed"
	CLSecurity       ChangelogCategory = "Security"
	CLPerformance    ChangelogCategory = "Performance"
	CLDeprecated     ChangelogCategory = "Deprecated"
	CLRemoved        ChangelogCategory = "Removed"
	CLBreaking       ChangelogCategory = "Breaking"
	CLDependencies   ChangelogCategory = "Dependencies"
	CLDocumentation  ChangelogCategory = "Documentation"
	CLBuild          ChangelogCategory = "Build"
	CLTests          ChangelogCategory = "Tests"
	CLInfrastructure ChangelogCategory = "Infrastructure"
	CLInternal       ChangelogCategory = "Internal"
	CLOther          ChangelogCategory = "Other"
)

// ChangelogCategories lists categories in display order (stakeholder priority).
var ChangelogCategories = []ChangelogCategory{
	CLAdded, CLFixed, CLChanged, CLSecurity, CLPerformance, CLBreaking,
	CLDependencies, CLDocumentation, CLBuild, CLTests, CLInfrastructure, CLInternal, CLOther,
}

// CLCategoryColors maps changelog categories to colors.
var CLCategoryColors = map[ChangelogCategory]string{
	CLAdded:          "#2ea043", // green
	CLFixed:          "#da3633", // red
	CLChanged:        "#58a6ff", // blue
	CLSecurity:       "#f85149", // bright red
	CLPerformance:    "#3fb950", // light green
	CLDeprecated:     "#d29922", // yellow
	CLRemoved:        "#ff7b72", // salmon
	CLBreaking:       "#f85149", // bright red
	CLDependencies:   "#79c0ff", // light blue
	CLDocumentation:  "#d29922", // yellow
	CLBuild:          "#f78166", // orange
	CLTests:          "#a371f7", // purple
	CLInfrastructure: "#db61a2", // pink
	CLInternal:       "#8b949e", // gray
	CLOther:          "#484f58", // darker gray
}

// CCToChangelogCategory maps conventional commit types to changelog categories.
var CCToChangelogCategory = map[ConventionalCommitType]ChangelogCategory{
	CCFeat:     CLAdded,
	CCFix:      CLFixed,
	CCDocs:     CLDocumentation,
	CCStyle:    CLInternal,
	CCRefactor: CLChanged,
	CCPerf:     CLPerformance,
	CCTest:     CLTests,
	CCBuild:    CLBuild,
	CCCI:       CLInfrastructure,
	CCChore:    CLInternal,
	CCRevert:   CLFixed,
	CCSecurity: CLSecurity,
	CCDeps:     CLDependencies,
	CCOther:    CLOther,
}

// MonthlyCommitTypes holds commit type counts for a single month.
type MonthlyCommitTypes struct {
	YearMonth string         `json:"year_month"` // "2025-01"
	Year      int            `json:"year"`
	Month     int            `json:"month"`
	MonthName string         `json:"month_name"` // "Jan"
	Total     int            `json:"total"`
	ByCCType  map[string]int `json:"by_cc_type"`            // Conventional commit types
	ByCLCat   map[string]int `json:"by_changelog_category"` // Changelog categories
}

// CommitTypeData holds commit type analysis data.
type CommitTypeData struct {
	Username   string               `json:"username"`
	From       string               `json:"from"`
	To         string               `json:"to"`
	TotalCount int                  `json:"total_count"`
	Monthly    []MonthlyCommitTypes `json:"monthly"`
	Summary    CommitTypeSummary    `json:"summary"`
}

// CommitTypeSummary holds aggregate commit type counts.
type CommitTypeSummary struct {
	ByCCType map[string]int `json:"by_cc_type"`
	ByCLCat  map[string]int `json:"by_changelog_category"`
}
