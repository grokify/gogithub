package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v84/github"
)

// ListContributorStats returns contributor statistics for a repository.
// This wraps the GitHub REST API endpoint: GET /repos/{owner}/{repo}/stats/contributors
//
// Note: GitHub may return 202 Accepted while computing statistics. This function
// automatically retries with exponential backoff until stats are available or
// the context is cancelled.
func ListContributorStats(ctx context.Context, gh *github.Client, owner, repo string) ([]*github.ContributorStats, error) {
	return listContributorStatsWithRetry(ctx, gh, owner, repo, 5, 2*time.Second)
}

// listContributorStatsWithRetry handles 202 Accepted responses by retrying.
func listContributorStatsWithRetry(ctx context.Context, gh *github.Client, owner, repo string, maxRetries int, initialBackoff time.Duration) ([]*github.ContributorStats, error) {
	backoff := initialBackoff

	for attempt := range maxRetries {
		stats, resp, err := gh.Repositories.ListContributorsStats(ctx, owner, repo)

		// Check for 202 Accepted (stats being computed)
		var acceptedErr *github.AcceptedError
		if errors.As(err, &acceptedErr) {
			if attempt < maxRetries-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					backoff *= 2 // exponential backoff
					continue
				}
			}
			return nil, fmt.Errorf("contributor stats still computing after %d retries", maxRetries)
		}

		if err != nil {
			return nil, fmt.Errorf("list contributor stats: %w", err)
		}

		// Check for empty response (can happen if repo has no commits)
		if resp.StatusCode == 204 || len(stats) == 0 {
			return []*github.ContributorStats{}, nil
		}

		return stats, nil
	}

	return nil, fmt.Errorf("failed to get contributor stats after %d attempts", maxRetries)
}

// GetContributorStats returns statistics for a specific contributor in a repository.
// Returns nil if the user has not contributed to the repository.
func GetContributorStats(ctx context.Context, gh *github.Client, owner, repo, username string) (*github.ContributorStats, error) {
	allStats, err := ListContributorStats(ctx, gh, owner, repo)
	if err != nil {
		return nil, err
	}

	for _, stats := range allStats {
		if stats.Author != nil && stats.Author.GetLogin() == username {
			return stats, nil
		}
	}

	return nil, nil // User not found among contributors
}

// ContributorSummary provides a simplified view of contributor statistics.
type ContributorSummary struct {
	Username       string
	TotalCommits   int
	TotalAdditions int
	TotalDeletions int
	FirstCommit    time.Time
	LastCommit     time.Time
}

// GetContributorSummary returns a summarized view of a contributor's statistics.
func GetContributorSummary(ctx context.Context, gh *github.Client, owner, repo, username string) (*ContributorSummary, error) {
	stats, err := GetContributorStats(ctx, gh, owner, repo, username)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return nil, nil
	}

	summary := &ContributorSummary{
		Username:     username,
		TotalCommits: stats.GetTotal(),
	}

	for _, week := range stats.Weeks {
		summary.TotalAdditions += week.GetAdditions()
		summary.TotalDeletions += week.GetDeletions()

		if week.GetCommits() > 0 {
			weekTime := week.Week.GetTime()
			if weekTime != nil {
				if summary.FirstCommit.IsZero() || weekTime.Before(summary.FirstCommit) {
					summary.FirstCommit = *weekTime
				}
				if weekTime.After(summary.LastCommit) {
					summary.LastCommit = *weekTime
				}
			}
		}
	}

	return summary, nil
}
