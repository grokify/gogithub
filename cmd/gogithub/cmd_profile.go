package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/grokify/gogithub/graphql"
	"github.com/grokify/gogithub/profile"
	"github.com/grokify/gogithub/profile/readme"
	"github.com/grokify/gogithub/profile/svg"
	"github.com/grokify/mogo/fmt/progress"
	"github.com/spf13/cobra"
)

var (
	profileUser              string
	profileFrom              string
	profileTo                string
	profileFormat            string
	profileOutput            string
	profileOutputRaw         string
	profileOutputAggregate   string
	profileOutputMonthly     string
	profileOutputMonthlyDir  string
	profileOutputReadme      string
	profileReadmeConfig      string
	profileOutputSVG         string
	profileSVGTheme          string
	profileSVGTitle          string
	profileOutputChart       string
	profileOutputChartJSON   string
	profileInput             string
	profileIncludeReleases   bool
	profileReleaseOrgs       string
	profileVisibility        string
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Fetch user contribution statistics",
	Long: `Fetch comprehensive GitHub contribution statistics for a user.
This provides data similar to what's shown on GitHub profile pages.

Examples:
  # Human-readable summary
  gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31

  # Aggregate JSON output
  gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31 --format json

  # Generate both raw and aggregate JSON files
  gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31 \
    --output-raw raw.json --output-aggregate aggregate.json

  # Include release counts (requires additional API calls)
  gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31 \
    --include-releases --output-raw raw.json

  # Generate aggregate from existing raw file (no API calls)
  gogithub profile --input raw.json --output aggregate.json

  # Generate README from API
  gogithub profile --user grokify --output-readme README.md --readme-config config.json

  # Generate README from existing raw JSON
  gogithub profile --input raw.json --output-readme README.md --readme-config config.json

  # Generate SVG stats card
  gogithub profile --user grokify --output-svg stats.svg

  # Generate SVG with custom theme
  gogithub profile --user grokify --output-svg stats.svg --svg-theme dark

  # Generate SVG from existing raw JSON (no API calls)
  gogithub profile --input raw.json --output-svg stats.svg --svg-theme tokyonight

  # Generate monthly lines chart (JSON IR + SVG)
  gogithub profile --user grokify --output-chart-json chart.json --output-chart chart.svg

  # Generate chart from existing raw JSON
  gogithub profile --input raw.json --output-chart lines.svg --svg-theme dark

  # Generate monthly activity JSON (creates or merges with existing file)
  gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31 \
    --output-monthly monthly.json

  # Add more months to existing file (automatically merges)
  gogithub profile --user grokify --from 2024-02-01 --to 2024-02-28 \
    --output-monthly monthly.json

  # Generate separate monthly files in a directory
  gogithub profile --user grokify --from 2024-01-01 --to 2024-03-31 \
    --output-monthly-dir ./stats/

  # Include releases only from specific orgs
  gogithub profile --user grokify --from 2024-01-01 --to 2024-03-31 \
    --include-releases --release-orgs grokify,plexusone \
    --output-monthly-dir ./stats/

  # Public repos only
  gogithub profile --user grokify --visibility public --output-monthly monthly.json

  # Private repos only
  gogithub profile --user grokify --visibility private --output-monthly monthly.json

Environment:
  GITHUB_TOKEN    Required for fetching from API. Not needed with --input.
                  Use a fine-grained token with "Public Repositories (read-only)"`,
	RunE: runProfile,
}

