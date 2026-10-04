# Repository Health

The `health` package summarizes the state of repositories: open issue and pull
request counts, and the latest run of every GitHub Actions workflow. It is
built for collecting the same information across a set of repositories with
one token, so it minimizes API calls and reports partial results when some
repositories fail.

```go
import (
    "github.com/grokify/gogithub/clientv1"
    "github.com/grokify/gogithub/health"
)

client, err := clientv1.NewClient(ctx, token)

results, err := health.CollectAll(ctx, client, []string{
    "grokify/gogithub",
    "grokify/mogo",
}, nil)
for _, r := range results {
    if r.Err != nil {
        fmt.Printf("%s: %v\n", r.FullName, r.Err)
        continue
    }
    h := r.Health
    fmt.Printf("%s: %s, %d issues, %d PRs\n", r.FullName, h.State, h.OpenIssues, h.OpenPullRequests)
    for _, wf := range h.Workflows {
        fmt.Printf("  %s: %s\n", wf.Workflow.Name, wf.State)
    }
}
```

## One repository

```go
h, err := health.Collect(ctx, client, "grokify", "gogithub", nil)
```

`RepoHealth` holds:

| Field | Meaning |
|-------|---------|
| `Repository` | The `gogithub.Repository`, including `HTMLURL` and `DefaultBranch` |
| `OpenIssues` | Open issues, **excluding** pull requests |
| `OpenPullRequests` | Open pull requests |
| `Branch` | The branch whose runs were evaluated; empty with `AnyBranch` |
| `Workflows` | Every workflow defined in the repository, active or disabled |
| `State` | The combined state of the active workflows |

GitHub's `open_issues_count` counts pull requests as issues. `OpenIssues` is
that number minus the pull request count, which is why both are collected.

## Workflow states

Each `WorkflowHealth` pairs a `gogithub.Workflow` with its latest run and a
`State` derived from that run:

| State | Latest run |
|-------|-----------|
| `passing` | Completed with conclusion `success` |
| `failing` | Completed with `failure`, `timed_out`, `startup_failure`, `action_required`, or `stale` |
| `running` | Not yet completed (`queued`, `in_progress`, `waiting`, ...) |
| `inconclusive` | Completed with `cancelled`, `skipped`, `neutral`, or an unrecognized conclusion |
| `none` | No run on the evaluated branch |

`RunState` exposes this mapping for a single `gogithub.WorkflowRun`.

### Linking to workflows

Each `WorkflowHealth` carries the full `gogithub.Workflow` and latest
`gogithub.WorkflowRun`, so a dashboard can link to:

| Target | Field |
|--------|-------|
| The workflow definition file | `Workflow.HTMLURL` |
| All runs of the workflow | `health.RunsURL(repository, workflow)` — derived, since the API does not return it |
| The latest run and its logs | `LatestRun.HTMLURL` |
| The status badge image | `Workflow.BadgeURL` |

`RunsURL` handles both repository workflows (`.github/workflows/ci.yml` →
`.../actions/workflows/ci.yml`) and GitHub's own dynamic workflows
(`dynamic/pages/pages-build-deployment` →
`.../actions/workflows/pages/pages-build-deployment`).

The repository `State` is the most severe state among **active** workflows:
`failing` if any failed, else `running`, else `passing`, else `inconclusive`,
else `none`. Workflows that GitHub has disabled (`disabled_manually`,
`disabled_inactivity`) appear in `Workflows` with their `Workflow.State` but
do not affect the repository state. `OverallState` exposes this combination.

## Options

```go
results, err := health.CollectAll(ctx, client, repos, &health.Options{
    Branch:      "develop", // default: each repository's default branch
    Concurrency: 8,         // default: 4 repositories at a time
})
```

| Option | Default | Effect |
|--------|---------|--------|
| `Branch` | default branch | Branch whose runs are evaluated |
| `AnyBranch` | `false` | Evaluate each workflow's latest run on any branch, ignoring `Branch` |
| `Concurrency` | `4` | Repositories collected at once by `CollectAll` |
| `RunsPerPage` | `100` | Recent runs fetched per repository |

Runs triggered by a tag or a release are not associated with a branch, so a
release workflow shows `none` under the default branch filter. Use `AnyBranch`
to evaluate such workflows; the run's `HeadBranch` then says what triggered it.

## API cost

Collecting one repository takes four requests, independent of how many pull
requests or workflows it has:

1. `GET /repos/{owner}/{repo}` — counts and the default branch
2. `GET /repos/{owner}/{repo}/pulls?per_page=1` — the pull request count, read from the pagination header
3. `GET /repos/{owner}/{repo}/actions/workflows` — the workflows
4. `GET /repos/{owner}/{repo}/actions/runs?branch=...&per_page=100` — recent runs across all workflows

The latest run of each workflow is taken from the single page of recent runs.
An active workflow with no run on that page (one that runs rarely next to
ones that run often, or that never ran on the branch) costs one more request
for that workflow alone. Workflows with no runs at all on the branch pay this
on every collection; `AnyBranch` avoids it for workflows that run on other
refs.

A hundred repositories therefore cost roughly 400–500 requests per
collection against the 5,000 per hour core limit, so a dashboard can refresh
every few minutes. Search API endpoints, which have a separate limit of 30
requests per minute, are not used.

All four requests return `ETag`s. With an [`etagcache`](etagcache.md)
transport on the client, a refresh in which nothing changed costs no rate
limit at all, since GitHub answers `304 Not Modified`:

```go
cache := etagcache.NewTransport(nil)
client, err := clientv1.NewClientWithOptions(ctx, clientv1.ClientOptions{Token: token, Transport: cache})
```

## Partial results

`CollectAll` returns a `Result` per input repository, in input order, and
keeps collecting when one fails. The returned error joins every per-repository
error; a result with `Err` set has a nil `Health`. Use the results even when
the error is non-nil to render what succeeded:

```go
results, err := health.CollectAll(ctx, client, repos, nil)
render(results)
if err != nil {
    log.Printf("some repositories could not be collected: %v", err)
}
```

Invalid names (not `owner/name`) are reported the same way without making a
request.

## Token requirements

- A classic token needs the `repo` scope for private repositories and for
  Actions data; public repositories need no scope.
- A fine-grained token needs read access to **Metadata**, **Pull requests**,
  and **Actions** for each repository.
- Organizations that forbid personal access tokens return 404 for their
  private repositories, reported in `Result.Err`. See
  [Repository Operations](repo.md#check-access-to-a-repository) for
  detecting this and the [auth guide](auth.md#oauth-app-authentication) for
  the OAuth app alternative.

## Command line

The `gogithub health` command wraps `CollectAll` with text and JSON output.
See the [CLI guide](cli.md#health).

## Related

- [`clientv1`](clientv1.md) — `CountPullRequests`, `ListWorkflows`,
  `ListRepositoryWorkflowRuns`, `ListWorkflowRuns`
- `checks` — check runs and suites for a single commit, including waiting for
  them to complete
