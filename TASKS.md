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
- [ ] Add field-level docs to `profile.Options` struct
- [ ] Add usage examples to `repo/batch.go` doc comments
- [ ] Document `search.Must*` methods' silent failure behavior

## Phase 2: High Value (Medium Effort)

### 2.1 Create Generic Pagination Helper
- [ ] Create `internal/pagination/pagination.go` with generic paginator
- [ ] Update `release/release.go` to use paginator
- [ ] Update `checks/checks.go` to use paginator
- [ ] Update `tag/tag.go` to use paginator
- [ ] Add unit tests for pagination helper

**Files:** `release/release.go`, `checks/checks.go`, `tag/tag.go`

### 2.2 Standardize Error Types
- [ ] Create base `OperationError` type in `errors/errors.go`
- [ ] Update `AuthError` to use standard pattern
- [ ] Update `PRError` to use standard pattern
- [ ] Update `CommitError`, `BatchError`, `BranchError`, `ForkError`
- [ ] Ensure all errors implement `Unwrap()` for Go 1.13+ chains

**Files:** `errors/errors.go`, `auth/auth.go`, `pr/pullrequest.go`, `repo/*.go`

### 2.3 Extract Profile CLI Logic to Library
- [ ] Create `profile/converter/converter.go` for JSON/struct conversions
- [ ] Move `profileToRaw()`, `profileToAggregate()`, `rawToAggregate()`, `rawToProfile()` from CLI
- [ ] Create `profile/output/writer.go` for file writing operations
- [ ] Simplify `cmd/gogithub/cmd_profile.go` to orchestrate components
- [ ] Add unit tests for extracted functions

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
