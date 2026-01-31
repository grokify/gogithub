// Package checks provides GitHub check runs operations.
package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v82/github"
)

// ListCheckRuns lists check runs for a commit SHA or branch.
func ListCheckRuns(ctx context.Context, gh *github.Client, owner, repo, ref string) ([]*github.CheckRun, error) {
	var allChecks []*github.CheckRun

	opts := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		result, resp, err := gh.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, opts)
		if err != nil {
			return nil, fmt.Errorf("list check runs: %w", err)
		}

		allChecks = append(allChecks, result.CheckRuns...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allChecks, nil
}

// ListCheckRunsForPR lists check runs for a pull request.
func ListCheckRunsForPR(ctx context.Context, gh *github.Client, owner, repo string, prNumber int) ([]*github.CheckRun, error) {
	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("get PR: %w", err)
	}

	sha := pr.GetHead().GetSHA()
	if sha == "" {
		return nil, nil
	}

	return ListCheckRuns(ctx, gh, owner, repo, sha)
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
func GetChecksStatus(checks []*github.CheckRun) *ChecksStatus {
	status := &ChecksStatus{
		Total: len(checks),
	}

	for _, c := range checks {
		switch {
		case c.GetStatus() != "completed":
			status.Pending++
			status.AnyPending = true
		case c.GetConclusion() == "success":
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
func AllChecksPassed(checks []*github.CheckRun) bool {
	if len(checks) == 0 {
		return false
	}

	for _, c := range checks {
		if c.GetStatus() != "completed" || c.GetConclusion() != "success" {
			return false
		}
	}
	return true
}

// WaitForChecks polls until all checks complete or timeout.
// Returns the final check runs and whether all passed.
func WaitForChecks(ctx context.Context, gh *github.Client, owner, repo, ref string, timeout, pollInterval time.Duration) ([]*github.CheckRun, bool, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		checks, err := ListCheckRuns(ctx, gh, owner, repo, ref)
		if err != nil {
			return nil, false, err
		}

		allComplete := true
		for _, c := range checks {
			if c.GetStatus() != "completed" {
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
	checks, err := ListCheckRuns(ctx, gh, owner, repo, ref)
	if err != nil {
		return nil, false, err
	}
	return checks, AllChecksPassed(checks), nil
}

// GetCheckRun retrieves a specific check run by ID.
func GetCheckRun(ctx context.Context, gh *github.Client, owner, repo string, checkRunID int64) (*github.CheckRun, error) {
	check, _, err := gh.Checks.GetCheckRun(ctx, owner, repo, checkRunID)
	return check, err
}

// ListCheckSuites lists check suites for a commit.
func ListCheckSuites(ctx context.Context, gh *github.Client, owner, repo, ref string) ([]*github.CheckSuite, error) {
	var allSuites []*github.CheckSuite

	opts := &github.ListCheckSuiteOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		result, resp, err := gh.Checks.ListCheckSuitesForRef(ctx, owner, repo, ref, opts)
		if err != nil {
			return nil, fmt.Errorf("list check suites: %w", err)
		}

		allSuites = append(allSuites, result.CheckSuites...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allSuites, nil
}
