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

### 3.2 Split Repo Package
- [ ] Create `repo/batch/` subpackage for batch operations
- [ ] Create `repo/content/` subpackage for file content
- [ ] Create `repo/stats/` subpackage for contributor statistics
- [ ] Maintain backward compatibility with re-exports
- [ ] Update documentation

**Files:** `repo/*.go`

### 3.3 Add Missing Test Coverage
- [ ] Add tests for `release/release.go`
- [ ] Add tests for `tag/tag.go`
- [ ] Add tests for `graphql/*.go`
- [ ] Add tests for `repo/fork.go`
- [ ] Add tests for `repo/branch.go`
- [ ] Add tests for `repo/list.go`

## Completed

- [x] v0.12.0 release (2026-04-06)
