package search

import (
	"testing"
	"time"

	"github.com/grokify/gogithub"
)

func TestIssueAuthorUsername(t *testing.T) {
	tests := []struct {
		name        string
		issue       *Issue
		expected    string
		expectError error
	}{
		{
			name:        "nil issue",
			issue:       &Issue{Issue: nil},
			expected:    "",
			expectError: ErrIssueIsNotSet,
		},
		{
			name:        "nil user",
			issue:       &Issue{Issue: &gogithub.Issue{User: nil}},
			expected:    "",
			expectError: ErrUserIsNotSet,
		},
		{
			name:        "empty login",
			issue:       &Issue{Issue: &gogithub.Issue{User: &gogithub.User{Login: ""}}},
			expected:    "",
			expectError: ErrUserLoginIsNotSet,
		},
		{
			name:        "valid username",
			issue:       &Issue{Issue: &gogithub.Issue{User: &gogithub.User{Login: "grokify"}}},
			expected:    "grokify",
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.issue.AuthorUsername()
			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("AuthorUsername() error = %v, want %v", err, tt.expectError)
				}
			} else {
				if err != nil {
					t.Errorf("AuthorUsername() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("AuthorUsername() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestIssueAuthorUserID(t *testing.T) {
	tests := []struct {
		name        string
		issue       *Issue
		expected    int64
		expectError error
	}{
		{
			name:        "nil issue",
			issue:       &Issue{Issue: nil},
			expected:    -1,
			expectError: ErrIssueIsNotSet,
		},
		{
			name:        "nil user",
			issue:       &Issue{Issue: &gogithub.Issue{User: nil}},
			expected:    -1,
			expectError: ErrUserIsNotSet,
		},
		{
			name:        "valid user ID",
			issue:       &Issue{Issue: &gogithub.Issue{User: &gogithub.User{ID: 12345}}},
			expected:    12345,
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.issue.AuthorUserID()
			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("AuthorUserID() error = %v, want %v", err, tt.expectError)
				}
			} else {
				if err != nil {
					t.Errorf("AuthorUserID() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("AuthorUserID() = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}

func TestIssueCreatedTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		issue       *Issue
		expectError error
	}{
		{
			name:        "nil issue",
			issue:       &Issue{Issue: nil},
			expectError: ErrIssueIsNotSet,
		},
		{
			name:        "zero CreatedAt",
			issue:       &Issue{Issue: &gogithub.Issue{CreatedAt: time.Time{}}},
			expectError: ErrIssueCreatedAtIsNotSet,
		},
		{
			name:        "valid time",
			issue:       &Issue{Issue: &gogithub.Issue{CreatedAt: now}},
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.issue.CreatedTime()
			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("CreatedTime() error = %v, want %v", err, tt.expectError)
				}
			} else {
				if err != nil {
					t.Errorf("CreatedTime() unexpected error: %v", err)
				}
				if result.IsZero() {
					t.Error("CreatedTime() returned zero time for valid input")
				}
			}
		})
	}
}

func TestIssueCreatedAge(t *testing.T) {
	// Test with a time in the past
	pastTime := time.Now().Add(-24 * time.Hour)

	issue := &Issue{Issue: &gogithub.Issue{CreatedAt: pastTime}}
	age, err := issue.CreatedAge()
	if err != nil {
		t.Errorf("CreatedAge() unexpected error: %v", err)
	}

	// Age should be approximately 24 hours (with some tolerance)
	if age < 23*time.Hour || age > 25*time.Hour {
		t.Errorf("CreatedAge() = %v, expected approximately 24 hours", age)
	}
}

func TestIssueCreatedAgeError(t *testing.T) {
	issue := &Issue{Issue: nil}
	_, err := issue.CreatedAge()
	if err != ErrIssueIsNotSet {
		t.Errorf("CreatedAge() error = %v, want %v", err, ErrIssueIsNotSet)
	}
}

func TestIssueMustAuthorUsername(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected string
	}{
		{
			name:     "nil issue returns empty",
			issue:    &Issue{Issue: nil},
			expected: "",
		},
		{
			name:     "valid username",
			issue:    &Issue{Issue: &gogithub.Issue{User: &gogithub.User{Login: "grokify"}}},
			expected: "grokify",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.issue.MustAuthorUsername()
			if result != tt.expected {
				t.Errorf("MustAuthorUsername() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIssueMustAuthorUserID(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected int64
	}{
		{
			name:     "nil issue returns -1",
			issue:    &Issue{Issue: nil},
			expected: -1,
		},
		{
			name:     "valid user ID",
			issue:    &Issue{Issue: &gogithub.Issue{User: &gogithub.User{ID: 12345}}},
			expected: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.issue.MustAuthorUserID()
			if result != tt.expected {
				t.Errorf("MustAuthorUserID() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestIssuesRepositoryIssueCounts(t *testing.T) {
	issues := Issues{
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo1"},
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo1"},
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo2"},
	}

	t.Run("API URLs", func(t *testing.T) {
		counts := issues.RepositoryIssueCounts(false)

		if counts["https://api.github.com/repos/owner/repo1"] != 2 {
			t.Errorf("repo1 count = %d, want 2", counts["https://api.github.com/repos/owner/repo1"])
		}
		if counts["https://api.github.com/repos/owner/repo2"] != 1 {
			t.Errorf("repo2 count = %d, want 1", counts["https://api.github.com/repos/owner/repo2"])
		}
	})

	t.Run("HTML URLs", func(t *testing.T) {
		counts := issues.RepositoryIssueCounts(true)

		if counts["https://github.com/owner/repo1"] != 2 {
			t.Errorf("repo1 HTML count = %d, want 2", counts["https://github.com/owner/repo1"])
		}
		if counts["https://github.com/owner/repo2"] != 1 {
			t.Errorf("repo2 HTML count = %d, want 1", counts["https://github.com/owner/repo2"])
		}
	})
}

func TestIssuesRepositoryIssueCountsEmpty(t *testing.T) {
	issues := Issues{}
	counts := issues.RepositoryIssueCounts(false)

	if len(counts) != 0 {
		t.Errorf("Empty issues should return empty map, got %d entries", len(counts))
	}
}

func TestIssuesTableRepos(t *testing.T) {
	issues := Issues{
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo1"},
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo1"},
		&gogithub.Issue{RepositoryURL: "https://api.github.com/repos/owner/repo2"},
	}

	tbl := issues.TableRepos("Test Repos", true)

	if tbl.Name != "Test Repos" {
		t.Errorf("Table name = %q, want %q", tbl.Name, "Test Repos")
	}

	if len(tbl.Rows) != 2 {
		t.Errorf("Table rows = %d, want 2", len(tbl.Rows))
	}
}

func TestIssuesTable(t *testing.T) {
	now := time.Now()

	issues := Issues{
		&gogithub.Issue{
			User:      &gogithub.User{Login: "grokify", ID: 12345},
			Title:     "Test Issue",
			HTMLURL:   "https://github.com/owner/repo/issues/1",
			State:     "open",
			CreatedAt: now,
		},
	}

	tbl, err := issues.Table("Test Issues")
	if err != nil {
		t.Fatalf("Table() error: %v", err)
	}

	if tbl.Name != "Test Issues" {
		t.Errorf("Table name = %q, want %q", tbl.Name, "Test Issues")
	}

	if len(tbl.Columns) != 7 {
		t.Errorf("Table columns = %d, want 7", len(tbl.Columns))
	}

	if len(tbl.Rows) != 1 {
		t.Errorf("Table rows = %d, want 1", len(tbl.Rows))
	}

	if len(tbl.Rows) > 0 {
		row := tbl.Rows[0]
		if row[0] != "grokify" {
			t.Errorf("Author = %q, want %q", row[0], "grokify")
		}
		if row[2] != "Test Issue" {
			t.Errorf("Title = %q, want %q", row[2], "Test Issue")
		}
		if row[4] != "open" {
			t.Errorf("State = %q, want %q", row[4], "open")
		}
	}
}

func TestIssuesTableError(t *testing.T) {
	// Issue with zero CreatedAt should cause error
	issues := Issues{
		&gogithub.Issue{
			User:      &gogithub.User{Login: "grokify", ID: 12345},
			CreatedAt: time.Time{},
		},
	}

	_, err := issues.Table("Test")
	if err == nil {
		t.Error("Table() should return error for issue with zero CreatedAt")
	}
}
