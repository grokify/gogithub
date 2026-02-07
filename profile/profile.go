package profile

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/gogithub/graphql"
	"github.com/grokify/gogithub/release"
	"github.com/shurcooL/githubv4"
)

// UserProfile contains comprehensive GitHub contribution statistics for a user.
// This aggregates data from multiple GitHub API endpoints to provide a complete
// picture of a user's contributions, similar to what's shown on their profile page.
type UserProfile struct {
	Username string
	From     time.Time
	To       time.Time

	// Summary counts (from GitHub's contribution collection)
	// TotalCommits is the official GitHub count shown on the profile page.
	TotalCommits      int
	TotalIssues       int
	TotalPRs          int
	TotalReviews      int
	TotalReposCreated int

	// Private contributions (requires authentication as the user)
	RestrictedContributions int

	// Detailed commit stats from traversing default branch history.
	// CommitsDefaultBranch may differ from TotalCommits because it only
	// counts commits on default branches, missing feature branches,
	// squash-merged commits, and inaccessible repositories.
	CommitsDefaultBranch int
	TotalAdditions       int
	TotalDeletions       int

	// Repository data
	ReposContributedTo int
	RepoStats          []RepoContribution

	// Time-series data
	Calendar *ContributionCalendar
	Activity *ActivityTimeline
}

// RepoContribution contains contribution statistics for a single repository.
type RepoContribution struct {
	Owner     string
	Name      string
	FullName  string // "owner/repo"
	IsPrivate bool
	Commits   int
	Additions int
	Deletions int
	Releases  int // Number of releases (optional, may be 0 if not fetched)
}

// ProgressInfo contains information about the current progress state.
type ProgressInfo struct {
	Stage       int    // Current stage number (1-based)
	TotalStages int    // Total number of stages
	Description string // What's happening in this stage
	Current     int    // Current item within stage (0 if not applicable)
	Total       int    // Total items in stage (0 if not applicable)
	Done        bool   // True if this stage is complete
}

// ProgressFunc is called to report progress during profile fetching.
type ProgressFunc func(info ProgressInfo)

// Options configures what data to fetch for a user profile.
type Options struct {
	// Visibility filters which repositories to include.
	// Default is VisibilityAll.
	Visibility graphql.Visibility

	// IncludeReleases fetches release counts for each contributed repository.
	// This requires additional API calls and may be slow for users with many repos.
	IncludeReleases bool

	// MaxReleaseFetchRepos limits how many repos to fetch releases for.
	// 0 means no limit. Only used if IncludeReleases is true.
	MaxReleaseFetchRepos int

	// Progress is called to report progress during fetching.
	// If nil, no progress is reported.
	Progress ProgressFunc
}

// DefaultOptions returns sensible default options.
func DefaultOptions() *Options {
	return &Options{
		Visibility:           graphql.VisibilityAll,
		IncludeReleases:      false,
		MaxReleaseFetchRepos: 0,
	}
}

