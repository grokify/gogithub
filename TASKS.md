# Refactoring Tasks

Identified refactoring opportunities for gogithub, organized by priority.

## Phase 1: Quick Wins (Low Effort)

### 1.1 Move Duplicate Token Client to Shared Location
- [x] Move `NewTokenClient` from `graphql/client.go` to `auth/auth.go`
- [x] Update `graphql/client.go` to import from `auth`
- [x] Verify no circular imports

**Files:** `auth/auth.go`, `graphql/client.go`

### 1.2 Extract Hardcoded Values to Constants
- [x] Extract file mode `"100644"` in `repo/batch.go` to constant
- [x] Extract `"refs/heads/"` and `"refs/tags/"` prefixes to constants
- [x] Extract JWT expiry duration in `auth/app.go` to constant
- [x] Extract "already exists" error matching strings

**Files:** `repo/batch.go`, `repo/branch.go`, `tag/tag.go`, `auth/app.go`

### 1.3 Improve Documentation
- [x] Add field-level docs to `profile.Options` struct (already complete)
- [x] Add usage examples to `repo/batch.go` doc comments
- [x] Document `search.Must*` methods' non-panic behavior

## Phase 2: High Value (Medium Effort)

### 2.1 Use go-github Built-in Iterators
- [x] Update `release/release.go` to use `ListReleasesIter` and `ListReleaseAssetsIter`
- [x] Update `checks/checks.go` to use `ListCheckRunsForRefIter` and `ListCheckSuitesForRefIter`
- [x] Update `tag/tag.go` to use `ListTagsIter`

**Note:** Instead of creating a custom pagination helper, we leverage go-github v84's
built-in Go 1.23+ range-over-func iterators which handle pagination automatically.

**Files:** `release/release.go`, `checks/checks.go`, `tag/tag.go`

### 2.2 Standardize Error Types
- [x] Review error types for consistency
- [x] Verify all errors implement `Unwrap()` for Go 1.13+ chains

**Status:** After review, the existing error types (`AuthError`, `PRError`, `CommitError`,
`BatchError`, `BranchError`, `ForkError`) already follow a consistent Go-idiomatic pattern:
- Each has context-specific fields appropriate for its domain
- All implement `Error()` and `Unwrap()` methods
- Creating a base `OperationError` type would add complexity without benefit since Go
  favors composition over inheritance

**Files:** `errors/errors.go`, `auth/auth.go`, `pr/pullrequest.go`, `repo/*.go`

### 2.3 Extract Profile CLI Logic to Library
- [ ] Create `profile/converter/converter.go` for JSON/struct conversions
- [ ] Move `profileToRaw()`, `profileToAggregate()`, `rawToAggregate()`, `rawToProfile()` from CLI
- [ ] Create `profile/output/writer.go` for file writing operations
- [ ] Simplify `cmd/gogithub/cmd_profile.go` to orchestrate components
- [ ] Add unit tests for extracted functions

**Status:** Deferred. The CLI file is ~1100 lines with working logic. Extracting would
provide better testability and reuse but requires significant effort. Consider for a
future version when additional consumers of these conversions are needed.

**Files:** `cmd/gogithub/cmd_profile.go`, `profile/`

## Phase 3: Structural (High Effort)

### 3.1 Add Testable Interfaces
- [ ] Create `internal/ghclient/interfaces.go` with REST/GraphQL interfaces
- [ ] Update `profile.GetUserProfile()` to accept interfaces
- [ ] Create mock implementations for testing
- [ ] Update existing tests to use mocks where appropriate

**Status:** Deferred. Requires significant refactoring for modest benefit. Consider when
adding complex features that need thorough testing.

### 3.2 Split Repo Package
- [ ] Create `repo/batch/` subpackage for batch operations
- [ ] Create `repo/content/` subpackage for file content
- [ ] Create `repo/stats/` subpackage for contributor statistics
- [ ] Maintain backward compatibility with re-exports
- [ ] Update documentation

**Status:** Deferred. Would break import paths for consumers. Current package size is
manageable. Consider if repo package grows significantly.

**Files:** `repo/*.go`

### 3.3 Add Missing Test Coverage
- [x] Add tests for constants (`repo/constants_test.go`, `tag/constants_test.go`)
- [x] Add tests for error types (`repo/errors_test.go`)
- [x] Add test for `JWTExpiry` constant (`auth/app_test.go`)
- [ ] Add tests for `release/release.go` (requires HTTP mocking)
- [ ] Add tests for `graphql/*.go` (requires HTTP mocking)
- [ ] Add tests for `repo/fork.go`, `repo/branch.go`, `repo/list.go` (requires HTTP mocking)

