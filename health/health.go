// Package health summarizes the state of repositories: open issue and pull
// request counts, and the latest run of every GitHub Actions workflow.
//
// It is designed for collecting the same information across a set of
// repositories with one token, so it minimizes API calls. Collecting one
// repository takes four requests regardless of how many pull requests or
// workflows it has: the repository, a pull request count, the workflow list,
// and one page of recent runs across all workflows. A further request is made
// only for an active workflow whose latest run is not on that page.
package health

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/gogithub/repo"
)

// State summarizes a workflow's latest run, or a repository's workflows as a
// whole.
type State string

const (
	// StatePassing means the latest run completed successfully.
	StatePassing State = "passing"
	// StateFailing means the latest run completed unsuccessfully, including
	// timeouts, startup failures, and runs that need action.
	StateFailing State = "failing"
	// StateRunning means the latest run has not completed.
	StateRunning State = "running"
	// StateInconclusive means the latest run completed without a pass or
	// fail outcome: cancelled, skipped, neutral, or an unrecognized
	// conclusion.
	StateInconclusive State = "inconclusive"
	// StateNone means there is no run to evaluate.
	StateNone State = "none"
)

// GitHub Actions run status and conclusion values used to derive a State.
const (
	RunStatusCompleted = "completed"

	ConclusionSuccess        = "success"
	ConclusionFailure        = "failure"
	ConclusionTimedOut       = "timed_out"
	ConclusionStartupFailure = "startup_failure"
	ConclusionActionRequired = "action_required"
	ConclusionStale          = "stale"
	ConclusionCancelled      = "cancelled"
	ConclusionSkipped        = "skipped"
	ConclusionNeutral        = "neutral"
)

// WorkflowStateActive is the Workflow.State of a workflow that can run.
// Other values ("disabled_manually", "disabled_inactivity") mark workflows
// that are listed but excluded from a repository's overall State.
const WorkflowStateActive = "active"

// DefaultRunsPerPage is the number of recent runs fetched per repository.
// It is GitHub's maximum page size.
const DefaultRunsPerPage = 100

// DefaultConcurrency is the number of repositories collected at once by
// CollectAll when Options.Concurrency is zero.
const DefaultConcurrency = 4

// Options configures collection.
type Options struct {
	// Branch selects the branch whose workflow runs are evaluated. Empty
	// means the repository's default branch.
	Branch string
	// AnyBranch evaluates the latest run of each workflow regardless of
	// branch, ignoring Branch. Use it to see workflows triggered by tags or
	// releases, whose runs are not associated with a branch.
	AnyBranch bool
	// Concurrency is the number of repositories CollectAll processes at
	// once. Default: DefaultConcurrency.
	Concurrency int
	// RunsPerPage is how many recent runs to fetch per repository. Default:
	// DefaultRunsPerPage. A workflow whose latest run is older than the page
	// costs one extra request.
	RunsPerPage int
}

// WorkflowHealth is the latest run of one workflow.
type WorkflowHealth struct {
	Workflow *gogithub.Workflow
	// LatestRun is nil when the workflow has no run on the evaluated branch.
	LatestRun *gogithub.WorkflowRun
	State     State
}

// RepoHealth is the collected state of one repository.
type RepoHealth struct {
	Repository *gogithub.Repository
	// OpenIssues is the number of open issues, excluding pull requests.
	OpenIssues int
	// OpenPullRequests is the number of open pull requests.
	OpenPullRequests int
	// Branch is the branch whose runs were evaluated; empty when
	// Options.AnyBranch was set.
	Branch string
	// Workflows lists every workflow defined in the repository, active or
	// not, in the order GitHub returns them.
	Workflows []WorkflowHealth
	// State summarizes the active workflows: failing if any fails, else
	// running if any is running, else passing if any passed, else
	// inconclusive if any completed, else none.
	State State
}

// Result pairs a repository name with its collected health or the error
// that prevented collection.
type Result struct {
	// FullName is the "owner/name" given to CollectAll.
	FullName string
	// Health is nil when Err is set.
	Health *RepoHealth
	Err    error
}

// RunState derives the State of a single run. A nil run is StateNone.
func RunState(run *gogithub.WorkflowRun) State {
	if run == nil {
		return StateNone
	}
	if run.Status != RunStatusCompleted {
		return StateRunning
	}
	switch run.Conclusion {
	case ConclusionSuccess:
		return StatePassing
	case ConclusionFailure, ConclusionTimedOut, ConclusionStartupFailure, ConclusionActionRequired, ConclusionStale:
		return StateFailing
	default:
		return StateInconclusive
	}
}

