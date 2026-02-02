package profile

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v82/github"
	"github.com/grokify/gogithub/graphql"
)

const testUsername = "grokify"

func getTestToken(t *testing.T) string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}
	return token
}

func TestGetUserProfileIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	// Use a 1-month window to minimize API calls
	to := time.Now()
	from := to.AddDate(0, -1, 0)

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, nil)
	if err != nil {
		t.Fatalf("GetUserProfile failed: %v", err)
	}

	// Verify basic fields
	if profile.Username != testUsername {
		t.Errorf("Username = %q, want %q", profile.Username, testUsername)
	}

	if profile.From.IsZero() {
		t.Error("From time should not be zero")
	}

	if profile.To.IsZero() {
		t.Error("To time should not be zero")
	}

	// Log results for manual verification
	t.Logf("Profile for %s:", profile.Username)
	t.Logf("  Total commits: %d", profile.TotalCommits)
	t.Logf("  Total PRs: %d", profile.TotalPRs)
	t.Logf("  Total issues: %d", profile.TotalIssues)
	t.Logf("  Total reviews: %d", profile.TotalReviews)
	t.Logf("  Repos contributed to: %d", profile.ReposContributedTo)
	t.Logf("  Additions: %d, Deletions: %d", profile.TotalAdditions, profile.TotalDeletions)
}

func TestGetUserProfileWithOptionsIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	to := time.Now()
	from := to.AddDate(0, -1, 0)

	opts := &Options{
		Visibility:      graphql.VisibilityPublic,
		IncludeReleases: false, // Skip releases to speed up test
	}

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, opts)
	if err != nil {
		t.Fatalf("GetUserProfile with options failed: %v", err)
	}

	if profile.Username != testUsername {
		t.Errorf("Username = %q, want %q", profile.Username, testUsername)
	}

	// With VisibilityPublic, all repos should be public
	for _, repo := range profile.RepoStats {
		if repo.IsPrivate {
			t.Errorf("Expected only public repos with VisibilityPublic, got private: %s", repo.FullName)
		}
	}

	t.Logf("Public repos contributed to: %d", len(profile.RepoStats))
}

func TestGetUserProfileCalendarIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	// Use a 3-month window to get meaningful calendar data
	to := time.Now()
	from := to.AddDate(0, -3, 0)

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, nil)
	if err != nil {
		t.Fatalf("GetUserProfile failed: %v", err)
	}

	if profile.Calendar == nil {
		t.Fatal("Calendar should not be nil")
	}

	t.Logf("Calendar stats:")
	t.Logf("  Total contributions: %d", profile.Calendar.TotalContributions)
	t.Logf("  Days with contributions: %d", profile.Calendar.DaysWithContributions())
	t.Logf("  Longest streak: %d days", profile.Calendar.LongestStreak())
	t.Logf("  Current streak: %d days", profile.Calendar.CurrentStreak())
	t.Logf("  Number of weeks: %d", len(profile.Calendar.Weeks))

	// Verify calendar structure
	for _, week := range profile.Calendar.Weeks {
		if week.StartDate.Weekday() != time.Sunday {
			t.Errorf("Week start date %v is not a Sunday", week.StartDate)
		}
	}
}

func TestGetUserProfileActivityIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	// Use a 3-month window
	to := time.Now()
	from := to.AddDate(0, -3, 0)

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, nil)
	if err != nil {
		t.Fatalf("GetUserProfile failed: %v", err)
	}

	if profile.Activity == nil {
		t.Fatal("Activity should not be nil")
	}

	t.Logf("Activity timeline:")
	t.Logf("  Months with data: %d", len(profile.Activity.Months))
	t.Logf("  Total commits: %d", profile.Activity.TotalCommits())
	t.Logf("  Total contributions: %d", profile.Activity.TotalContributions())
	t.Logf("  Months with activity: %d", profile.Activity.MonthsWithActivity())

	if most := profile.Activity.MostActiveMonth(); most != nil {
		t.Logf("  Most active month: %s %d (%d contributions)",
			most.MonthName(), most.Year, most.TotalContributions())
	}

	// Log monthly breakdown
	for _, m := range profile.Activity.Months {
		if m.TotalContributions() > 0 {
			t.Logf("  %s %d:", m.MonthName(), m.Year)
			if s := m.CommitSummary(); s != "" {
				t.Logf("    - %s", s)
			}
			if s := m.PRSummary(); s != "" {
				t.Logf("    - %s", s)
			}
		}
	}
}

func TestGetUserProfileWithReleasesIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	// Use a longer window to find repos with releases
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	opts := &Options{
		Visibility:           graphql.VisibilityPublic,
		IncludeReleases:      true,
		MaxReleaseFetchRepos: 5, // Limit to 5 repos to speed up test
	}

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, opts)
	if err != nil {
		t.Fatalf("GetUserProfile with releases failed: %v", err)
	}

	// Count repos with releases
	reposWithReleases := 0
	totalReleases := 0
	for _, repo := range profile.RepoStats {
		if repo.Releases > 0 {
			reposWithReleases++
			totalReleases += repo.Releases
			t.Logf("  %s: %d releases", repo.FullName, repo.Releases)
		}
	}

	t.Logf("Repos with releases: %d, Total releases: %d", reposWithReleases, totalReleases)
}

func TestGetUserProfileTopReposIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, nil)
	if err != nil {
		t.Fatalf("GetUserProfile failed: %v", err)
	}

	t.Logf("Top 5 repos by commits:")
	for i, repo := range profile.TopReposByCommits(5) {
		t.Logf("  %d. %s: %d commits (+%d/-%d)",
			i+1, repo.FullName, repo.Commits, repo.Additions, repo.Deletions)
	}

	t.Logf("\nTop 5 repos by additions:")
	for i, repo := range profile.TopReposByAdditions(5) {
		t.Logf("  %d. %s: +%d lines",
			i+1, repo.FullName, repo.Additions)
	}
}

func TestGetUserProfileSummaryIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	to := time.Now()
	from := to.AddDate(0, -1, 0)

	profile, err := GetUserProfile(ctx, restClient, gqlClient, testUsername, from, to, nil)
	if err != nil {
		t.Fatalf("GetUserProfile failed: %v", err)
	}

	summary := profile.Summary()
	if summary == "" {
		t.Error("Summary should not be empty")
	}

	t.Logf("Summary: %s", summary)
}

func TestGetUserProfileInvalidUserIntegration(t *testing.T) {
	token := getTestToken(t)
	ctx := context.Background()

	restClient := github.NewClient(nil).WithAuthToken(token)
	gqlClient := graphql.NewClient(ctx, token)

	to := time.Now()
	from := to.AddDate(0, -1, 0)

	// Use a username that's very unlikely to exist
	_, err := GetUserProfile(ctx, restClient, gqlClient, "this-user-definitely-does-not-exist-12345", from, to, nil)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}

	t.Logf("Got expected error: %v", err)
}
