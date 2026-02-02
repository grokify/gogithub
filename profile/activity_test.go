package profile

import (
	"testing"
	"time"
)

func TestMonthlyActivityYearMonth(t *testing.T) {
	m := MonthlyActivity{Year: 2024, Month: time.January}
	if got := m.YearMonth(); got != "2024-01" {
		t.Errorf("YearMonth() = %q, want %q", got, "2024-01")
	}

	m = MonthlyActivity{Year: 2024, Month: time.December}
	if got := m.YearMonth(); got != "2024-12" {
		t.Errorf("YearMonth() = %q, want %q", got, "2024-12")
	}
}

func TestMonthlyActivityMonthName(t *testing.T) {
	m := MonthlyActivity{Year: 2024, Month: time.January}
	if got := m.MonthName(); got != "January" {
		t.Errorf("MonthName() = %q, want %q", got, "January")
	}
}

func TestMonthlyActivityTotalContributions(t *testing.T) {
	m := MonthlyActivity{
		Commits: 50,
		Issues:  10,
		PRs:     5,
		Reviews: 15,
	}

	if got := m.TotalContributions(); got != 80 {
		t.Errorf("TotalContributions() = %d, want %d", got, 80)
	}
}

func TestMonthlyActivityCommitRepoCount(t *testing.T) {
	m := MonthlyActivity{
		CommitsByRepo: map[string]int{
			"user/repo1": 10,
			"user/repo2": 20,
			"org/repo3":  5,
		},
	}

	if got := m.CommitRepoCount(); got != 3 {
		t.Errorf("CommitRepoCount() = %d, want %d", got, 3)
	}
}

