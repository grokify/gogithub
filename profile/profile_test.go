package profile

import (
	"testing"
	"time"

	"github.com/grokify/gogithub/graphql"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Visibility != graphql.VisibilityAll {
		t.Errorf("Visibility = %v, want VisibilityAll", opts.Visibility)
	}

	if opts.IncludeReleases {
		t.Error("IncludeReleases should be false by default")
	}

	if opts.MaxReleaseFetchRepos != 0 {
		t.Errorf("MaxReleaseFetchRepos = %d, want 0", opts.MaxReleaseFetchRepos)
	}
}

func TestUserProfileSummary(t *testing.T) {
	profile := &UserProfile{
		Username:           "testuser",
		TotalCommits:       150,
		TotalAdditions:     10000,
		TotalDeletions:     3000,
		ReposContributedTo: 12,
		TotalPRs:           25,
		TotalIssues:        10,
		TotalReviews:       50,
	}

	summary := profile.Summary()
	expected := "testuser: 150 commits (+10000/-3000) in 12 repos, 25 PRs, 10 issues, 50 reviews"

	if summary != expected {
		t.Errorf("Summary() = %q, want %q", summary, expected)
	}
}

func TestUserProfileTopReposByCommits(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "user/repo1", Commits: 50},
			{FullName: "user/repo2", Commits: 100},
			{FullName: "user/repo3", Commits: 25},
			{FullName: "user/repo4", Commits: 75},
		},
	}

	top := profile.TopReposByCommits(2)
	if len(top) != 2 {
		t.Fatalf("TopReposByCommits(2) returned %d repos, want 2", len(top))
	}

	if top[0].FullName != "user/repo2" || top[0].Commits != 100 {
		t.Errorf("First = %v, want user/repo2 with 100 commits", top[0])
	}

	if top[1].FullName != "user/repo4" || top[1].Commits != 75 {
		t.Errorf("Second = %v, want user/repo4 with 75 commits", top[1])
	}
}

func TestUserProfileTopReposByCommitsAll(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "repo1", Commits: 10},
			{FullName: "repo2", Commits: 20},
		},
	}

	// Request more than available
	top := profile.TopReposByCommits(10)
	if len(top) != 2 {
		t.Errorf("TopReposByCommits(10) returned %d repos, want 2", len(top))
	}

	// Zero means all
	top = profile.TopReposByCommits(0)
	if len(top) != 2 {
		t.Errorf("TopReposByCommits(0) returned %d repos, want 2", len(top))
	}
}

func TestUserProfileTopReposByAdditions(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "repo1", Additions: 5000},
			{FullName: "repo2", Additions: 10000},
			{FullName: "repo3", Additions: 2000},
		},
	}

	top := profile.TopReposByAdditions(2)
	if len(top) != 2 {
		t.Fatalf("TopReposByAdditions(2) returned %d repos, want 2", len(top))
	}

	if top[0].FullName != "repo2" {
		t.Errorf("First = %s, want repo2", top[0].FullName)
	}

	if top[1].FullName != "repo1" {
		t.Errorf("Second = %s, want repo1", top[1].FullName)
	}
}

func TestUserProfilePublicRepos(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "user/public1", IsPrivate: false},
			{FullName: "user/private1", IsPrivate: true},
			{FullName: "user/public2", IsPrivate: false},
			{FullName: "user/private2", IsPrivate: true},
		},
	}

	public := profile.PublicRepos()
	if len(public) != 2 {
		t.Fatalf("PublicRepos() returned %d repos, want 2", len(public))
	}

	for _, repo := range public {
		if repo.IsPrivate {
			t.Errorf("PublicRepos() returned private repo: %s", repo.FullName)
		}
	}
}

func TestUserProfilePrivateRepos(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "user/public1", IsPrivate: false},
			{FullName: "user/private1", IsPrivate: true},
			{FullName: "user/public2", IsPrivate: false},
			{FullName: "user/private2", IsPrivate: true},
		},
	}

	private := profile.PrivateRepos()
	if len(private) != 2 {
		t.Fatalf("PrivateRepos() returned %d repos, want 2", len(private))
	}

	for _, repo := range private {
		if !repo.IsPrivate {
			t.Errorf("PrivateRepos() returned public repo: %s", repo.FullName)
		}
	}
}

func TestUserProfileEmptyRepoStats(t *testing.T) {
	profile := &UserProfile{}

	top := profile.TopReposByCommits(5)
	if len(top) != 0 {
		t.Errorf("TopReposByCommits on empty profile returned %d repos", len(top))
	}

	public := profile.PublicRepos()
	if len(public) != 0 {
		t.Errorf("PublicRepos on empty profile returned %d repos", len(public))
	}

	private := profile.PrivateRepos()
	if len(private) != 0 {
		t.Errorf("PrivateRepos on empty profile returned %d repos", len(private))
	}
}