**Note:** API-calling functions require HTTP client mocking infrastructure. Tests added
for constants, error types, and pure functions that don't need mocking.

## Phase 4: Health and Access (v0.18.0 follow-ups)

### 4.1 Fix the `Update Profile README` Workflow

- [ ] Change `go-version: '1.23'` to `go-version: 'stable'` in `.github/workflows/update-profile-readme.yml`
- [ ] Verify with a manual `workflow_dispatch` run

**Status:** Open. The scheduled run has failed every week since August at the
`Install gogithub` step: `setup-go` pins Go 1.23 with `GOTOOLCHAIN=local`, while
`go.mod` requires a newer Go, so `go install ...@latest` refuses to build. Surfaced by
`gogithub health --repo grokify/gogithub`. The workflow is a copyable example for
profile repositories, so `stable` is the right fix for people copying it too.

**Decision needed first:** once fixed, the schedule resumes committing
`chore: update profile stats SVG` to `main` weekly (the last such commits are from
early August). Options: keep as is (dogfoods the feature), drop the `schedule` trigger
in this repo so it is `workflow_dispatch`-only (users keep the schedule in their copy),
or commit the SVG to a non-`main` branch.

**Files:** `.github/workflows/update-profile-readme.yml`, `stats.svg`

### 4.2 `repo-access --external`

- [ ] Add a `--external` flag (or `repo.ListExternalRepos`) that lists grants on repositories owned by **other users**, not only non-member organizations
- [ ] Share the filter logic with `repo.FilterNonMemberOrgRepos`; keep case-insensitive owner matching
- [ ] Document in `docs/guides/repo.md` and `docs/guides/cli.md`

**Status:** Open. `--non-member-orgs` deliberately excludes user-owned repositories,
so a direct collaborator grant on another person's repository is only visible in the
unfiltered listing.

**Files:** `repo/access.go`, `cmd/gogithub/cmd_repo_access.go`

### 4.3 Conditional Requests (ETag) for Polling Consumers

- [x] Add `etagcache` transport: `If-None-Match` on repeat `GET`s, `304` returned as the cached `200`
- [x] `MemoryStore` (LRU) for long-running processes, `FileStore` for CLI runs
- [x] `clientv1.ClientOptions.Transport` hook; `gogithub health --cache-dir`
- [ ] Consider a bounded `FileStore` (eviction by age or size) if cache directories grow

**Files:** `etagcache/`, `clientv1/client_impl.go`, `cmd/gogithub/cmd_health.go`

### 4.4 Health Package Extensions

- [ ] Open issue/PR counts by label or age (e.g. "stale > 30 days") — needs `ListIssues`/`ListPullRequests` rather than counts, so make it opt-in
- [ ] Latest release and tag per repository (`GetLatestRelease` is already in `clientv1`)
- [ ] Dependabot alert and code scanning alert counts (new `clientv1` methods; need `security_events` scope)
- [ ] Branch protection summary for the default branch (`GetBranchProtection` exists)
- [ ] GraphQL variant that batches issue and PR counts for ~25 repositories per request, for very large repository sets

**Status:** Ideas. Each adds requests per repository; keep the default `Collect` at four.

### 4.5 Repository Hygiene

- [ ] Decide whether `stats.svg` belongs in the library repository (see 4.1)
- [ ] Add a tracked-tree check for local filesystem paths (`/Users/`, `/home/`, `$HOME`) to the shared lint workflow

## Completed

- [x] v0.18.0 - Repository access listing, OAuth app authentication, repository health,
  conditional requests, `bulk_git_rm` cleanup
- [x] v0.12.1 release (2026-04-06) - Code quality improvements
  - Phase 1: Quick Wins (all complete)
  - Phase 2.1: go-github iterators (complete)
  - Phase 2.2: Error types review (complete)
  - Phase 3.3: Constants/error tests (partial)
- [x] v0.12.0 release (2026-04-06)

## Future Ideas

Potential improvements identified during refactoring (not prioritized):

- [ ] Add `context.Context` timeout helpers for long-running operations
- [ ] Consider exposing go-github iterators directly for streaming use cases
- [ ] Add retry logic with exponential backoff for rate-limited requests
- [ ] Create `examples/` directory with runnable code samples
- [ ] Add benchmarks for pagination-heavy operations
