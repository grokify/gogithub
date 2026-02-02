package repo

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-github/v82/github"
)

func getTestToken(t *testing.T) string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}
	return token
}

func TestListContributorStatsIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	// Use a well-known public repo
	stats, err := ListContributorStats(ctx, client, "grokify", "gogithub")
	if err != nil {
		t.Fatalf("ListContributorStats failed: %v", err)
	}

	if len(stats) == 0 {
		t.Log("No contributor stats returned (this may be expected for new repos)")
		return
	}

	t.Logf("Found %d contributors", len(stats))

	// Log top contributors
	for i, stat := range stats {
		if i >= 5 {
			break
		}
		if stat.Author != nil && stat.Total != nil {
			t.Logf("  %s: %d commits", stat.Author.GetLogin(), *stat.Total)
		}
	}
}

func TestGetContributorStatsIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	stats, err := GetContributorStats(ctx, client, "grokify", "gogithub", "grokify")
	if err != nil {
		t.Fatalf("GetContributorStats failed: %v", err)
	}

	if stats == nil {
		t.Log("No stats returned for user (may not be a contributor)")
		return
	}

	t.Logf("Stats for grokify in gogithub:")
	t.Logf("  Total commits: %d", stats.GetTotal())
	t.Logf("  Weeks of data: %d", len(stats.Weeks))

	// Sum up additions/deletions from weeks
	var additions, deletions int
	for _, week := range stats.Weeks {
		additions += week.GetAdditions()
		deletions += week.GetDeletions()
	}
	t.Logf("  Total additions: %d", additions)
	t.Logf("  Total deletions: %d", deletions)
}

func TestGetContributorSummaryIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	summary, err := GetContributorSummary(ctx, client, "grokify", "gogithub", "grokify")
	if err != nil {
		t.Fatalf("GetContributorSummary failed: %v", err)
	}

	if summary == nil {
		t.Log("No summary returned for user")
		return
	}

	t.Logf("Summary for %s:", summary.Username)
	t.Logf("  Total commits: %d", summary.TotalCommits)
	t.Logf("  Total additions: %d", summary.TotalAdditions)
	t.Logf("  Total deletions: %d", summary.TotalDeletions)
	t.Logf("  First commit: %s", summary.FirstCommit.Format("2006-01-02"))
	t.Logf("  Last commit: %s", summary.LastCommit.Format("2006-01-02"))

	// Verify data consistency
	if summary.TotalCommits <= 0 {
		t.Error("Expected positive commit count for active contributor")
	}
}

func TestGetContributorStatsNotFoundIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	// Look for a user who is definitely not a contributor
	stats, err := GetContributorStats(ctx, client, "grokify", "gogithub", "torvalds")
	if err != nil {
		t.Fatalf("GetContributorStats failed: %v", err)
	}

	if stats != nil {
		t.Errorf("Expected nil stats for non-contributor, got %+v", stats)
	}
}

func TestListContributorStatsNonExistentRepoIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	_, err := ListContributorStats(ctx, client, "grokify", "this-repo-does-not-exist-12345")
	if err == nil {
		t.Error("Expected error for non-existent repo, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

func TestListContributorStatsLargeRepoIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	client := github.NewClient(nil).WithAuthToken(token)

	// Test with a larger public repo (go-github itself)
	stats, err := ListContributorStats(ctx, client, "google", "go-github")
	if err != nil {
		t.Fatalf("ListContributorStats failed: %v", err)
	}

	t.Logf("go-github has %d contributors", len(stats))

	if len(stats) < 10 {
		t.Errorf("Expected at least 10 contributors for go-github, got %d", len(stats))
	}

	// Verify the stats have data
	for _, stat := range stats[:min(3, len(stats))] {
		if stat.Author == nil {
			t.Error("Contributor has nil author")
			continue
		}
		if stat.Total == nil || *stat.Total <= 0 {
			t.Errorf("Contributor %s has invalid total commits", stat.Author.GetLogin())
		}
	}
}
