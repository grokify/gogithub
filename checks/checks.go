// Package checks provides GitHub check runs operations.
package checks

import (
	"context"
	"time"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ListCheckRuns lists check runs for a commit SHA or branch.
func ListCheckRuns(ctx context.Context, client clientv1.Client, owner, repo, ref string) ([]*gogithub.CheckRun, error) {
	return client.ListCheckRuns(ctx, owner, repo, ref)
}

// ListCheckRunsForPR lists check runs for a pull request.
func ListCheckRunsForPR(ctx context.Context, client clientv1.Client, owner, repo string, prNumber int) ([]*gogithub.CheckRun, error) {
	pr, err := client.GetPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, err
	}

	if pr.Head == nil || pr.Head.SHA == "" {
		return nil, nil
	}

	return client.ListCheckRuns(ctx, owner, repo, pr.Head.SHA)
}

// ChecksStatus represents the aggregate status of check runs.
type ChecksStatus struct {
	Total      int
	Passed     int
	Failed     int
	Pending    int
	AllPassed  bool
	AnyFailed  bool
	AnyPending bool
}

// GetChecksStatus returns aggregate status of check runs.
func GetChecksStatus(checks []*gogithub.CheckRun) *ChecksStatus {
	status := &ChecksStatus{
		Total: len(checks),
	}

	for _, c := range checks {
		switch {
		case c.Status != "completed":
			status.Pending++
			status.AnyPending = true
		case c.Conclusion == "success":
			status.Passed++
		default:
			status.Failed++
			status.AnyFailed = true
		}
	}

	status.AllPassed = status.Total > 0 && status.Passed == status.Total

	return status
}

// AllChecksPassed returns true if all check runs completed successfully.
func AllChecksPassed(checks []*gogithub.CheckRun) bool {
	if len(checks) == 0 {
		return false
	}

	for _, c := range checks {
		if c.Status != "completed" || c.Conclusion != "success" {
			return false
		}
	}
	return true
}

// WaitForChecks polls until all checks complete or timeout.
// Returns the final check runs and whether all passed.
func WaitForChecks(ctx context.Context, client clientv1.Client, owner, repo, ref string, timeout, pollInterval time.Duration) ([]*gogithub.CheckRun, bool, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		checks, err := client.ListCheckRuns(ctx, owner, repo, ref)
		if err != nil {
			return nil, false, err
		}

		allComplete := true
		for _, c := range checks {
			if c.Status != "completed" {
				allComplete = false
				break
			}
		}

		if allComplete {
			return checks, AllChecksPassed(checks), nil
		}

		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}

	// Return current state after timeout
	checks, err := client.ListCheckRuns(ctx, owner, repo, ref)
	if err != nil {
		return nil, false, err
	}
	return checks, AllChecksPassed(checks), nil
}

// GetCheckRun retrieves a specific check run by ID.
func GetCheckRun(ctx context.Context, client clientv1.Client, owner, repo string, checkRunID int64) (*gogithub.CheckRun, error) {
	return client.GetCheckRun(ctx, owner, repo, checkRunID)
}

// ListCheckSuites lists check suites for a commit.
func ListCheckSuites(ctx context.Context, client clientv1.Client, owner, repo, ref string) ([]*gogithub.CheckSuite, error) {
	return client.ListCheckSuites(ctx, owner, repo, ref)
}