// GetUserProfile fetches comprehensive profile statistics for a GitHub user.
// This function makes multiple API calls to gather all data:
//   - GraphQL: contributionsCollection for summary stats and calendar
//   - GraphQL: commit history for additions/deletions per repo
//   - REST: release counts (optional)
func GetUserProfile(ctx context.Context, restClient *github.Client, gqlClient *githubv4.Client, username string, from, to time.Time, opts *Options) (*UserProfile, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Determine total stages
	totalStages := 4 // contrib stats, commit details, process repos, build timeline
	if opts.IncludeReleases && restClient != nil {
		totalStages = 5 // add releases stage
	}

	progress := opts.Progress
	if progress == nil {
		progress = func(ProgressInfo) {} // no-op
	}

	report := func(stage int, desc string, current, total int, done bool) {
		progress(ProgressInfo{
			Stage:       stage,
			TotalStages: totalStages,
			Description: desc,
			Current:     current,
			Total:       total,
			Done:        done,
		})
	}

	profile := &UserProfile{
		Username: username,
		From:     from,
		To:       to,
	}

	// Stage 1: Fetch contribution stats (commits, issues, PRs, reviews, repos created)
	report(1, "Fetching contribution statistics", 0, 0, false)
	contribStats, err := graphql.GetContributionStatsMultiYear(ctx, gqlClient, username, from, to)
	if err != nil {
		return nil, fmt.Errorf("get contribution stats: %w", err)
	}

	profile.TotalCommits = contribStats.TotalCommitContributions
	profile.TotalIssues = contribStats.TotalIssueContributions
	profile.TotalPRs = contribStats.TotalPRContributions
	profile.TotalReviews = contribStats.TotalPRReviewContributions
	profile.TotalReposCreated = contribStats.TotalRepositoryContributions
	profile.RestrictedContributions = contribStats.RestrictedContributions

	// Build contribution calendar from the GraphQL data
	profile.Calendar = buildCalendarFromContributions(contribStats.ContributionsByMonth, from, to)
	report(1, "Fetching contribution statistics", 0, 0, true)

	// Stage 2: Fetch detailed commit stats (additions/deletions per repo)
	report(2, "Fetching commit details", 0, 0, false)
	commitStatsProgress := func(current, total int) {
		report(2, "Fetching commit details", current, total, false)
	}
	commitStats, err := graphql.GetCommitStatsWithProgress(ctx, gqlClient, username, from, to, opts.Visibility, commitStatsProgress)
	if err != nil {
		return nil, fmt.Errorf("get commit stats: %w", err)
	}

	profile.CommitsDefaultBranch = commitStats.TotalCommits
	profile.TotalAdditions = commitStats.Additions
	profile.TotalDeletions = commitStats.Deletions
	profile.ReposContributedTo = len(commitStats.ByRepo)
	report(2, "Fetching commit details", 0, 0, true)

	// Stage 3: Process repositories
	repoCount := len(commitStats.ByRepo)
	report(3, "Processing repositories", 0, repoCount, false)

	profile.RepoStats = make([]RepoContribution, 0, repoCount)
	for i, repo := range commitStats.ByRepo {
		profile.RepoStats = append(profile.RepoStats, RepoContribution{
			Owner:     repo.Owner,
			Name:      repo.Name,
			FullName:  fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			IsPrivate: repo.IsPrivate,
			Commits:   repo.Commits,
			Additions: repo.Additions,
			Deletions: repo.Deletions,
		})
		if (i+1)%10 == 0 || i+1 == repoCount {
			report(3, "Processing repositories", i+1, repoCount, false)
		}
	}
	report(3, "Processing repositories", repoCount, repoCount, true)

	// Stage 4: Build activity timeline
	report(4, "Building activity timeline", 0, 0, false)
	profile.Activity = buildActivityTimeline(username, from, to, contribStats, commitStats)
	report(4, "Building activity timeline", 0, 0, true)

	// Stage 5 (optional): Fetch release counts
	if opts.IncludeReleases && restClient != nil {
		fetchReleaseCounts(ctx, restClient, profile, opts.MaxReleaseFetchRepos, progress, totalStages)
	}

	return profile, nil
}

// buildCalendarFromContributions creates a ContributionCalendar from monthly contribution data.
// Note: This provides monthly granularity. For daily granularity, use the GraphQL
// contributionCalendar field directly.
func buildCalendarFromContributions(monthly []graphql.MonthlyContribution, from, to time.Time) *ContributionCalendar {
	// Create a day for the first of each month with that month's contribution count
	var days []CalendarDay

	for _, m := range monthly {
		date := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
		if date.Before(from) || date.After(to) {
			continue
		}
		days = append(days, CalendarDay{
			Date:              date,
			Weekday:           date.Weekday(),
			ContributionCount: m.Count,
			Level:             CalculateLevel(m.Count),
		})
	}

	return NewCalendarFromDays(days)
}

// buildActivityTimeline creates an ActivityTimeline from contribution and commit stats.
func buildActivityTimeline(username string, from, to time.Time, contribStats *graphql.ContributionStats, commitStats *graphql.CommitStats) *ActivityTimeline {
	timeline := &ActivityTimeline{
		Username: username,
		From:     from,
		To:       to,
		Months:   []MonthlyActivity{},
	}

	// Create a map of year-month to activity
	activityMap := make(map[string]*MonthlyActivity)

	// Add monthly contributions from contribution stats
	for _, mc := range contribStats.ContributionsByMonth {
		key := mc.YearMonth()
		if _, exists := activityMap[key]; !exists {
			activityMap[key] = &MonthlyActivity{
				Year:          mc.Year,
				Month:         mc.Month,
				CommitsByRepo: make(map[string]int),
			}
		}
		// Note: ContributionsByMonth.Count includes all contribution types
		// We'll get more specific breakdowns from commit stats
	}

	// Add monthly commit stats with additions/deletions
	for _, mcs := range commitStats.ByMonth {
		key := mcs.YearMonth()
		activity, exists := activityMap[key]
		if !exists {
			activity = &MonthlyActivity{
				Year:          mcs.Year,
				Month:         mcs.Month,
				CommitsByRepo: make(map[string]int),
			}
			activityMap[key] = activity
		}
		activity.Commits = mcs.Commits
		activity.Additions = mcs.Additions
		activity.Deletions = mcs.Deletions
	}

	// Add per-repo commit breakdowns
	// Note: commitStats.ByRepo doesn't have monthly breakdown, so we distribute
	// proportionally or leave CommitsByRepo empty for now
	// A more accurate implementation would need per-month per-repo data from GraphQL
	for _, repo := range commitStats.ByRepo {
		fullName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
		// Add to all months proportionally (simplified approach)
		// In reality, you'd want per-month per-repo data
		for _, activity := range activityMap {
			if activity.Commits > 0 && repo.Commits > 0 {
				// Estimate repo's contribution to this month
				ratio := float64(activity.Commits) / float64(commitStats.TotalCommits)
				estimated := int(float64(repo.Commits) * ratio)
				if estimated > 0 {
					activity.CommitsByRepo[fullName] += estimated
				}
			}
		}
	}

	// Convert map to sorted slice
	for _, activity := range activityMap {
		timeline.Months = append(timeline.Months, *activity)
	}

	timeline.SortByDateDesc() // Most recent first

	return timeline
}