func init() {
	profileCmd.Flags().StringVarP(&profileUser, "user", "u", "", "GitHub username")
	profileCmd.Flags().StringVarP(&profileFrom, "from", "f", "", "Start date (YYYY-MM-DD), defaults to 1 year ago")
	profileCmd.Flags().StringVarP(&profileTo, "to", "t", "", "End date (YYYY-MM-DD), defaults to today")
	profileCmd.Flags().StringVar(&profileFormat, "format", "summary", "Output format: summary, json")
	profileCmd.Flags().StringVarP(&profileOutput, "output", "o", "", "Output file (defaults to stdout)")
	profileCmd.Flags().StringVar(&profileOutputRaw, "output-raw", "", "Output raw JSON file (includes all per-repo data)")
	profileCmd.Flags().StringVar(&profileOutputAggregate, "output-aggregate", "", "Output aggregate JSON file")
	profileCmd.Flags().StringVar(&profileOutputMonthly, "output-monthly", "", "Output monthly JSON file (merges with existing)")
	profileCmd.Flags().StringVar(&profileOutputReadme, "output-readme", "", "Output README.md file")
	profileCmd.Flags().StringVar(&profileReadmeConfig, "readme-config", "", "README config JSON file")
	profileCmd.Flags().StringVar(&profileOutputSVG, "output-svg", "", "Output SVG stats card file")
	profileCmd.Flags().StringVar(&profileSVGTheme, "svg-theme", "default", "SVG theme: default, dark, radical, tokyonight, gruvbox, dracula, nord, catppuccin")
	profileCmd.Flags().StringVar(&profileSVGTitle, "svg-title", "", "Custom title for SVG card (default: username's GitHub Stats)")
	profileCmd.Flags().StringVar(&profileOutputChart, "output-chart", "", "Output monthly lines chart SVG file")
	profileCmd.Flags().StringVar(&profileOutputChartJSON, "output-chart-json", "", "Output monthly lines chart JSON IR file")
	profileCmd.Flags().StringVarP(&profileInput, "input", "i", "", "Input raw JSON file (skips API calls)")
	profileCmd.Flags().BoolVar(&profileIncludeReleases, "include-releases", false, "Fetch release counts for contributed repositories")
	profileCmd.Flags().StringVar(&profileReleaseOrgs, "release-orgs", "", "Comma-separated list of orgs/owners to count releases for (e.g., grokify,plexusone)")
	profileCmd.Flags().StringVar(&profileOutputMonthlyDir, "output-monthly-dir", "", "Output directory for individual monthly JSON files")
	profileCmd.Flags().StringVar(&profileVisibility, "visibility", "all", "Repository visibility filter: all, public, private")
}

func runProfile(cmd *cobra.Command, args []string) error {
	// Mode 1: Read from input file
	if profileInput != "" {
		return runProfileFromInput()
	}

	// Mode 2: Fetch from API
	if profileUser == "" {
		return fmt.Errorf("--user is required when not using --input")
	}

	return runProfileFromAPI()
}

func runProfileFromInput() error {
	fmt.Fprintf(os.Stderr, "Reading from %s...\n", profileInput)

	data, err := os.ReadFile(profileInput)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	var raw RawJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse input JSON: %w", err)
	}

	p := rawToProfile(&raw)

	// Build opts from CLI flags for metadata
	visibility, err := parseVisibility(profileVisibility)
	if err != nil {
		return err
	}
	opts := &profile.Options{
		Visibility:      visibility,
		IncludeReleases: profileIncludeReleases,
		ReleaseOrgs:     parseCommaSeparated(profileReleaseOrgs),
	}

	// Generate SVG if requested
	if profileOutputSVG != "" {
		if err := generateSVG(p, profileOutputSVG, profileSVGTheme, profileSVGTitle); err != nil {
			return err
		}
	}

	// Generate chart JSON IR if requested
	if profileOutputChartJSON != "" {
		if err := generateChartJSON(p, profileOutputChartJSON, profileSVGTheme); err != nil {
			return err
		}
	}

	// Generate chart SVG if requested
	if profileOutputChart != "" {
		if err := generateChartSVG(p, profileOutputChart, profileSVGTheme, profileSVGTitle); err != nil {
			return err
		}
	}

	// Generate README if requested
	if profileOutputReadme != "" {
		if err := generateReadme(p, profileOutputReadme, profileReadmeConfig); err != nil {
			return err
		}
	}

	// Generate monthly JSON if requested (with merge)
	if profileOutputMonthly != "" {
		if err := writeMonthlyOutput(p, profileOutputMonthly, opts); err != nil {
			return err
		}
	}

	// Generate individual monthly files in directory
	if profileOutputMonthlyDir != "" {
		written, err := profile.WriteMonthlyFiles(profileOutputMonthlyDir, p, opts)
		if err != nil {
			return fmt.Errorf("write monthly files: %w", err)
		}
		for _, path := range written {
			fmt.Fprintf(os.Stderr, "Wrote %s\n", path)
		}
	}

	// If specific outputs were requested, return early
	if profileOutputSVG != "" || profileOutputChartJSON != "" || profileOutputChart != "" || profileOutputReadme != "" || profileOutputMonthly != "" || profileOutputMonthlyDir != "" {
		if profileOutput == "" {
			return nil
		}
	}

	// Generate aggregate from raw
	aggregate := rawToAggregate(&raw)

	// Output aggregate
	output, err := json.MarshalIndent(aggregate, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal aggregate: %w", err)
	}

	return writeOutput(string(output)+"\n", profileOutput, "aggregate")
}

