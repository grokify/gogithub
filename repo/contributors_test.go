package repo

import (
	"testing"
	"time"
)

func TestContributorSummary(t *testing.T) {
	summary := &ContributorSummary{
		Username:       "testuser",
		TotalCommits:   100,
		TotalAdditions: 5000,
		TotalDeletions: 2000,
		FirstCommit:    time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		LastCommit:     time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC),
	}

	if summary.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", summary.Username)
	}

	if summary.TotalCommits != 100 {
		t.Errorf("expected 100 commits, got %d", summary.TotalCommits)
	}

	if summary.TotalAdditions != 5000 {
		t.Errorf("expected 5000 additions, got %d", summary.TotalAdditions)
	}

	if summary.TotalDeletions != 2000 {
		t.Errorf("expected 2000 deletions, got %d", summary.TotalDeletions)
	}

	// Verify date range
	expectedDuration := summary.LastCommit.Sub(summary.FirstCommit)
	if expectedDuration.Hours() < 24*365 { // Should be > 1 year
		t.Errorf("expected duration > 1 year, got %v", expectedDuration)
	}
}

func TestContributorSummaryZeroValues(t *testing.T) {
	summary := &ContributorSummary{}

	if summary.Username != "" {
		t.Errorf("expected empty username, got %q", summary.Username)
	}

	if summary.TotalCommits != 0 {
		t.Errorf("expected 0 commits, got %d", summary.TotalCommits)
	}

	if !summary.FirstCommit.IsZero() {
		t.Errorf("expected zero FirstCommit, got %v", summary.FirstCommit)
	}

	if !summary.LastCommit.IsZero() {
		t.Errorf("expected zero LastCommit, got %v", summary.LastCommit)
	}
}