// OverallState combines workflow states by severity: failing, running,
// passing, inconclusive, none. Workflows that are not active are ignored.
func OverallState(workflows []WorkflowHealth) State {
	overall := StateNone
	for _, wf := range workflows {
		if wf.Workflow != nil && wf.Workflow.State != WorkflowStateActive {
			continue
		}
		if rank(wf.State) > rank(overall) {
			overall = wf.State
		}
	}
	return overall
}

func rank(s State) int {
	switch s {
	case StateFailing:
		return 4
	case StateRunning:
		return 3
	case StatePassing:
		return 2
	case StateInconclusive:
		return 1
	default:
		return 0
	}
}

// Collect gathers the health of one repository.
func Collect(ctx context.Context, client clientv1.Client, owner, name string, opts *Options) (*RepoHealth, error) {
	if opts == nil {
		opts = &Options{}
	}

	repository, err := client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}

	openPRs, err := client.CountPullRequests(ctx, owner, name, "open")
	if err != nil {
		return nil, err
	}

	workflows, err := client.ListWorkflows(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	branch := ""
	if !opts.AnyBranch {
		branch = opts.Branch
		if branch == "" {
			branch = repository.DefaultBranch
		}
	}

	health := &RepoHealth{
		Repository:       repository,
		OpenIssues:       max(repository.OpenIssuesCount-openPRs, 0),
		OpenPullRequests: openPRs,
		Branch:           branch,
	}
	if len(workflows) == 0 {
		health.State = StateNone
		return health, nil
	}

	perPage := opts.RunsPerPage
	if perPage <= 0 {
		perPage = DefaultRunsPerPage
	}
	runs, err := client.ListRepositoryWorkflowRuns(ctx, owner, name, &clientv1.ListWorkflowRunsOptions{
		Branch:  branch,
		PerPage: perPage,
	})
	if err != nil {
		return nil, err
	}
	latest := latestRunsByWorkflow(runs)

	health.Workflows = make([]WorkflowHealth, 0, len(workflows))
	for _, wf := range workflows {
		run, ok := latest[wf.ID]
		if !ok && wf.State == WorkflowStateActive {
			// The latest run is older than the page of recent runs, which
			// happens to workflows that run rarely next to ones that run
			// often. Ask for that workflow alone.
			run, err = latestWorkflowRun(ctx, client, owner, name, wf.ID, branch)
			if err != nil {
				return nil, err
			}
		}
		health.Workflows = append(health.Workflows, WorkflowHealth{
			Workflow:  wf,
			LatestRun: run,
			State:     RunState(run),
		})
	}
	health.State = OverallState(health.Workflows)
	return health, nil
}

// latestRunsByWorkflow keeps the first run seen per workflow. Runs arrive
// most recent first.
func latestRunsByWorkflow(runs []*gogithub.WorkflowRun) map[int64]*gogithub.WorkflowRun {
	latest := make(map[int64]*gogithub.WorkflowRun)
	for _, run := range runs {
		if _, seen := latest[run.WorkflowID]; !seen {
			latest[run.WorkflowID] = run
		}
	}
	return latest
}

func latestWorkflowRun(ctx context.Context, client clientv1.Client, owner, name string, workflowID int64, branch string) (*gogithub.WorkflowRun, error) {
	runs, err := client.ListWorkflowRuns(ctx, owner, name, workflowID, &clientv1.ListWorkflowRunsOptions{
		Branch:  branch,
		PerPage: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return runs[0], nil
}

// CollectAll gathers the health of several repositories, given as
// "owner/name", collecting opts.Concurrency repositories at a time. Results
// are in input order. A repository that cannot be collected has Result.Err
// set and does not stop the others; the returned error joins every
// per-repository error, so callers that want partial results should use the
// results even when err is non-nil.
func CollectAll(ctx context.Context, client clientv1.Client, fullNames []string, opts *Options) ([]Result, error) {
	if opts == nil {
		opts = &Options{}
	}
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	results := make([]Result, len(fullNames))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, fullName := range fullNames {
		results[i].FullName = fullName
		owner, name, err := repo.ParseRepoName(fullName)
		if err != nil {
			results[i].Err = err
			continue
		}
		wg.Add(1)
		go func(i int, owner, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			health, err := Collect(ctx, client, owner, name, opts)
			if err != nil {
				results[i].Err = fmt.Errorf("%s/%s: %w", owner, name, err)
				return
			}
			results[i].Health = health
		}(i, owner, name)
	}
	wg.Wait()

	errs := make([]error, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, r.Err)
		}
	}
	return results, errors.Join(errs...)
}
