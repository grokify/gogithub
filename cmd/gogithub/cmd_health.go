package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
	"github.com/grokify/gogithub/etagcache"
	"github.com/grokify/gogithub/health"
	"github.com/spf13/cobra"
)

var (
	healthRepos       []string
	healthReposFile   string
	healthBranch      string
	healthAnyBranch   bool
	healthConcurrency int
	healthFormat      string
	healthCacheDir    string
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show open issues, open pull requests, and workflow status for repositories",
	Long: `Show the health of a set of repositories: open issue and pull request
counts, and the latest run of every GitHub Actions workflow.

Each repository costs four API requests regardless of how many pull requests
or workflows it has, so a set of repositories can be refreshed frequently
within the API rate limit. With --cache-dir, responses are cached between
runs and GitHub is asked only whether each one changed; unchanged responses
do not count against the rate limit.

Workflow runs are evaluated on the default branch unless --branch or
--any-branch is given. A repository is failing if any active workflow's
latest run failed, running if any is in progress, and passing otherwise.
Disabled workflows are listed but do not affect the repository state.

Examples:
  gogithub health --repo grokify/gogithub --repo grokify/mogo
  gogithub health --repos-file repos.txt            # One owner/name per line
  gogithub health --repos-file repos.txt -f json    # For dashboards
  gogithub health --repo owner/name --any-branch    # Include tag-triggered workflows
  gogithub health --repos-file repos.txt --cache-dir ~/.cache/gogithub   # Conditional requests

Authentication:
  GITHUB_TOKEN    Token with the 'repo' scope (private repositories and
                  Actions), or a fine-grained token with metadata, pull
                  requests, and actions read access to each repository.`,
	RunE: runHealth,
	// Errors here are per-repository collection failures, not usage errors,
	// and main prints them.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	healthCmd.Flags().StringSliceVar(&healthRepos, "repo", nil, "Repository to check (owner/name); repeatable")
	healthCmd.Flags().StringVar(&healthReposFile, "repos-file", "", "File with one owner/name per line (# comments allowed)")
	healthCmd.Flags().StringVar(&healthBranch, "branch", "", "Branch to evaluate workflow runs on (default: each repository's default branch)")
	healthCmd.Flags().BoolVar(&healthAnyBranch, "any-branch", false, "Evaluate each workflow's latest run on any branch or tag")
	healthCmd.Flags().IntVar(&healthConcurrency, "concurrency", health.DefaultConcurrency, "Repositories to collect at once")
	healthCmd.Flags().StringVarP(&healthFormat, "format", "f", repoAccessFormatText, "Output format: text or json")
	healthCmd.Flags().StringVar(&healthCacheDir, "cache-dir", "", "Directory for cached responses, enabling conditional requests across runs")
	healthCmd.MarkFlagsMutuallyExclusive("branch", "any-branch")
}

// healthEntry is one repository in health output.
type healthEntry struct {
	Repository       string                `json:"repository"`
	State            string                `json:"state"`
	OpenIssues       int                   `json:"openIssues"`
	OpenPullRequests int                   `json:"openPullRequests"`
	Branch           string                `json:"branch,omitempty"`
	URL              string                `json:"url,omitempty"`
	Workflows        []healthWorkflowEntry `json:"workflows"`
	Error            string                `json:"error,omitempty"`
}

// healthWorkflowEntry is one workflow in health output.
type healthWorkflowEntry struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	WorkflowState string `json:"workflowState"`
	State         string `json:"state"`
	// WorkflowURL is the workflow definition file; RunsURL lists all of its
	// runs; RunURL is the latest run; BadgeURL is the status badge image.
	WorkflowURL   string     `json:"workflowUrl,omitempty"`
	RunsURL       string     `json:"runsUrl,omitempty"`
	RunStatus     string     `json:"runStatus,omitempty"`
	RunConclusion string     `json:"runConclusion,omitempty"`
	RunBranch     string     `json:"runBranch,omitempty"`
	RunEvent      string     `json:"runEvent,omitempty"`
	RunURL        string     `json:"runUrl,omitempty"`
	RunUpdatedAt  *time.Time `json:"runUpdatedAt,omitempty"`
	BadgeURL      string     `json:"badgeUrl,omitempty"`
}

func newHealthEntry(r health.Result) healthEntry {
	entry := healthEntry{Repository: r.FullName, Workflows: []healthWorkflowEntry{}}
	if r.Err != nil {
		entry.State = "error"
		entry.Error = r.Err.Error()
		return entry
	}
	h := r.Health
	entry.State = string(h.State)
	entry.OpenIssues = h.OpenIssues
	entry.OpenPullRequests = h.OpenPullRequests
	entry.Branch = h.Branch
	if h.Repository != nil {
		entry.Repository = h.Repository.FullName
		entry.URL = h.Repository.HTMLURL
	}
	for _, wf := range h.Workflows {
		entry.Workflows = append(entry.Workflows, newHealthWorkflowEntry(h.Repository, wf))
	}
	return entry
}