func TestRepoContribution(t *testing.T) {
	repo := RepoContribution{
		Owner:     "grokify",
		Name:      "gogithub",
		FullName:  "grokify/gogithub",
		IsPrivate: false,
		Commits:   100,
		Additions: 5000,
		Deletions: 1000,
		Releases:  10,
	}

	if repo.FullName != "grokify/gogithub" {
		t.Errorf("FullName = %q, want %q", repo.FullName, "grokify/gogithub")
	}

	if repo.IsPrivate {
		t.Error("IsPrivate should be false")
	}

	// Verify all fields are set correctly
	if repo.Commits != 100 {
		t.Errorf("Commits = %d, want 100", repo.Commits)
	}

	if repo.Additions != 5000 {
		t.Errorf("Additions = %d, want 5000", repo.Additions)
	}

	if repo.Deletions != 1000 {
		t.Errorf("Deletions = %d, want 1000", repo.Deletions)
	}

	if repo.Releases != 10 {
		t.Errorf("Releases = %d, want 10", repo.Releases)
	}
}

func TestBuildCalendarFromContributions(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

	monthly := []graphql.MonthlyContribution{
		{Year: 2024, Month: time.January, Count: 50},
		{Year: 2024, Month: time.February, Count: 30},
		{Year: 2024, Month: time.March, Count: 40},
		{Year: 2023, Month: time.December, Count: 100}, // Out of range, should be excluded
	}

	cal := buildCalendarFromContributions(monthly, from, to)

	if cal.TotalContributions != 120 { // 50 + 30 + 40
		t.Errorf("TotalContributions = %d, want 120", cal.TotalContributions)
	}
}

func TestBuildCalendarFromContributionsEmpty(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

	cal := buildCalendarFromContributions(nil, from, to)

	if cal.TotalContributions != 0 {
		t.Errorf("TotalContributions = %d, want 0", cal.TotalContributions)
	}

	if len(cal.Weeks) != 0 {
		t.Errorf("len(Weeks) = %d, want 0", len(cal.Weeks))
	}
}

func TestOptionsWithVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility graphql.Visibility
	}{
		{"all", graphql.VisibilityAll},
		{"public", graphql.VisibilityPublic},
		{"private", graphql.VisibilityPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Visibility: tt.visibility}
			if opts.Visibility != tt.visibility {
				t.Errorf("Visibility = %v, want %v", opts.Visibility, tt.visibility)
			}
		})
	}
}

func TestUserProfileFields(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	profile := &UserProfile{
		Username:                "testuser",
		From:                    from,
		To:                      to,
		TotalCommits:            200,
		TotalIssues:             15,
		TotalPRs:                30,
		TotalReviews:            50,
		TotalReposCreated:       5,
		RestrictedContributions: 20,
		TotalAdditions:          15000,
		TotalDeletions:          5000,
		ReposContributedTo:      10,
	}

	if profile.Username != "testuser" {
		t.Errorf("Username = %q, want %q", profile.Username, "testuser")
	}

	if !profile.From.Equal(from) {
		t.Errorf("From = %v, want %v", profile.From, from)
	}

	if !profile.To.Equal(to) {
		t.Errorf("To = %v, want %v", profile.To, to)
	}

	if profile.TotalCommits != 200 {
		t.Errorf("TotalCommits = %d, want 200", profile.TotalCommits)
	}

	if profile.RestrictedContributions != 20 {
		t.Errorf("RestrictedContributions = %d, want 20", profile.RestrictedContributions)
	}
}

func TestTopReposByCommitsDoesNotMutate(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "repo1", Commits: 10},
			{FullName: "repo2", Commits: 50},
			{FullName: "repo3", Commits: 30},
		},
	}

	// Store original order
	originalFirst := profile.RepoStats[0].FullName

	// Get sorted top repos
	_ = profile.TopReposByCommits(2)

	// Verify original slice was not mutated
	if profile.RepoStats[0].FullName != originalFirst {
		t.Error("TopReposByCommits mutated the original RepoStats slice")
	}
}

func TestTopReposByAdditionsDoesNotMutate(t *testing.T) {
	profile := &UserProfile{
		RepoStats: []RepoContribution{
			{FullName: "repo1", Additions: 100},
			{FullName: "repo2", Additions: 500},
			{FullName: "repo3", Additions: 300},
		},
	}

	originalFirst := profile.RepoStats[0].FullName

	_ = profile.TopReposByAdditions(2)

	if profile.RepoStats[0].FullName != originalFirst {
		t.Error("TopReposByAdditions mutated the original RepoStats slice")
	}
}
