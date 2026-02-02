# Testing

GoGitHub includes both unit tests and integration tests. Unit tests run without external dependencies, while integration tests make real API calls to GitHub.

## Running Tests

### Unit Tests Only

Unit tests run without any configuration:

```bash
go test ./...
```

### Integration Tests

Integration tests require a GitHub token. They automatically skip when `GITHUB_TOKEN` is not set:

```bash
# Integration tests will skip
go test ./... -v

# Integration tests will run
GITHUB_TOKEN=your-token go test ./... -v
```

### Verbose Output

Use `-v` to see detailed test output, including integration test results:

```bash
GITHUB_TOKEN=your-token go test ./profile/... ./repo/... -v
```

### Run Specific Tests

```bash
# Run only integration tests
GITHUB_TOKEN=your-token go test ./... -v -run Integration

# Run only unit tests (skip integration)
go test ./... -v -skip Integration
```

## Token Requirements

Integration tests only need read access to public data. See [Authentication](auth.md#token-requirements-by-use-case) for token setup.

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Repository access: "Public Repositories (read-only)", no permissions |
| Classic PAT | No scopes required |

## Test Structure

Tests are organized by package with a consistent naming convention:

| File Pattern | Description |
|--------------|-------------|
| `*_test.go` | Unit tests (no external dependencies) |
| `*_integration_test.go` | Integration tests (require `GITHUB_TOKEN`) |

### Integration Test Pattern

All integration tests follow this pattern to skip when no token is available:

```go
func getTestToken(t *testing.T) string {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set, skipping integration test")
    }
    return token
}

func TestSomethingIntegration(t *testing.T) {
    token := getTestToken(t)
    // ... test code using token
}
```

## Available Integration Tests

### Profile Package

Tests in `profile/profile_integration_test.go`:

| Test | Description |
|------|-------------|
| `TestGetUserProfileIntegration` | Fetch profile for a real user |
| `TestGetUserProfileWithOptionsIntegration` | Fetch with visibility filter |
| `TestGetUserProfileCalendarIntegration` | Verify calendar and streak data |
| `TestGetUserProfileActivityIntegration` | Verify monthly activity timeline |
| `TestGetUserProfileWithReleasesIntegration` | Fetch with release counting |
| `TestGetUserProfileTopReposIntegration` | Verify top repos sorting |
| `TestGetUserProfileSummaryIntegration` | Verify summary output |
| `TestGetUserProfileInvalidUserIntegration` | Error handling for invalid user |

### Repo Package

Tests in `repo/contributors_integration_test.go`:

| Test | Description |
|------|-------------|
| `TestListContributorStatsIntegration` | List contributors for a repo |
| `TestGetContributorStatsIntegration` | Get stats for specific contributor |
| `TestGetContributorSummaryIntegration` | Get summarized contributor stats |
| `TestGetContributorStatsNotFoundIntegration` | Handle non-contributor lookup |
| `TestListContributorStatsNonExistentRepoIntegration` | Error handling for invalid repo |
| `TestListContributorStatsLargeRepoIntegration` | Test with large public repo |

## Writing New Tests

### Unit Tests

For pure logic that doesn't require API calls:

```go
func TestSomething(t *testing.T) {
    input := &SomeType{Field: "value"}
    result := input.SomeMethod()

    if result != expected {
        t.Errorf("SomeMethod() = %v, want %v", result, expected)
    }
}
```

### Integration Tests

For tests that require real API calls:

```go
func TestSomethingIntegration(t *testing.T) {
    token := getTestToken(t)
    ctx := context.Background()

    client := github.NewClient(nil).WithAuthToken(token)

    result, err := SomeAPICall(ctx, client, "owner", "repo")
    if err != nil {
        t.Fatalf("SomeAPICall failed: %v", err)
    }

    // Log results for manual verification
    t.Logf("Result: %+v", result)

    // Assert on structure, not specific values (real data changes)
    if result.Field == "" {
        t.Error("Field should not be empty")
    }
}
```

### Best Practices

1. **Name integration tests with `Integration` suffix** - Makes it easy to run or skip them selectively

2. **Use `t.Logf` for informational output** - Helps verify correct behavior when running with `-v`

3. **Don't assert on specific values** - Real GitHub data changes; assert on structure and non-empty values

4. **Use well-known public repos** - Tests use `grokify/gogithub` and `google/go-github` as stable test fixtures

5. **Handle 202 Accepted gracefully** - GitHub may return 202 while computing stats; the `repo` package handles retries automatically

## CI Considerations

In CI environments, you can:

1. **Run unit tests always** - No configuration needed

2. **Run integration tests conditionally** - Only when `GITHUB_TOKEN` secret is available

```yaml
# GitHub Actions example
- name: Run unit tests
  run: go test ./... -skip Integration

- name: Run integration tests
  if: ${{ secrets.GITHUB_TOKEN != '' }}
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: go test ./... -v -run Integration
```

## Linting

Run the linter before committing:

```bash
golangci-lint run
```

All test files should pass linting with zero issues.
