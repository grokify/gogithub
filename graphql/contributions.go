package graphql

import (
	"context"
	"sort"
	"time"

	"github.com/shurcooL/githubv4"
)

// ContributionStats represents aggregated contribution statistics from the
// contributionsCollection query. This provides quick stats similar to what's
// shown on a user's GitHub profile page.
//
// Note: RestrictedContributions includes all private contribution types
// (commits, issues, PRs, reviews), not just commits.
type ContributionStats struct {
	Username                     string
	From                         time.Time
	To                           time.Time
	TotalCommitContributions     int
	TotalIssueContributions      int
	TotalPRContributions         int
	TotalPRReviewContributions   int
	TotalRepositoryContributions int
	RestrictedContributions      int // All private contributions (not just commits)
	ContributionsByMonth         []MonthlyContribution
}

// MonthlyContribution represents contribution counts for a specific month.
type MonthlyContribution struct {
	Year  int
	Month time.Month
	Count int
}

// YearMonth returns a formatted string like "2024-01".
func (mc MonthlyContribution) YearMonth() string {
	return time.Date(mc.Year, mc.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

// contributionsQuery is the GraphQL query structure for fetching contribution statistics.
type contributionsQuery struct {
	User struct {
		ContributionsCollection struct {
			TotalCommitContributions            githubv4.Int
			TotalIssueContributions             githubv4.Int
			TotalPullRequestContributions       githubv4.Int
			TotalPullRequestReviewContributions githubv4.Int
			TotalRepositoryContributions        githubv4.Int
			RestrictedContributionsCount        githubv4.Int
			ContributionCalendar                struct {
				TotalContributions githubv4.Int
				Weeks              []struct {
					ContributionDays []struct {
						ContributionCount githubv4.Int
						Date              string
					}
				}
			}
		} `graphql:"contributionsCollection(from: $from, to: $to)"`
	} `graphql:"user(login: $login)"`
}

// GetContributionStats retrieves contribution statistics for a user within a date range.
// The GitHub API limits queries to a maximum of 1 year at a time.
func GetContributionStats(ctx context.Context, client *githubv4.Client, username string, from, to time.Time) (*ContributionStats, error) {
	var query contributionsQuery
	variables := map[string]any{
		"login": githubv4.String(username),
		"from":  githubv4.DateTime{Time: from},
		"to":    githubv4.DateTime{Time: to},
	}

	if err := client.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	cc := query.User.ContributionsCollection
	stats := &ContributionStats{
		Username:                     username,
		From:                         from,
		To:                           to,
		TotalCommitContributions:     int(cc.TotalCommitContributions),
		TotalIssueContributions:      int(cc.TotalIssueContributions),
		TotalPRContributions:         int(cc.TotalPullRequestContributions),
		TotalPRReviewContributions:   int(cc.TotalPullRequestReviewContributions),
		TotalRepositoryContributions: int(cc.TotalRepositoryContributions),
		RestrictedContributions:      int(cc.RestrictedContributionsCount),
		ContributionsByMonth:         aggregateByMonth(cc.ContributionCalendar.Weeks),
	}

	return stats, nil
}

// GetContributionStatsMultiYear retrieves contribution statistics spanning multiple years.
// It automatically handles the 1-year limit by making multiple queries.
func GetContributionStatsMultiYear(ctx context.Context, client *githubv4.Client, username string, from, to time.Time) (*ContributionStats, error) {
	if to.Sub(from) <= 365*24*time.Hour {
		return GetContributionStats(ctx, client, username, from, to)
	}

	// Split into yearly chunks
	stats := &ContributionStats{
		Username:             username,
		From:                 from,
		To:                   to,
		ContributionsByMonth: []MonthlyContribution{},
	}

	monthlyMap := make(map[string]int) // "2024-01" -> count

	current := from
	for current.Before(to) {
		end := current.AddDate(1, 0, 0)
		if end.After(to) {
			end = to
		}

		yearStats, err := GetContributionStats(ctx, client, username, current, end)
		if err != nil {
			return nil, err
		}

		stats.TotalCommitContributions += yearStats.TotalCommitContributions
		stats.TotalIssueContributions += yearStats.TotalIssueContributions
		stats.TotalPRContributions += yearStats.TotalPRContributions
		stats.TotalPRReviewContributions += yearStats.TotalPRReviewContributions
		stats.TotalRepositoryContributions += yearStats.TotalRepositoryContributions
		stats.RestrictedContributions += yearStats.RestrictedContributions

		for _, mc := range yearStats.ContributionsByMonth {
			monthlyMap[mc.YearMonth()] += mc.Count
		}

		current = end
	}

	// Convert map to sorted slice
	stats.ContributionsByMonth = mapToMonthlyContributions(monthlyMap)

	return stats, nil
}

// aggregateByMonth converts daily contribution data to monthly totals.
func aggregateByMonth(weeks []struct {
	ContributionDays []struct {
		ContributionCount githubv4.Int
		Date              string
	}
}) []MonthlyContribution {
	monthlyMap := make(map[string]int)

	for _, week := range weeks {
		for _, day := range week.ContributionDays {
			t, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				continue
			}
			key := t.Format("2006-01")
			monthlyMap[key] += int(day.ContributionCount)
		}
	}

	return mapToMonthlyContributions(monthlyMap)
}

// mapToMonthlyContributions converts a map of "YYYY-MM" -> count to a sorted slice.
func mapToMonthlyContributions(m map[string]int) []MonthlyContribution {
	result := make([]MonthlyContribution, 0, len(m))

	for ym, count := range m {
		t, err := time.Parse("2006-01", ym)
		if err != nil {
			continue
		}
		result = append(result, MonthlyContribution{
			Year:  t.Year(),
			Month: t.Month(),
			Count: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Year != result[j].Year {
			return result[i].Year < result[j].Year
		}
		return result[i].Month < result[j].Month
	})

	return result
}