// fetchReleaseCounts fetches release counts for repositories and aggregates by month.
// Errors fetching individual repos are silently ignored (e.g., lost access, deleted repo).
func fetchReleaseCounts(ctx context.Context, restClient *github.Client, profile *UserProfile, maxRepos int, progress ProgressFunc, totalStages int) {
	total := len(profile.RepoStats)
	if maxRepos > 0 && maxRepos < total {
		total = maxRepos
	}

	report := func(current int, done bool) {
		progress(ProgressInfo{
			Stage:       5,
			TotalStages: totalStages,
			Description: "Fetching release counts",
			Current:     current,
			Total:       total,
			Done:        done,
		})
	}

	report(0, false)

	// Track releases by month for aggregation
	releasesByMonth := make(map[string]int) // "2024-01" -> count

	count := 0
	for i := range profile.RepoStats {
		if maxRepos > 0 && count >= maxRepos {
			break
		}

		repo := &profile.RepoStats[i]
		releases, err := release.ListReleases(ctx, restClient, repo.Owner, repo.Name)
		if err != nil {
			// Skip repos we can't access (might have lost access or repo deleted)
			continue
		}

		// Count releases and aggregate by month
		repoReleaseCount := 0
		for _, rel := range releases {
			pubAt := rel.GetPublishedAt()
			if pubAt.IsZero() {
				// Fall back to created_at if not published
				pubAt = rel.GetCreatedAt()
			}
			if pubAt.IsZero() {
				repoReleaseCount++
				continue
			}

			// Check if release is within profile date range
			relTime := pubAt.Time
			if relTime.Before(profile.From) || relTime.After(profile.To) {
				continue
			}

			repoReleaseCount++
			key := relTime.Format("2006-01")
			releasesByMonth[key]++
		}
		repo.Releases = repoReleaseCount
		count++
		report(count, false)
	}

	// Update monthly activity with release counts
	if profile.Activity != nil {
		for i := range profile.Activity.Months {
			month := &profile.Activity.Months[i]
			key := time.Date(month.Year, month.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
			month.Releases = releasesByMonth[key]
		}
	}

	report(count, true)
}

// Summary returns a brief text summary of the profile.
func (p *UserProfile) Summary() string {
	return fmt.Sprintf("%s: %d commits (+%d/-%d) in %d repos, %d PRs, %d issues, %d reviews",
		p.Username,
		p.TotalCommits,
		p.TotalAdditions,
		p.TotalDeletions,
		p.ReposContributedTo,
		p.TotalPRs,
		p.TotalIssues,
		p.TotalReviews,
	)
}

// TopReposByCommits returns the top N repositories by commit count.
func (p *UserProfile) TopReposByCommits(n int) []RepoContribution {
	repos := make([]RepoContribution, len(p.RepoStats))
	copy(repos, p.RepoStats)

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Commits > repos[j].Commits
	})

	if n > 0 && len(repos) > n {
		repos = repos[:n]
	}
	return repos
}

// TopReposByAdditions returns the top N repositories by lines added.
func (p *UserProfile) TopReposByAdditions(n int) []RepoContribution {
	repos := make([]RepoContribution, len(p.RepoStats))
	copy(repos, p.RepoStats)

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Additions > repos[j].Additions
	})

	if n > 0 && len(repos) > n {
		repos = repos[:n]
	}
	return repos
}

// PublicRepos returns only public repository contributions.
func (p *UserProfile) PublicRepos() []RepoContribution {
	var public []RepoContribution
	for _, repo := range p.RepoStats {
		if !repo.IsPrivate {
			public = append(public, repo)
		}
	}
	return public
}

// PrivateRepos returns only private repository contributions.
func (p *UserProfile) PrivateRepos() []RepoContribution {
	var private []RepoContribution
	for _, repo := range p.RepoStats {
		if repo.IsPrivate {
			private = append(private, repo)
		}
	}
	return private
}