func TestMonthlyActivityCommitSummary(t *testing.T) {
	tests := []struct {
		name     string
		activity MonthlyActivity
		expected string
	}{
		{
			name:     "no commits",
			activity: MonthlyActivity{Commits: 0},
			expected: "",
		},
		{
			name: "single repo",
			activity: MonthlyActivity{
				Commits:       10,
				CommitsByRepo: map[string]int{"user/repo": 10},
			},
			expected: "Created 10 commits in 1 repository",
		},
		{
			name: "multiple repos",
			activity: MonthlyActivity{
				Commits: 25,
				CommitsByRepo: map[string]int{
					"user/repo1": 15,
					"user/repo2": 10,
				},
			},
			expected: "Created 25 commits in 2 repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.activity.CommitSummary()
			if got != tt.expected {
				t.Errorf("CommitSummary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonthlyActivityPRSummary(t *testing.T) {
	tests := []struct {
		name     string
		activity MonthlyActivity
		expected string
	}{
		{
			name:     "no PRs",
			activity: MonthlyActivity{PRs: 0},
			expected: "",
		},
		{
			name: "single repo",
			activity: MonthlyActivity{
				PRs:     3,
				PRRepos: []string{"user/repo"},
			},
			expected: "Opened 3 pull requests in 1 repository",
		},
		{
			name: "multiple repos",
			activity: MonthlyActivity{
				PRs:     5,
				PRRepos: []string{"user/repo1", "user/repo2", "org/repo3"},
			},
			expected: "Opened 5 pull requests in 3 repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.activity.PRSummary()
			if got != tt.expected {
				t.Errorf("PRSummary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonthlyActivityIssueSummary(t *testing.T) {
	tests := []struct {
		name     string
		activity MonthlyActivity
		expected string
	}{
		{
			name:     "no issues",
			activity: MonthlyActivity{Issues: 0},
			expected: "",
		},
		{
			name: "multiple repos",
			activity: MonthlyActivity{
				Issues:     8,
				IssueRepos: []string{"repo1", "repo2"},
			},
			expected: "Opened 8 issues in 2 repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.activity.IssueSummary()
			if got != tt.expected {
				t.Errorf("IssueSummary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonthlyActivityReviewSummary(t *testing.T) {
	tests := []struct {
		name     string
		reviews  int
		expected string
	}{
		{"no reviews", 0, ""},
		{"single review", 1, "Reviewed 1 pull request"},
		{"multiple reviews", 5, "Reviewed 5 pull requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MonthlyActivity{Reviews: tt.reviews}
			got := m.ReviewSummary()
			if got != tt.expected {
				t.Errorf("ReviewSummary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonthlyActivityRepoCreatedSummary(t *testing.T) {
	tests := []struct {
		name         string
		reposCreated []string
		expected     string
	}{
		{"no repos", nil, ""},
		{"single repo", []string{"user/newrepo"}, "Created 1 repository: user/newrepo"},
		{"multiple repos", []string{"repo1", "repo2", "repo3"}, "Created 3 repositories"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MonthlyActivity{ReposCreated: tt.reposCreated}
			got := m.RepoCreatedSummary()
			if got != tt.expected {
				t.Errorf("RepoCreatedSummary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonthlyActivityTopCommitRepos(t *testing.T) {
	m := MonthlyActivity{
		CommitsByRepo: map[string]int{
			"user/repo1": 50,
			"user/repo2": 30,
			"user/repo3": 100,
			"user/repo4": 10,
		},
	}

	top := m.TopCommitRepos(2)
	if len(top) != 2 {
		t.Fatalf("TopCommitRepos(2) returned %d repos, want 2", len(top))
	}

	if top[0].Repo != "user/repo3" || top[0].Commits != 100 {
		t.Errorf("First repo = %v, want user/repo3 with 100 commits", top[0])
	}

	if top[1].Repo != "user/repo1" || top[1].Commits != 50 {
		t.Errorf("Second repo = %v, want user/repo1 with 50 commits", top[1])
	}
}

func TestMonthlyActivityTopCommitReposAll(t *testing.T) {
	m := MonthlyActivity{
		CommitsByRepo: map[string]int{
			"repo1": 10,
			"repo2": 20,
		},
	}

	// Request more than available
	top := m.TopCommitRepos(10)
	if len(top) != 2 {
		t.Errorf("TopCommitRepos(10) returned %d repos, want 2", len(top))
	}

	// Zero means all
	top = m.TopCommitRepos(0)
	if len(top) != 2 {
		t.Errorf("TopCommitRepos(0) returned %d repos, want 2", len(top))
	}
}

func TestActivityTimelineTotalCommits(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.January, Commits: 50},
			{Year: 2024, Month: time.February, Commits: 30},
			{Year: 2024, Month: time.March, Commits: 20},
		},
	}

	if got := timeline.TotalCommits(); got != 100 {
		t.Errorf("TotalCommits() = %d, want 100", got)
	}
}

func TestActivityTimelineTotalContributions(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Commits: 50, Issues: 5, PRs: 3, Reviews: 10},
			{Commits: 30, Issues: 2, PRs: 1, Reviews: 5},
		},
	}

	if got := timeline.TotalContributions(); got != 106 {
		t.Errorf("TotalContributions() = %d, want 106", got)
	}
}

func TestActivityTimelineMostActiveMonth(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.January, Commits: 50, Issues: 5},
			{Year: 2024, Month: time.February, Commits: 100, Issues: 10}, // Most active (110)
			{Year: 2024, Month: time.March, Commits: 30, Issues: 2},
		},
	}

	most := timeline.MostActiveMonth()
	if most == nil {
		t.Fatal("MostActiveMonth() returned nil")
	}

	if most.Month != time.February {
		t.Errorf("MostActiveMonth() = %v, want February", most.Month)
	}
}

func TestActivityTimelineMostActiveMonthEmpty(t *testing.T) {
	timeline := &ActivityTimeline{}

	most := timeline.MostActiveMonth()
	if most != nil {
		t.Errorf("MostActiveMonth() should return nil for empty timeline")
	}
}

func TestActivityTimelineGetMonth(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.January, Commits: 50},
			{Year: 2024, Month: time.February, Commits: 30},
		},
	}

	m := timeline.GetMonth(2024, time.January)
	if m == nil {
		t.Fatal("GetMonth(2024, January) returned nil")
	}
	if m.Commits != 50 {
		t.Errorf("Commits = %d, want 50", m.Commits)
	}

	m = timeline.GetMonth(2024, time.March)
	if m != nil {
		t.Error("GetMonth(2024, March) should return nil")
	}
}

func TestActivityTimelineMonthsWithActivity(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.January, Commits: 50},
			{Year: 2024, Month: time.February, Commits: 0}, // No activity
			{Year: 2024, Month: time.March, Commits: 30},
		},
	}

	if got := timeline.MonthsWithActivity(); got != 2 {
		t.Errorf("MonthsWithActivity() = %d, want 2", got)
	}
}

func TestActivityTimelineAverageMonthlyContributions(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Commits: 60},
			{Commits: 40},
			{Commits: 20},
		},
	}

	avg := timeline.AverageMonthlyContributions()
	if avg != 40.0 {
		t.Errorf("AverageMonthlyContributions() = %f, want 40.0", avg)
	}
}

func TestActivityTimelineAverageMonthlyContributionsEmpty(t *testing.T) {
	timeline := &ActivityTimeline{}

	avg := timeline.AverageMonthlyContributions()
	if avg != 0.0 {
		t.Errorf("AverageMonthlyContributions() = %f, want 0.0 for empty timeline", avg)
	}
}

func TestActivityTimelineSortByDate(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.March},
			{Year: 2024, Month: time.January},
			{Year: 2023, Month: time.December},
			{Year: 2024, Month: time.February},
		},
	}

	timeline.SortByDate()

	expected := []struct {
		year  int
		month time.Month
	}{
		{2023, time.December},
		{2024, time.January},
		{2024, time.February},
		{2024, time.March},
	}

	for i, exp := range expected {
		if timeline.Months[i].Year != exp.year || timeline.Months[i].Month != exp.month {
			t.Errorf("Months[%d] = %d-%v, want %d-%v",
				i, timeline.Months[i].Year, timeline.Months[i].Month, exp.year, exp.month)
		}
	}
}

func TestActivityTimelineSortByDateDesc(t *testing.T) {
	timeline := &ActivityTimeline{
		Months: []MonthlyActivity{
			{Year: 2024, Month: time.January},
			{Year: 2024, Month: time.March},
			{Year: 2023, Month: time.December},
		},
	}

	timeline.SortByDateDesc()

	expected := []struct {
		year  int
		month time.Month
	}{
		{2024, time.March},
		{2024, time.January},
		{2023, time.December},
	}

	for i, exp := range expected {
		if timeline.Months[i].Year != exp.year || timeline.Months[i].Month != exp.month {
			t.Errorf("Months[%d] = %d-%v, want %d-%v",
				i, timeline.Months[i].Year, timeline.Months[i].Month, exp.year, exp.month)
		}
	}
}
