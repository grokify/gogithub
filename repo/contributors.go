package repo

import (
	"context"
	"time"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ListContributorStats returns contributor statistics for a repository.
// This wraps the GitHub REST API endpoint: GET /repos/{owner}/{repo}/stats/contributors
//
// Note: GitHub may return 202 Accepted while computing statistics. The clientv1
// implementation handles retries internally.
func ListContributorStats(ctx context.Context, client clientv1.Client, owner, repo string) ([]*gogithub.ContributorStats, error) {
	return client.GetContributorStats(ctx, owner, repo)
}

// GetContributorStats returns statistics for a specific contributor in a repository.
// Returns nil if the user has not contributed to the repository.
func GetContributorStats(ctx context.Context, client clientv1.Client, owner, repo, username string) (*gogithub.ContributorStats, error) {
	allStats, err := client.GetContributorStats(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	for _, stats := range allStats {
		if stats.Author != nil && stats.Author.Login == username {
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
func GetContributorSummary(ctx context.Context, client clientv1.Client, owner, repo, username string) (*ContributorSummary, error) {
	stats, err := GetContributorStats(ctx, client, owner, repo, username)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return nil, nil
	}

	summary := &ContributorSummary{
		Username:     username,
		TotalCommits: stats.Total,
	}

	for _, week := range stats.Weeks {
		summary.TotalAdditions += week.Additions
		summary.TotalDeletions += week.Deletions

		if week.Commits > 0 {
			weekTime := week.Week
			if !weekTime.IsZero() {
				if summary.FirstCommit.IsZero() || weekTime.Before(summary.FirstCommit) {
					summary.FirstCommit = weekTime
				}
				if weekTime.After(summary.LastCommit) {
					summary.LastCommit = weekTime
				}
			}
		}
	}

	return summary, nil
}
