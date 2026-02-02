package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/gogithub/graphql"
	"github.com/grokify/gogithub/profile"
	"github.com/spf13/cobra"
)

var (
	profileUser            string
	profileFrom            string
	profileTo              string
	profileFormat          string
	profileOutput          string
	profileOutputRaw       string
	profileOutputAggregate string
	profileInput           string
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

  # Generate aggregate from existing raw file (no API calls)
  gogithub profile --input raw.json --output aggregate.json

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
	profileCmd.Flags().StringVarP(&profileInput, "input", "i", "", "Input raw JSON file (skips API calls)")
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

	// Track last stage to avoid duplicate "done" messages
	lastStage := 0
	lastDone := false

	// Progress callback that writes to stderr with stage information and progress bar
	progressFunc := func(info profile.ProgressInfo) {
		// Clear line and move cursor to beginning for updates
		clearLine := "\r\033[K"

		// Calculate percentage
		var pct int
		if info.Done {
			pct = 100
		} else if info.Current > 0 && info.Total > 0 {
			pct = (info.Current * 100) / info.Total
		} else {
			pct = 0
		}

		// Build progress bar (20 chars wide)
		bar := renderProgressBar(pct, 20)

		if info.Done {
			if lastStage == info.Stage && lastDone {
				return // Already printed done for this stage
			}
			// Print completed stage with full bar
			fmt.Fprintf(os.Stderr, "%s[%d/%d] %-34s %s %3d%%\n",
				clearLine, info.Stage, info.TotalStages, info.Description, bar, pct)
			lastDone = true
		} else {
			lastDone = false
			// Show stage with current progress bar
			fmt.Fprintf(os.Stderr, "%s[%d/%d] %-34s %s %3d%%",
				clearLine, info.Stage, info.TotalStages, info.Description, bar, pct)
		}
		lastStage = info.Stage
	}

	opts := &profile.Options{
		Visibility:      graphql.VisibilityAll,
		IncludeReleases: false,
		Progress:        progressFunc,
	}

	p, err := profile.GetUserProfile(ctx, restClient, gqlClient, profileUser, from, to, opts)
	if err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	// Mode: Generate both raw and aggregate files
	if profileOutputRaw != "" || profileOutputAggregate != "" {
		return outputBothFormats(p)
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

func outputBothFormats(p *profile.UserProfile) error {
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
	GeneratedAt time.Time `json:"generated_at"`

	// Summary counts - GitHub official count from contributionsCollection
	TotalCommits            int `json:"total_commits"`
	TotalIssues             int `json:"total_issues"`
	TotalPRs                int `json:"total_prs"`
	TotalReviews            int `json:"total_reviews"`
	TotalReposCreated       int `json:"total_repos_created"`
	RestrictedContributions int `json:"restricted_contributions,omitempty"`

	// Commits from default branch traversal (may differ from total_commits)
	// This count only includes commits on default branches and may miss
	// feature branches, squash-merged commits, and inaccessible repos.
	CommitsDefaultBranch int `json:"commits_default_branch"`

	// Code stats (from default branch traversal)
	TotalAdditions int `json:"total_additions"`
	TotalDeletions int `json:"total_deletions"`

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
	GeneratedAt time.Time `json:"generated_at"`

	// Summary counts - GitHub official count from contributionsCollection
	TotalCommits            int `json:"total_commits"`
	TotalIssues             int `json:"total_issues"`
	TotalPRs                int `json:"total_prs"`
	TotalReviews            int `json:"total_reviews"`
	TotalReposCreated       int `json:"total_repos_created"`
	RestrictedContributions int `json:"restricted_contributions,omitempty"`

	// Commits from default branch traversal (may differ from total_commits)
	CommitsDefaultBranch int `json:"commits_default_branch"`

	// Code stats (from default branch traversal)
	TotalAdditions int `json:"total_additions"`
	TotalDeletions int `json:"total_deletions"`

	// Repo summary
	ReposContributedTo int `json:"repos_contributed_to"`

	// Calendar stats (computed)
	Calendar *CalendarStatsJSON `json:"calendar,omitempty"`

	// Monthly breakdown (from default branch traversal)
	Monthly []MonthlyJSON `json:"monthly,omitempty"`
}

type CalendarDataJSON struct {
	TotalContributions int            `json:"total_contributions"`
	Weeks              []CalendarWeek `json:"weeks,omitempty"`
}

type CalendarWeek struct {
	StartDate string        `json:"start_date"`
	Days      []CalendarDay `json:"days"`
}

type CalendarDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contribution_count"`
	Level             int    `json:"level"`
}

type CalendarStatsJSON struct {
	TotalContributions    int `json:"total_contributions"`
	DaysWithContributions int `json:"days_with_contributions"`
	LongestStreak         int `json:"longest_streak"`
	CurrentStreak         int `json:"current_streak"`
}

type MonthlyJSON struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	MonthName string `json:"month_name"`
	Commits   int    `json:"commits"`
	Issues    int    `json:"issues"`
	PRs       int    `json:"prs"`
	Reviews   int    `json:"reviews"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type RepoJSON struct {
	FullName  string `json:"full_name"`
	IsPrivate bool   `json:"is_private"`
	Commits   int    `json:"commits"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
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
	if p.Activity != nil {
		for _, m := range p.Activity.Months {
			raw.Monthly = append(raw.Monthly, MonthlyJSON{
				Year:      m.Year,
				Month:     int(m.Month),
				MonthName: m.MonthName(),
				Commits:   m.Commits,
				Issues:    m.Issues,
				PRs:       m.PRs,
				Reviews:   m.Reviews,
				Additions: m.Additions,
				Deletions: m.Deletions,
			})
		}
	}

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
	if p.Activity != nil {
		for _, m := range p.Activity.Months {
			agg.Monthly = append(agg.Monthly, MonthlyJSON{
				Year:      m.Year,
				Month:     int(m.Month),
				MonthName: m.MonthName(),
				Commits:   m.Commits,
				Issues:    m.Issues,
				PRs:       m.PRs,
				Reviews:   m.Reviews,
				Additions: m.Additions,
				Deletions: m.Deletions,
			})
		}
	}

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

// renderProgressBar creates a visual progress bar string.
// width is the number of characters for the bar (excluding brackets).
// Returns a string like "[████████░░░░░░░░░░░░]"
func renderProgressBar(percent, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * width) / 100
	empty := width - filled

	// Use Unicode block characters for a smooth look
	// █ (full block) for filled, ░ (light shade) for empty
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return "[" + bar + "]"
}