func runProfileFromAPI() error {
	token := ensureToken()

	// Parse dates
	from, to, err := parseDateRange(profileFrom, profileTo)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Create clients
	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	fmt.Fprintf(os.Stderr, "Fetching profile for '%s' from %s to %s\n\n",
		profileUser, from.Format("2006-01-02"), to.Format("2006-01-02"))

	// Create progress renderer
	renderer := progress.NewMultiStageRenderer(os.Stderr)

	// Progress callback that converts profile.ProgressInfo to progress.StageInfo
	progressFunc := func(info profile.ProgressInfo) {
		renderer.Update(progress.StageInfo{
			Stage:       info.Stage,
			TotalStages: info.TotalStages,
			Description: info.Description,
			Current:     info.Current,
			Total:       info.Total,
			Done:        info.Done,
		})
	}

	visibility, err := parseVisibility(profileVisibility)
	if err != nil {
		return err
	}

	opts := &profile.Options{
		Visibility:      visibility,
		IncludeReleases: profileIncludeReleases,
		ReleaseOrgs:     parseCommaSeparated(profileReleaseOrgs),
		Progress:        progressFunc,
	}

	p, err := profile.GetUserProfile(ctx, restClient, gqlClient, profileUser, from, to, opts)
	if err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	// Mode: Generate specific output files
	if profileOutputRaw != "" || profileOutputAggregate != "" || profileOutputMonthly != "" || profileOutputMonthlyDir != "" || profileOutputReadme != "" || profileOutputSVG != "" || profileOutputChart != "" || profileOutputChartJSON != "" {
		return outputBothFormats(p, opts)
	}

	// Mode: Single output (legacy behavior)
	var output string
	switch profileFormat {
	case "json":
		output, err = formatAggregateJSON(p)
	case "summary":
		output = formatSummary(p)
	default:
		return fmt.Errorf("unknown format: %s (use 'summary' or 'json')", profileFormat)
	}
	if err != nil {
		return err
	}

	return writeOutput(output, profileOutput, "output")
}

func outputBothFormats(p *profile.UserProfile, opts *profile.Options) error {
	// Generate and write raw JSON
	if profileOutputRaw != "" {
		raw := profileToRaw(p)
		data, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal raw JSON: %w", err)
		}
		if err := writeOutput(string(data)+"\n", profileOutputRaw, "raw"); err != nil {
			return err
		}
	}

	// Generate and write aggregate JSON
	if profileOutputAggregate != "" {
		aggregate := profileToAggregate(p)
		data, err := json.MarshalIndent(aggregate, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal aggregate JSON: %w", err)
		}
		if err := writeOutput(string(data)+"\n", profileOutputAggregate, "aggregate"); err != nil {
			return err
		}
	}

	// Generate and write monthly JSON (with merge)
	if profileOutputMonthly != "" {
		if err := writeMonthlyOutput(p, profileOutputMonthly, opts); err != nil {
			return err
		}
	}

	// Generate individual monthly files in directory
	if profileOutputMonthlyDir != "" {
		written, err := profile.WriteMonthlyFiles(profileOutputMonthlyDir, p, opts)
		if err != nil {
			return fmt.Errorf("write monthly files: %w", err)
		}
		for _, path := range written {
			fmt.Fprintf(os.Stderr, "Wrote %s\n", path)
		}
	}

	// Generate README if requested
	if profileOutputReadme != "" {
		if err := generateReadme(p, profileOutputReadme, profileReadmeConfig); err != nil {
			return err
		}
	}

	// Generate SVG if requested
	if profileOutputSVG != "" {
		if err := generateSVG(p, profileOutputSVG, profileSVGTheme, profileSVGTitle); err != nil {
			return err
		}
	}

	// Generate chart JSON IR if requested
	if profileOutputChartJSON != "" {
		if err := generateChartJSON(p, profileOutputChartJSON, profileSVGTheme); err != nil {
			return err
		}
	}

	// Generate chart SVG if requested
	if profileOutputChart != "" {
		if err := generateChartSVG(p, profileOutputChart, profileSVGTheme, profileSVGTitle); err != nil {
			return err
		}
	}

	return nil
}

func writeOutput(content, filename, label string) error {
	if filename != "" {
		if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
			return fmt.Errorf("write %s file: %w", label, err)
		}
		fmt.Fprintf(os.Stderr, "Wrote %s\n", filename)
	} else {
		fmt.Print(content)
	}
	return nil
}

func parseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	var from, to time.Time
	var err error

	if toStr == "" {
		to = time.Now()
	} else {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to date: %w", err)
		}
		// Set to end of day
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.UTC)
	}

	if fromStr == "" {
		from = to.AddDate(-1, 0, 0) // Default to 1 year ago
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from date: %w", err)
		}
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	}

	return from, to, nil
}