func newHealthWorkflowEntry(repository *gogithub.Repository, wf health.WorkflowHealth) healthWorkflowEntry {
	entry := healthWorkflowEntry{State: string(wf.State)}
	if wf.Workflow != nil {
		entry.Name = wf.Workflow.Name
		entry.Path = wf.Workflow.Path
		entry.WorkflowState = wf.Workflow.State
		entry.WorkflowURL = wf.Workflow.HTMLURL
		entry.RunsURL = health.RunsURL(repository, wf.Workflow)
		entry.BadgeURL = wf.Workflow.BadgeURL
	}
	if run := wf.LatestRun; run != nil {
		entry.RunStatus = run.Status
		entry.RunConclusion = run.Conclusion
		entry.RunBranch = run.HeadBranch
		entry.RunEvent = run.Event
		entry.RunURL = run.HTMLURL
		if !run.UpdatedAt.IsZero() {
			updated := run.UpdatedAt
			entry.RunUpdatedAt = &updated
		}
	}
	return entry
}

func runHealth(cmd *cobra.Command, args []string) error {
	if healthFormat != repoAccessFormatText && healthFormat != repoAccessFormatJSON {
		return fmt.Errorf("unsupported format %q (use %s or %s)", healthFormat, repoAccessFormatText, repoAccessFormatJSON)
	}
	repos, err := healthRepoList(healthRepos, healthReposFile)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("no repositories given; use --repo or --repos-file")
	}

	ctx := context.Background()
	client, cache, err := newHealthClient(ctx, healthCacheDir)
	if err != nil {
		return err
	}

	results, collectErr := health.CollectAll(ctx, client, repos, &health.Options{
		Branch:      healthBranch,
		AnyBranch:   healthAnyBranch,
		Concurrency: healthConcurrency,
	})
	entries := make([]healthEntry, 0, len(results))
	for _, r := range results {
		entries = append(entries, newHealthEntry(r))
	}
	if err := writeHealth(os.Stdout, entries, healthFormat); err != nil {
		return err
	}
	if cache != nil {
		fmt.Fprintln(os.Stderr, cache.Stats())
	}
	// Partial results were written; report what failed and exit non-zero.
	return collectErr
}

// newHealthClient creates a client from GITHUB_TOKEN. With a cache
// directory, requests go through an ETag cache persisted there, and the
// returned transport reports hit statistics.
func newHealthClient(ctx context.Context, cacheDir string) (clientv1.Client, *etagcache.Transport, error) {
	opts := clientv1.ClientOptions{Token: ensureToken()}
	var cache *etagcache.Transport
	if cacheDir != "" {
		store, err := etagcache.NewFileStore(cacheDir)
		if err != nil {
			return nil, nil, err
		}
		cache = etagcache.NewTransportWithStore(nil, store)
		cache.OnStoreError = func(err error) { fmt.Fprintln(os.Stderr, "cache:", err) }
		opts.Transport = cache
	}
	client, err := clientv1.NewClientWithOptions(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("creating github client: %w", err)
	}
	return client, cache, nil
}

// healthRepoList merges --repo values with the lines of --repos-file,
// skipping blank lines and # comments, and removes duplicates.
func healthRepoList(repos []string, file string) ([]string, error) {
	all := append([]string(nil), repos...)
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		fromFile, err := readRepoLines(strings.NewReader(string(data)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		all = append(all, fromFile...)
	}
	seen := make(map[string]bool, len(all))
	unique := make([]string, 0, len(all))
	for _, r := range all {
		r = strings.TrimSpace(r)
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		unique = append(unique, r)
	}
	return unique, nil
}

func readRepoLines(r io.Reader) ([]string, error) {
	var repos []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		repos = append(repos, line)
	}
	return repos, scanner.Err()
}

func writeHealth(w io.Writer, entries []healthEntry, format string) error {
	if format == repoAccessFormatJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "REPOSITORY\tSTATE\tISSUES\tPRS\tWORKFLOWS"); err != nil {
		return err
	}
	for _, e := range entries {
		if e.Error != "" {
			if _, err := fmt.Fprintf(tw, "%s\t%s\t\t\t%s\n", e.Repository, e.State, e.Error); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n", e.Repository, e.State, e.OpenIssues, e.OpenPullRequests, workflowSummary(e.Workflows)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	// Per-workflow detail for every repository that has workflows.
	tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "\nREPOSITORY\tWORKFLOW\tSTATE\tCONCLUSION\tBRANCH\tUPDATED\tURL"); err != nil {
		return err
	}
	for _, e := range entries {
		for _, wf := range e.Workflows {
			updated := ""
			if wf.RunUpdatedAt != nil {
				updated = wf.RunUpdatedAt.UTC().Format(time.RFC3339)
			}
			conclusion := wf.RunConclusion
			if wf.WorkflowState != health.WorkflowStateActive && wf.WorkflowState != "" {
				conclusion = wf.WorkflowState
			}
			if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", e.Repository, wf.Name, wf.State, conclusion, wf.RunBranch, updated, wf.RunURL); err != nil {
				return err
			}
		}
	}
	return tw.Flush()
}

// workflowSummary renders "2 passing, 1 failing" style counts of active
// workflows for the repository table.
func workflowSummary(workflows []healthWorkflowEntry) string {
	counts := make(map[string]int)
	for _, wf := range workflows {
		if wf.WorkflowState != "" && wf.WorkflowState != health.WorkflowStateActive {
			continue
		}
		counts[wf.State]++
	}
	var parts []string
	for _, state := range []health.State{health.StateFailing, health.StateRunning, health.StatePassing, health.StateInconclusive, health.StateNone} {
		if n := counts[string(state)]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, state))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}
