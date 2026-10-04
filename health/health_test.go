package health

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// healthClient stubs the clientv1.Client methods used by Collect. Calling
// any other method panics on the nil embedded interface.
type healthClient struct {
	clientv1.Client

	repo         *gogithub.Repository
	repoErr      error
	prCount      int
	prCountErr   error
	workflows    []*gogithub.Workflow
	workflowsErr error
	runs         []*gogithub.WorkflowRun
	runsErr      error
	// perWorkflow holds the fallback ListWorkflowRuns result by workflow ID.
	perWorkflow map[int64][]*gogithub.WorkflowRun

	mu              sync.Mutex
	gotRunsOpts     *clientv1.ListWorkflowRunsOptions
	fallbackCalls   []int64
	fallbackOpts    *clientv1.ListWorkflowRunsOptions
	collectInFlight atomic.Int32
	maxInFlight     atomic.Int32
}

func (c *healthClient) GetRepository(_ context.Context, _, _ string) (*gogithub.Repository, error) {
	n := c.collectInFlight.Add(1)
	for {
		cur := c.maxInFlight.Load()
		if n <= cur || c.maxInFlight.CompareAndSwap(cur, n) {
			break
		}
	}
	return c.repo, c.repoErr
}

func (c *healthClient) CountPullRequests(_ context.Context, _, _, state string) (int, error) {
	if state != "open" {
		return 0, errors.New("unexpected state " + state)
	}
	return c.prCount, c.prCountErr
}

func (c *healthClient) ListWorkflows(_ context.Context, _, _ string) ([]*gogithub.Workflow, error) {
	return c.workflows, c.workflowsErr
}

func (c *healthClient) ListRepositoryWorkflowRuns(_ context.Context, _, _ string, opts *clientv1.ListWorkflowRunsOptions) ([]*gogithub.WorkflowRun, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gotRunsOpts = opts
	c.collectInFlight.Add(-1)
	return c.runs, c.runsErr
}

func (c *healthClient) ListWorkflowRuns(_ context.Context, _, _ string, workflowID int64, opts *clientv1.ListWorkflowRunsOptions) ([]*gogithub.WorkflowRun, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fallbackCalls = append(c.fallbackCalls, workflowID)
	c.fallbackOpts = opts
	return c.perWorkflow[workflowID], nil
}

func newRepo() *gogithub.Repository {
	return &gogithub.Repository{
		FullName:        "o/r",
		DefaultBranch:   "main",
		OpenIssuesCount: 10,
	}
}

func run(workflowID int64, status, conclusion string) *gogithub.WorkflowRun {
	return &gogithub.WorkflowRun{WorkflowID: workflowID, Status: status, Conclusion: conclusion, HeadBranch: "main"}
}