// RawJSON contains all data needed to regenerate aggregates without API calls.
type RawJSON struct {
	Username    string    `json:"username"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	GeneratedAt time.Time `json:"generatedAt"`

	// Summary counts - GitHub official count from contributionsCollection
	TotalCommits            int `json:"totalCommits"`
	TotalIssues             int `json:"totalIssues"`
	TotalPRs                int `json:"totalPrs"`
	TotalReviews            int `json:"totalReviews"`
	TotalReposCreated       int `json:"totalReposCreated"`
	RestrictedContributions int `json:"restrictedContributions,omitempty"`

	// Commits from default branch traversal (may differ from totalCommits)
	// This count only includes commits on default branches and may miss
	// feature branches, squash-merged commits, and inaccessible repos.
	CommitsDefaultBranch int `json:"commitsDefaultBranch"`

	// Code stats (from default branch traversal)
	TotalAdditions int `json:"totalAdditions"`
	TotalDeletions int `json:"totalDeletions"`

	// Release stats (optional, requires IncludeReleases option)
	TotalReleases int `json:"totalReleases,omitempty"`

	// Per-repo details (full data)
	Repos []RepoJSON `json:"repos"`

	// Monthly breakdown (full data, from default branch traversal)
	Monthly []MonthlyJSON `json:"monthly"`

	// Calendar data
	Calendar *CalendarDataJSON `json:"calendar,omitempty"`
}

// AggregateJSON is the summarized output structure.
type AggregateJSON struct {
	Username    string    `json:"username"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	GeneratedAt time.Time `json:"generatedAt"`

	// Summary counts - GitHub official count from contributionsCollection
	TotalCommits            int `json:"totalCommits"`
	TotalIssues             int `json:"totalIssues"`
	TotalPRs                int `json:"totalPrs"`
	TotalReviews            int `json:"totalReviews"`
	TotalReposCreated       int `json:"totalReposCreated"`
	RestrictedContributions int `json:"restrictedContributions,omitempty"`

	// Commits from default branch traversal (may differ from totalCommits)
	CommitsDefaultBranch int `json:"commitsDefaultBranch"`

	// Code stats (from default branch traversal)
	TotalAdditions int `json:"totalAdditions"`
	TotalDeletions int `json:"totalDeletions"`
	NetAdditions   int `json:"netAdditions"`

	// Release stats (optional, requires IncludeReleases option)
	TotalReleases int `json:"totalReleases,omitempty"`

	// Repo summary
	ReposContributedTo int `json:"reposContributedTo"`

	// Calendar stats (computed)
	Calendar *CalendarStatsJSON `json:"calendar,omitempty"`

	// Monthly breakdown (from default branch traversal)
	Monthly []MonthlyJSON `json:"monthly,omitempty"`
}

type CalendarDataJSON struct {
	TotalContributions int            `json:"totalContributions"`
	Weeks              []CalendarWeek `json:"weeks,omitempty"`
}

type CalendarWeek struct {
	StartDate string        `json:"startDate"`
	Days      []CalendarDay `json:"days"`
}

type CalendarDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
	Level             int    `json:"level"`
}

type CalendarStatsJSON struct {
	TotalContributions    int `json:"totalContributions"`
	DaysWithContributions int `json:"daysWithContributions"`
	LongestStreak         int `json:"longestStreak"`
	CurrentStreak         int `json:"currentStreak"`
}

type MonthlyJSON struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	MonthName string `json:"monthName"`
	Commits   int    `json:"commits"`
	Issues    int    `json:"issues"`
	PRs       int    `json:"prs"`
	Reviews   int    `json:"reviews"`
	Releases  int    `json:"releases"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	// Computed fields for enhanced monthly stats
	NetAdditions         int `json:"netAdditions"`
	RepoCountContributed int `json:"repoCountContributed"`
	RepoCountCreated     int `json:"repoCountCreated"`
	// Underlying data for repo tracking
	CommitsByRepo map[string]int `json:"commitsByRepo,omitempty"`
	ReposCreated  []string       `json:"reposCreated,omitempty"`
}

type RepoJSON struct {
	FullName  string `json:"fullName"`
	IsPrivate bool   `json:"isPrivate"`
	Commits   int    `json:"commits"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// MonthlyOutputJSON is the output structure for --output-monthly.
// This format is designed for incremental updates and merging.
// QueryMetadataJSON captures the parameters used to generate the output.
type QueryMetadataJSON struct {
	Username        string    `json:"username"`
	From            string    `json:"from"`                     // YYYY-MM-DD format
	To              string    `json:"to"`                       // YYYY-MM-DD format
	Visibility      string    `json:"visibility"`               // all, public, private
	IncludeReleases bool      `json:"includeReleases"`
	ReleaseOrgs     []string  `json:"releaseOrgs,omitempty"`
	GeneratedAt     time.Time `json:"generatedAt"`
}

type MonthlyOutputJSON struct {
	Metadata QueryMetadataJSON `json:"metadata"`
	Username string            `json:"username"`
	Months   []MonthlyJSON     `json:"months"`
}

