package profile

import (
	"fmt"
	"sort"
	"time"
)

// MonthlyActivity represents contribution activity for a single month.
// This mirrors the activity feed format shown on GitHub user profiles.
type MonthlyActivity struct {
	Year  int
	Month time.Month

	// Contribution counts
	Commits  int
	Issues   int
	PRs      int
	Reviews  int
	Releases int // Releases published in contributed repos this month

	// Commit details
	Additions int
	Deletions int

	// Repository breakdown for commits
	CommitsByRepo map[string]int // "owner/repo" -> commit count

	// Repos where user opened issues/PRs this month
	IssueRepos []string
	PRRepos    []string

	// New repos created this month
	ReposCreated []string
}

// YearMonth returns a formatted string like "2024-01".
func (m *MonthlyActivity) YearMonth() string {
	return time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

// MonthName returns the full month name (e.g., "January").
func (m *MonthlyActivity) MonthName() string {
	return m.Month.String()
}

// TotalContributions returns the sum of all contribution types.
func (m *MonthlyActivity) TotalContributions() int {
	return m.Commits + m.Issues + m.PRs + m.Reviews
}

// CommitRepoCount returns the number of distinct repos with commits.
func (m *MonthlyActivity) CommitRepoCount() int {
	return len(m.CommitsByRepo)
}

// CommitSummary returns a GitHub-style summary string.
// Example: "Created 42 commits in 5 repositories"
func (m *MonthlyActivity) CommitSummary() string {
	if m.Commits == 0 {
		return ""
	}
	repos := m.CommitRepoCount()
	if repos == 1 {
		return fmt.Sprintf("Created %d commits in 1 repository", m.Commits)
	}
	return fmt.Sprintf("Created %d commits in %d repositories", m.Commits, repos)
}

// PRSummary returns a GitHub-style summary for PRs.
// Example: "Opened 3 pull requests in 2 repositories"
func (m *MonthlyActivity) PRSummary() string {
	if m.PRs == 0 {
		return ""
	}
	repos := len(m.PRRepos)
	if repos == 0 {
		repos = 1 // Fallback if repo info not available
	}
	if repos == 1 {
		return fmt.Sprintf("Opened %d pull requests in 1 repository", m.PRs)
	}
	return fmt.Sprintf("Opened %d pull requests in %d repositories", m.PRs, repos)
}

// IssueSummary returns a GitHub-style summary for issues.
// Example: "Opened 5 issues in 3 repositories"
func (m *MonthlyActivity) IssueSummary() string {
	if m.Issues == 0 {
		return ""
	}
	repos := len(m.IssueRepos)
	if repos == 0 {
		repos = 1
	}
	if repos == 1 {
		return fmt.Sprintf("Opened %d issues in 1 repository", m.Issues)
	}
	return fmt.Sprintf("Opened %d issues in %d repositories", m.Issues, repos)
}

// ReviewSummary returns a summary for PR reviews.
func (m *MonthlyActivity) ReviewSummary() string {
	if m.Reviews == 0 {
		return ""
	}
	if m.Reviews == 1 {
		return "Reviewed 1 pull request"
	}
	return fmt.Sprintf("Reviewed %d pull requests", m.Reviews)
}

// RepoCreatedSummary returns a summary for created repos.
func (m *MonthlyActivity) RepoCreatedSummary() string {
	count := len(m.ReposCreated)
	if count == 0 {
		return ""
	}
	if count == 1 {
		return fmt.Sprintf("Created 1 repository: %s", m.ReposCreated[0])
	}
	return fmt.Sprintf("Created %d repositories", count)
}

// TopCommitRepos returns the top N repositories by commit count.
func (m *MonthlyActivity) TopCommitRepos(n int) []RepoCommitCount {
	var repos []RepoCommitCount
	for repo, count := range m.CommitsByRepo {
		repos = append(repos, RepoCommitCount{Repo: repo, Commits: count})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Commits > repos[j].Commits
	})

	if n > 0 && len(repos) > n {
		repos = repos[:n]
	}
	return repos
}

// RepoCommitCount pairs a repository name with its commit count.
type RepoCommitCount struct {
	Repo    string
	Commits int
}

// ActivityTimeline represents a chronological list of monthly activity.
type ActivityTimeline struct {
	Username string
	From     time.Time
	To       time.Time
	Months   []MonthlyActivity
}

// TotalCommits returns the sum of commits across all months.
func (t *ActivityTimeline) TotalCommits() int {
	total := 0
	for _, m := range t.Months {
		total += m.Commits
	}
	return total
}

// TotalContributions returns the sum of all contribution types across all months.
func (t *ActivityTimeline) TotalContributions() int {
	total := 0
	for _, m := range t.Months {
		total += m.TotalContributions()
	}
	return total
}

// MostActiveMonth returns the month with the most total contributions.
func (t *ActivityTimeline) MostActiveMonth() *MonthlyActivity {
	if len(t.Months) == 0 {
		return nil
	}

	maxIdx := 0
	maxContrib := t.Months[0].TotalContributions()

	for i, m := range t.Months[1:] {
		if contrib := m.TotalContributions(); contrib > maxContrib {
			maxContrib = contrib
			maxIdx = i + 1
		}
	}

	return &t.Months[maxIdx]
}

// GetMonth returns activity for a specific year/month.
// Returns nil if not found.
func (t *ActivityTimeline) GetMonth(year int, month time.Month) *MonthlyActivity {
	for i := range t.Months {
		if t.Months[i].Year == year && t.Months[i].Month == month {
			return &t.Months[i]
		}
	}
	return nil
}

// MonthsWithActivity returns the count of months that have any contributions.
func (t *ActivityTimeline) MonthsWithActivity() int {
	count := 0
	for _, m := range t.Months {
		if m.TotalContributions() > 0 {
			count++
		}
	}
	return count
}

// AverageMonthlyContributions returns the average contributions per month.
func (t *ActivityTimeline) AverageMonthlyContributions() float64 {
	if len(t.Months) == 0 {
		return 0
	}
	return float64(t.TotalContributions()) / float64(len(t.Months))
}

// SortByDate sorts months chronologically (oldest first).
func (t *ActivityTimeline) SortByDate() {
	sort.Slice(t.Months, func(i, j int) bool {
		if t.Months[i].Year != t.Months[j].Year {
			return t.Months[i].Year < t.Months[j].Year
		}
		return t.Months[i].Month < t.Months[j].Month
	})
}

// SortByDateDesc sorts months reverse chronologically (newest first).
func (t *ActivityTimeline) SortByDateDesc() {
	sort.Slice(t.Months, func(i, j int) bool {
		if t.Months[i].Year != t.Months[j].Year {
			return t.Months[i].Year > t.Months[j].Year
		}
		return t.Months[i].Month > t.Months[j].Month
	})
}