func TestRunsURL(t *testing.T) {
	repo := &gogithub.Repository{HTMLURL: "https://github.com/o/r"}
	tests := []struct {
		name     string
		repo     *gogithub.Repository
		workflow *gogithub.Workflow
		want     string
	}{
		{"workflow file", repo, &gogithub.Workflow{Path: ".github/workflows/ci.yml"}, "https://github.com/o/r/actions/workflows/ci.yml"},
		{"dynamic workflow", repo, &gogithub.Workflow{Path: "dynamic/pages/pages-build-deployment"}, "https://github.com/o/r/actions/workflows/pages/pages-build-deployment"},
		{"dependabot workflow", repo, &gogithub.Workflow{Path: "dynamic/dependabot/dependabot-updates"}, "https://github.com/o/r/actions/workflows/dependabot/dependabot-updates"},
		{"unknown prefix", repo, &gogithub.Workflow{Path: "elsewhere/nested/build.yml"}, "https://github.com/o/r/actions/workflows/build.yml"},
		{"nil repository", nil, &gogithub.Workflow{Path: ".github/workflows/ci.yml"}, ""},
		{"no repository URL", &gogithub.Repository{}, &gogithub.Workflow{Path: ".github/workflows/ci.yml"}, ""},
		{"nil workflow", repo, nil, ""},
		{"no path", repo, &gogithub.Workflow{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RunsURL(tt.repo, tt.workflow); got != tt.want {
				t.Errorf("RunsURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunState(t *testing.T) {
	tests := []struct {
		name string
		run  *gogithub.WorkflowRun
		want State
	}{
		{"nil", nil, StateNone},
		{"queued", run(1, "queued", ""), StateRunning},
		{"in progress", run(1, "in_progress", ""), StateRunning},
		{"success", run(1, RunStatusCompleted, ConclusionSuccess), StatePassing},
		{"failure", run(1, RunStatusCompleted, ConclusionFailure), StateFailing},
		{"timed out", run(1, RunStatusCompleted, ConclusionTimedOut), StateFailing},
		{"startup failure", run(1, RunStatusCompleted, ConclusionStartupFailure), StateFailing},
		{"action required", run(1, RunStatusCompleted, ConclusionActionRequired), StateFailing},
		{"cancelled", run(1, RunStatusCompleted, ConclusionCancelled), StateInconclusive},
		{"skipped", run(1, RunStatusCompleted, ConclusionSkipped), StateInconclusive},
		{"unknown conclusion", run(1, RunStatusCompleted, "something_new"), StateInconclusive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RunState(tt.run); got != tt.want {
				t.Errorf("RunState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOverallState(t *testing.T) {
	active := &gogithub.Workflow{State: WorkflowStateActive}
	disabled := &gogithub.Workflow{State: "disabled_manually"}
	tests := []struct {
		name      string
		workflows []WorkflowHealth
		want      State
	}{
		{"no workflows", nil, StateNone},
		{"all passing", []WorkflowHealth{{active, nil, StatePassing}, {active, nil, StatePassing}}, StatePassing},
		{"one failing", []WorkflowHealth{{active, nil, StatePassing}, {active, nil, StateFailing}}, StateFailing},
		{"running beats passing", []WorkflowHealth{{active, nil, StatePassing}, {active, nil, StateRunning}}, StateRunning},
		{"failing beats running", []WorkflowHealth{{active, nil, StateRunning}, {active, nil, StateFailing}}, StateFailing},
		{"passing beats inconclusive", []WorkflowHealth{{active, nil, StateInconclusive}, {active, nil, StatePassing}}, StatePassing},
		{"inconclusive beats none", []WorkflowHealth{{active, nil, StateNone}, {active, nil, StateInconclusive}}, StateInconclusive},
		{"disabled failing ignored", []WorkflowHealth{{active, nil, StatePassing}, {disabled, nil, StateFailing}}, StatePassing},
		{"only disabled", []WorkflowHealth{{disabled, nil, StateFailing}}, StateNone},
		{"nil workflow counts", []WorkflowHealth{{nil, nil, StateFailing}}, StateFailing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OverallState(tt.workflows); got != tt.want {
				t.Errorf("OverallState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollect(t *testing.T) {
	c := &healthClient{
		repo:    newRepo(),
		prCount: 3,
		workflows: []*gogithub.Workflow{
			{ID: 1, Name: "CI", State: WorkflowStateActive},
			{ID: 2, Name: "Lint", State: WorkflowStateActive},
			{ID: 3, Name: "Release", State: WorkflowStateActive},
			{ID: 4, Name: "Old", State: "disabled_inactivity"},
		},
		runs: []*gogithub.WorkflowRun{
			run(1, RunStatusCompleted, ConclusionSuccess), // latest CI
			run(2, "in_progress", ""),                     // latest Lint
			run(1, RunStatusCompleted, ConclusionFailure), // older CI, ignored
		},
		perWorkflow: map[int64][]*gogithub.WorkflowRun{
			3: {run(3, RunStatusCompleted, ConclusionFailure)},
		},
	}

	h, err := Collect(context.Background(), c, "o", "r", nil)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if h.OpenIssues != 7 {
		t.Errorf("OpenIssues = %d, want 7 (10 open_issues_count - 3 PRs)", h.OpenIssues)
	}
	if h.OpenPullRequests != 3 {
		t.Errorf("OpenPullRequests = %d, want 3", h.OpenPullRequests)
	}
	if h.Branch != "main" {
		t.Errorf("Branch = %q, want default branch main", h.Branch)
	}
	if c.gotRunsOpts == nil || c.gotRunsOpts.Branch != "main" || c.gotRunsOpts.PerPage != DefaultRunsPerPage {
		t.Errorf("ListRepositoryWorkflowRuns opts = %+v, want branch main, per page %d", c.gotRunsOpts, DefaultRunsPerPage)
	}

	wantStates := map[string]State{"CI": StatePassing, "Lint": StateRunning, "Release": StateFailing, "Old": StateNone}
	if len(h.Workflows) != len(wantStates) {
		t.Fatalf("len(Workflows) = %d, want %d", len(h.Workflows), len(wantStates))
	}
	for _, wf := range h.Workflows {
		if wf.State != wantStates[wf.Workflow.Name] {
			t.Errorf("%s: State = %q, want %q", wf.Workflow.Name, wf.State, wantStates[wf.Workflow.Name])
		}
	}
	if h.Workflows[0].LatestRun == nil || h.Workflows[0].LatestRun.Conclusion != ConclusionSuccess {
		t.Errorf("CI LatestRun = %+v, want the most recent (success) run", h.Workflows[0].LatestRun)
	}
	if h.State != StateFailing {
		t.Errorf("State = %q, want failing (Release failed)", h.State)
	}

	// Only the active workflow missing from the page triggers a fallback;
	// the disabled one does not.
	if len(c.fallbackCalls) != 1 || c.fallbackCalls[0] != 3 {
		t.Errorf("fallback calls = %v, want [3]", c.fallbackCalls)
	}
	if c.fallbackOpts == nil || c.fallbackOpts.Branch != "main" || c.fallbackOpts.PerPage != 1 {
		t.Errorf("fallback opts = %+v, want branch main, per page 1", c.fallbackOpts)
	}
}

func TestCollectOptions(t *testing.T) {
	t.Run("explicit branch", func(t *testing.T) {
		c := &healthClient{repo: newRepo(), workflows: []*gogithub.Workflow{{ID: 1, State: WorkflowStateActive}}}
		h, err := Collect(context.Background(), c, "o", "r", &Options{Branch: "develop", RunsPerPage: 25})
		if err != nil {
			t.Fatalf("Collect() error = %v", err)
		}
		if h.Branch != "develop" || c.gotRunsOpts.Branch != "develop" {
			t.Errorf("Branch = %q / opts %q, want develop", h.Branch, c.gotRunsOpts.Branch)
		}
		if c.gotRunsOpts.PerPage != 25 {
			t.Errorf("PerPage = %d, want 25", c.gotRunsOpts.PerPage)
		}
		if h.State != StateNone || len(c.fallbackCalls) != 1 {
			t.Errorf("State = %q, fallbacks = %v; want none with one fallback", h.State, c.fallbackCalls)
		}
	})

	t.Run("any branch", func(t *testing.T) {
		c := &healthClient{repo: newRepo(), workflows: []*gogithub.Workflow{{ID: 1, State: WorkflowStateActive}}}
		h, err := Collect(context.Background(), c, "o", "r", &Options{Branch: "ignored", AnyBranch: true})
		if err != nil {
			t.Fatalf("Collect() error = %v", err)
		}
		if h.Branch != "" || c.gotRunsOpts.Branch != "" || c.fallbackOpts.Branch != "" {
			t.Errorf("branch filters = %q/%q/%q, want none", h.Branch, c.gotRunsOpts.Branch, c.fallbackOpts.Branch)
		}
	})

	t.Run("no workflows skips run listing", func(t *testing.T) {
		c := &healthClient{repo: newRepo(), prCount: 12, runsErr: errors.New("should not be called")}
		h, err := Collect(context.Background(), c, "o", "r", nil)
		if err != nil {
			t.Fatalf("Collect() error = %v", err)
		}
		if h.State != StateNone || len(h.Workflows) != 0 {
			t.Errorf("State = %q, Workflows = %v; want none", h.State, h.Workflows)
		}
		if h.OpenIssues != 0 {
			t.Errorf("OpenIssues = %d, want 0 (count never negative)", h.OpenIssues)
		}
	})
}

func TestCollectErrors(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name   string
		client *healthClient
	}{
		{"repository", &healthClient{repoErr: boom}},
		{"pull request count", &healthClient{repo: newRepo(), prCountErr: boom}},
		{"workflows", &healthClient{repo: newRepo(), workflowsErr: boom}},
		{"runs", &healthClient{repo: newRepo(), workflows: []*gogithub.Workflow{{ID: 1}}, runsErr: boom}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Collect(context.Background(), tt.client, "o", "r", nil)
			if !errors.Is(err, boom) {
				t.Errorf("Collect() error = %v, want wrapped boom", err)
			}
		})
	}
}

func TestCollectAll(t *testing.T) {
	c := &healthClient{
		repo:      newRepo(),
		prCount:   1,
		workflows: []*gogithub.Workflow{{ID: 1, State: WorkflowStateActive}},
		runs:      []*gogithub.WorkflowRun{run(1, RunStatusCompleted, ConclusionSuccess)},
	}
	names := []string{"o/a", "bad-name", "o/b", "o/c"}

	results, err := CollectAll(context.Background(), c, names, &Options{Concurrency: 2})
	if err == nil {
		t.Fatal("CollectAll() error = nil, want error for bad-name")
	}
	if !strings.Contains(err.Error(), "bad-name") {
		t.Errorf("error = %v, want mention of bad-name", err)
	}
	if len(results) != len(names) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(names))
	}
	for i, r := range results {
		if r.FullName != names[i] {
			t.Errorf("results[%d].FullName = %q, want %q (input order)", i, r.FullName, names[i])
		}
		if names[i] == "bad-name" {
			if r.Err == nil || r.Health != nil {
				t.Errorf("bad-name: Err = %v, Health = %v; want error only", r.Err, r.Health)
			}
			continue
		}
		if r.Err != nil || r.Health == nil || r.Health.State != StatePassing {
			t.Errorf("%s: Err = %v, Health = %+v; want passing", r.FullName, r.Err, r.Health)
		}
	}
	if got := c.maxInFlight.Load(); got > 2 {
		t.Errorf("max concurrent collections = %d, want <= 2", got)
	}
}

func TestCollectAllPartialFailure(t *testing.T) {
	boom := errors.New("boom")
	c := &healthClient{repoErr: boom}
	results, err := CollectAll(context.Background(), c, []string{"o/a"}, nil)
	if !errors.Is(err, boom) {
		t.Errorf("CollectAll() error = %v, want wrapped boom", err)
	}
	if len(results) != 1 || !errors.Is(results[0].Err, boom) || !strings.HasPrefix(results[0].Err.Error(), "o/a: ") {
		t.Errorf("results = %+v, want one error prefixed with the repo name", results)
	}
}

func TestCollectAllEmpty(t *testing.T) {
	results, err := CollectAll(context.Background(), &healthClient{}, nil, nil)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}