// writeMonthlyOutput writes monthly data to a file, merging with existing data if present.
// Months are sorted in descending chronological order (newest first).
func writeMonthlyOutput(p *profile.UserProfile, outputPath string, opts *profile.Options) error {
	// Build new months from profile
	newMonths := make([]MonthlyJSON, 0)
	if p.Activity != nil {
		for _, m := range p.Activity.Months {
			newMonths = append(newMonths, MonthlyJSON{
				Year:                 m.Year,
				Month:                int(m.Month),
				MonthName:            m.MonthName(),
				Commits:              m.Commits,
				Issues:               m.Issues,
				PRs:                  m.PRs,
				Reviews:              m.Reviews,
				Releases:             m.Releases,
				Additions:            m.Additions,
				Deletions:            m.Deletions,
				NetAdditions:         m.NetAdditions(),
				RepoCountContributed: m.CommitRepoCount(),
				RepoCountCreated:     m.RepoCountCreated(),
				CommitsByRepo:        m.CommitsByRepo,
				ReposCreated:         m.ReposCreated,
			})
		}
	}

	// Try to read existing file
	var existing MonthlyOutputJSON
	existingData, err := os.ReadFile(outputPath)
	if err == nil {
		if err := json.Unmarshal(existingData, &existing); err != nil {
			return fmt.Errorf("parse existing monthly file: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing monthly file: %w", err)
	}

	// Merge: create map of existing months, then update/add new months
	monthMap := make(map[string]MonthlyJSON)
	for _, m := range existing.Months {
		key := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
		monthMap[key] = m
	}
	for _, m := range newMonths {
		key := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
		monthMap[key] = m // Overwrites existing or adds new
	}

	// Convert map back to slice
	merged := make([]MonthlyJSON, 0, len(monthMap))
	for _, m := range monthMap {
		merged = append(merged, m)
	}

	// Sort descending by year, then month (newest first)
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Year != merged[j].Year {
			return merged[i].Year > merged[j].Year
		}
		return merged[i].Month > merged[j].Month
	})

	// Build metadata
	visibility := "all"
	var releaseOrgs []string
	includeReleases := false
	if opts != nil {
		switch opts.Visibility {
		case graphql.VisibilityPublic:
			visibility = "public"
		case graphql.VisibilityPrivate:
			visibility = "private"
		}
		releaseOrgs = opts.ReleaseOrgs
		includeReleases = opts.IncludeReleases
	}

	// Build output
	output := MonthlyOutputJSON{
		Metadata: QueryMetadataJSON{
			Username:        p.Username,
			From:            p.From.Format("2006-01-02"),
			To:              p.To.Format("2006-01-02"),
			Visibility:      visibility,
			IncludeReleases: includeReleases,
			ReleaseOrgs:     releaseOrgs,
			GeneratedAt:     time.Now().UTC(),
		},
		Username: p.Username,
		Months:   merged,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal monthly JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write monthly file: %w", err)
	}

	action := "Wrote"
	if len(existing.Months) > 0 {
		action = "Merged"
	}
	fmt.Fprintf(os.Stderr, "%s %s (%d months)\n", action, outputPath, len(merged))
	return nil
}

func profileToRaw(p *profile.UserProfile) *RawJSON {
	raw := &RawJSON{
		Username:                p.Username,
		From:                    p.From,
		To:                      p.To,
		GeneratedAt:             time.Now().UTC(),
		TotalCommits:            p.TotalCommits,
		TotalIssues:             p.TotalIssues,
		TotalPRs:                p.TotalPRs,
		TotalReviews:            p.TotalReviews,
		TotalReposCreated:       p.TotalReposCreated,
		RestrictedContributions: p.RestrictedContributions,
		CommitsDefaultBranch:    p.CommitsDefaultBranch,
		TotalAdditions:          p.TotalAdditions,
		TotalDeletions:          p.TotalDeletions,
		Repos:                   []RepoJSON{},
		Monthly:                 []MonthlyJSON{},
	}

	// Repos
	for _, r := range p.RepoStats {
		raw.Repos = append(raw.Repos, RepoJSON{
			FullName:  r.FullName,
			IsPrivate: r.IsPrivate,
			Commits:   r.Commits,
			Additions: r.Additions,
			Deletions: r.Deletions,
		})
	}

	// Monthly
	totalReleases := 0
	if p.Activity != nil {
		for _, m := range p.Activity.Months {
			raw.Monthly = append(raw.Monthly, MonthlyJSON{
				Year:                 m.Year,
				Month:                int(m.Month),
				MonthName:            m.MonthName(),
				Commits:              m.Commits,
				Issues:               m.Issues,
				PRs:                  m.PRs,
				Reviews:              m.Reviews,
				Releases:             m.Releases,
				Additions:            m.Additions,
				Deletions:            m.Deletions,
				NetAdditions:         m.NetAdditions(),
				RepoCountContributed: m.CommitRepoCount(),
				RepoCountCreated:     m.RepoCountCreated(),
				CommitsByRepo:        m.CommitsByRepo,
				ReposCreated:         m.ReposCreated,
			})
			totalReleases += m.Releases
		}
	}
	raw.TotalReleases = totalReleases

	// Calendar
	if p.Calendar != nil {
		raw.Calendar = &CalendarDataJSON{
			TotalContributions: p.Calendar.TotalContributions,
			Weeks:              []CalendarWeek{},
		}
		for _, w := range p.Calendar.Weeks {
			week := CalendarWeek{
				StartDate: w.StartDate.Format("2006-01-02"),
				Days:      []CalendarDay{},
			}
			for _, d := range w.Days {
				if !d.Date.IsZero() {
					week.Days = append(week.Days, CalendarDay{
						Date:              d.Date.Format("2006-01-02"),
						ContributionCount: d.ContributionCount,
						Level:             int(d.Level),
					})
				}
			}
			raw.Calendar.Weeks = append(raw.Calendar.Weeks, week)
		}
	}

	return raw
}

func profileToAggregate(p *profile.UserProfile) *AggregateJSON {
	agg := &AggregateJSON{
		Username:                p.Username,
		From:                    p.From,
		To:                      p.To,
		GeneratedAt:             time.Now().UTC(),
		TotalCommits:            p.TotalCommits,
		TotalIssues:             p.TotalIssues,
		TotalPRs:                p.TotalPRs,
		TotalReviews:            p.TotalReviews,
		TotalReposCreated:       p.TotalReposCreated,
		RestrictedContributions: p.RestrictedContributions,
		CommitsDefaultBranch:    p.CommitsDefaultBranch,
		TotalAdditions:          p.TotalAdditions,
		TotalDeletions:          p.TotalDeletions,
		NetAdditions:            p.TotalAdditions - p.TotalDeletions,
		ReposContributedTo:      p.ReposContributedTo,
		Monthly:                 []MonthlyJSON{},
	}

	// Calendar stats
	if p.Calendar != nil {
		agg.Calendar = &CalendarStatsJSON{
			TotalContributions:    p.Calendar.TotalContributions,
			DaysWithContributions: p.Calendar.DaysWithContributions(),
			LongestStreak:         p.Calendar.LongestStreak(),
			CurrentStreak:         p.Calendar.CurrentStreak(),
		}
	}

	// Monthly breakdown
	totalReleases := 0
	if p.Activity != nil {
		for _, m := range p.Activity.Months {
			agg.Monthly = append(agg.Monthly, MonthlyJSON{
				Year:                 m.Year,
				Month:                int(m.Month),
				MonthName:            m.MonthName(),
				Commits:              m.Commits,
				Issues:               m.Issues,
				PRs:                  m.PRs,
				Reviews:              m.Reviews,
				Releases:             m.Releases,
				Additions:            m.Additions,
				Deletions:            m.Deletions,
				NetAdditions:         m.NetAdditions(),
				RepoCountContributed: m.CommitRepoCount(),
				RepoCountCreated:     m.RepoCountCreated(),
				CommitsByRepo:        m.CommitsByRepo,
				ReposCreated:         m.ReposCreated,
			})
			totalReleases += m.Releases
		}
	}
	agg.TotalReleases = totalReleases

	return agg
}

func rawToAggregate(raw *RawJSON) *AggregateJSON {
	agg := &AggregateJSON{
		Username:                raw.Username,
		From:                    raw.From,
		To:                      raw.To,
		GeneratedAt:             time.Now().UTC(),
		TotalCommits:            raw.TotalCommits,
		TotalIssues:             raw.TotalIssues,
		TotalPRs:                raw.TotalPRs,
		TotalReviews:            raw.TotalReviews,
		TotalReposCreated:       raw.TotalReposCreated,
		RestrictedContributions: raw.RestrictedContributions,
		CommitsDefaultBranch:    raw.CommitsDefaultBranch,
		TotalAdditions:          raw.TotalAdditions,
		TotalDeletions:          raw.TotalDeletions,
		NetAdditions:            raw.TotalAdditions - raw.TotalDeletions,
		TotalReleases:           raw.TotalReleases,
		ReposContributedTo:      len(raw.Repos),
		Monthly:                 raw.Monthly,
	}

	// Compute calendar stats from raw calendar data
	if raw.Calendar != nil {
		daysWithContributions := 0
		longestStreak := 0
		currentStreak := 0

		// Flatten days and compute stats
		var allDays []struct {
			date  string
			count int
		}
		for _, w := range raw.Calendar.Weeks {
			for _, d := range w.Days {
				allDays = append(allDays, struct {
					date  string
					count int
				}{d.Date, d.ContributionCount})
				if d.ContributionCount > 0 {
					daysWithContributions++
				}
			}
		}

		// Compute streaks (simplified - assumes days are in order)
		streak := 0
		for _, d := range allDays {
			if d.count > 0 {
				streak++
				if streak > longestStreak {
					longestStreak = streak
				}
			} else {
				streak = 0
			}
		}

		// Current streak (from end)
		for i := len(allDays) - 1; i >= 0; i-- {
			if allDays[i].count > 0 {
				currentStreak++
			} else {
				break
			}
		}

		agg.Calendar = &CalendarStatsJSON{
			TotalContributions:    raw.Calendar.TotalContributions,
			DaysWithContributions: daysWithContributions,
			LongestStreak:         longestStreak,
			CurrentStreak:         currentStreak,
		}
	}

	return agg
}

func formatAggregateJSON(p *profile.UserProfile) (string, error) {
	agg := profileToAggregate(p)
	data, err := json.MarshalIndent(agg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// generateReadme creates a README file from profile data and optional config.
func generateReadme(p *profile.UserProfile, outputPath, configPath string) error {
	// Load config if provided
	var cfg *readme.Config
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("read readme config: %w", err)
		}
		cfg = &readme.Config{}
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("parse readme config: %w", err)
		}
	}

	// Create generator and generate README
	gen, err := readme.NewGenerator()
	if err != nil {
		return fmt.Errorf("create readme generator: %w", err)
	}

	if err := gen.GenerateToFile(p, cfg, outputPath); err != nil {
		return fmt.Errorf("generate readme: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", outputPath)
	return nil
}

// generateSVG creates an SVG stats card from profile data.
func generateSVG(p *profile.UserProfile, outputPath, themeName, title string) error {
	svgContent := svg.GenerateSVG(p, themeName, title)

	if err := os.WriteFile(outputPath, []byte(svgContent), 0600); err != nil {
		return fmt.Errorf("write SVG file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", outputPath)
	return nil
}

// generateChartJSON creates a JSON IR for the monthly lines chart.
func generateChartJSON(p *profile.UserProfile, outputPath, themeName string) error {
	jsonBytes, err := svg.GenerateMonthlyLinesJSON(p, themeName)
	if err != nil {
		return fmt.Errorf("generate chart JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0600); err != nil {
		return fmt.Errorf("write chart JSON file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", outputPath)
	return nil
}

// generateChartSVG creates an SVG monthly lines chart from profile data.
func generateChartSVG(p *profile.UserProfile, outputPath, themeName, title string) error {
	svgContent := svg.GenerateMonthlyLinesSVG(p, themeName, title)

	if err := os.WriteFile(outputPath, []byte(svgContent), 0600); err != nil {
		return fmt.Errorf("write chart SVG file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %s\n", outputPath)
	return nil
}

// rawToProfile converts RawJSON back to a UserProfile for README generation.
func rawToProfile(raw *RawJSON) *profile.UserProfile {
	p := &profile.UserProfile{
		Username:                raw.Username,
		From:                    raw.From,
		To:                      raw.To,
		TotalCommits:            raw.TotalCommits,
		TotalIssues:             raw.TotalIssues,
		TotalPRs:                raw.TotalPRs,
		TotalReviews:            raw.TotalReviews,
		TotalReposCreated:       raw.TotalReposCreated,
		RestrictedContributions: raw.RestrictedContributions,
		CommitsDefaultBranch:    raw.CommitsDefaultBranch,
		TotalAdditions:          raw.TotalAdditions,
		TotalDeletions:          raw.TotalDeletions,
		ReposContributedTo:      len(raw.Repos),
	}

	// Convert repo stats
	for _, r := range raw.Repos {
		p.RepoStats = append(p.RepoStats, profile.RepoContribution{
			FullName:  r.FullName,
			IsPrivate: r.IsPrivate,
			Commits:   r.Commits,
			Additions: r.Additions,
			Deletions: r.Deletions,
		})
	}

	// Convert calendar data
	if raw.Calendar != nil {
		var days []profile.CalendarDay
		for _, w := range raw.Calendar.Weeks {
			for _, d := range w.Days {
				date, err := time.Parse("2006-01-02", d.Date)
				if err != nil {
					continue
				}
				days = append(days, profile.CalendarDay{
					Date:              date,
					Weekday:           date.Weekday(),
					ContributionCount: d.ContributionCount,
					Level:             profile.ContributionLevel(d.Level),
				})
			}
		}
		p.Calendar = profile.NewCalendarFromDays(days)
	}

	// Convert monthly data to Activity
	if len(raw.Monthly) > 0 {
		p.Activity = &profile.ActivityTimeline{
			Username: raw.Username,
			From:     raw.From,
			To:       raw.To,
			Months:   make([]profile.MonthlyActivity, 0, len(raw.Monthly)),
		}
		for _, m := range raw.Monthly {
			commitsByRepo := m.CommitsByRepo
			if commitsByRepo == nil {
				commitsByRepo = make(map[string]int)
			}
			p.Activity.Months = append(p.Activity.Months, profile.MonthlyActivity{
				Year:          m.Year,
				Month:         time.Month(m.Month),
				Commits:       m.Commits,
				Issues:        m.Issues,
				PRs:           m.PRs,
				Reviews:       m.Reviews,
				Releases:      m.Releases,
				Additions:     m.Additions,
				Deletions:     m.Deletions,
				CommitsByRepo: commitsByRepo,
				ReposCreated:  m.ReposCreated,
			})
		}
	}

	return p
}

func formatSummary(p *profile.UserProfile) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Profile: %s ===\n", p.Username))
	sb.WriteString(fmt.Sprintf("Period: %s to %s\n\n", p.From.Format("2006-01-02"), p.To.Format("2006-01-02")))

	sb.WriteString("Contributions (GitHub official):\n")
	sb.WriteString(fmt.Sprintf("  Commits:       %d\n", p.TotalCommits))
	sb.WriteString(fmt.Sprintf("  Pull Requests: %d\n", p.TotalPRs))
	sb.WriteString(fmt.Sprintf("  Issues:        %d\n", p.TotalIssues))
	sb.WriteString(fmt.Sprintf("  Reviews:       %d\n", p.TotalReviews))
	sb.WriteString(fmt.Sprintf("  Repos Created: %d\n", p.TotalReposCreated))
	sb.WriteString("\n")

	sb.WriteString("Code Changes (from default branch history):\n")
	sb.WriteString(fmt.Sprintf("  Commits:   %d\n", p.CommitsDefaultBranch))
	sb.WriteString(fmt.Sprintf("  Additions: +%d\n", p.TotalAdditions))
	sb.WriteString(fmt.Sprintf("  Deletions: -%d\n", p.TotalDeletions))
	sb.WriteString(fmt.Sprintf("  Net:       %+d\n", p.TotalAdditions-p.TotalDeletions))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Repositories Contributed To: %d\n", p.ReposContributedTo))
	sb.WriteString("\n")

	// Calendar stats
	if p.Calendar != nil {
		sb.WriteString("Activity:\n")
		sb.WriteString(fmt.Sprintf("  Days with contributions: %d\n", p.Calendar.DaysWithContributions()))
		sb.WriteString(fmt.Sprintf("  Longest streak:          %d days\n", p.Calendar.LongestStreak()))
		sb.WriteString(fmt.Sprintf("  Current streak:          %d days\n", p.Calendar.CurrentStreak()))
		sb.WriteString("\n")
	}

	// Top repos
	if len(p.RepoStats) > 0 {
		sb.WriteString("Top Repositories by Commits:\n")
		for i, repo := range p.TopReposByCommits(5) {
			sb.WriteString(fmt.Sprintf("  %d. %s: %d commits (+%d/-%d)\n",
				i+1, repo.FullName, repo.Commits, repo.Additions, repo.Deletions))
		}
		sb.WriteString("\n")
	}

	// Monthly activity
	if p.Activity != nil && len(p.Activity.Months) > 0 {
		sb.WriteString("Monthly Activity (from default branch history):\n")
		for _, m := range p.Activity.Months {
			if m.Commits > 0 || m.PRs > 0 || m.Issues > 0 {
				sb.WriteString(fmt.Sprintf("  %s %d:\n", m.MonthName(), m.Year))
				if m.Commits > 0 {
					sb.WriteString(fmt.Sprintf("    - %s\n", m.CommitSummary()))
				}
				if m.PRs > 0 {
					sb.WriteString(fmt.Sprintf("    - %s\n", m.PRSummary()))
				}
				if m.Issues > 0 {
					sb.WriteString(fmt.Sprintf("    - %s\n", m.IssueSummary()))
				}
			}
		}
	}

	return sb.String()
}

// parseCommaSeparated splits a comma-separated string into a slice.
// Returns nil for empty input.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// parseVisibility converts a visibility string to graphql.Visibility.
func parseVisibility(s string) (graphql.Visibility, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "all", "":
		return graphql.VisibilityAll, nil
	case "public":
		return graphql.VisibilityPublic, nil
	case "private":
		return graphql.VisibilityPrivate, nil
	default:
		return graphql.VisibilityAll, fmt.Errorf("invalid visibility %q: use 'all', 'public', or 'private'", s)
	}
}
