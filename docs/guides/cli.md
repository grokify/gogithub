# Command Line Interface

GoGitHub includes a CLI tool for common GitHub operations without writing code.

## Installation

```bash
go install github.com/grokify/gogithub/cmd/gogithub@latest
```

Or build from source:

```bash
git clone https://github.com/grokify/gogithub.git
cd gogithub
go build ./cmd/gogithub
```

## Authentication

Set the `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=your-token
```

For public data, use a fine-grained token with "Public Repositories (read-only)" access and no additional permissions. See [Authentication](auth.md#token-requirements-by-use-case) for details.

## Commands

### profile

Fetch comprehensive user contribution statistics.

```bash
gogithub profile --user <username> [flags]
```

#### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--user` | `-u` | GitHub username | (required) |
| `--from` | `-f` | Start date (YYYY-MM-DD) | 1 year ago |
| `--to` | `-t` | End date (YYYY-MM-DD) | today |
| `--format` | | Output format: `summary`, `json` | `summary` |
| `--output` | `-o` | Output file (stdout if not specified) | |
| `--output-raw` | | Output raw JSON file with all data | |
| `--output-aggregate` | | Output aggregate JSON file | |
| `--input` | `-i` | Input raw JSON file (skip API calls) | |
| `--include-releases` | | Fetch release counts for contributed repos | `false` |

#### Examples

**Human-readable summary:**

```bash
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31
```

Output:

```
Fetching profile for 'grokify' from 2024-01-01 to 2024-12-31

[1/4] Fetching contribution statistics   [████████████████████] 100%
[2/4] Fetching commit details            [████████████████████] 100%
[3/4] Processing repositories            [████████████████████] 100%
[4/4] Building activity timeline         [████████████████████] 100%

=== Profile: grokify ===
Period: 2024-01-01 to 2024-12-31

Contributions (GitHub official):
  Commits:       1125
  Pull Requests: 45
  Issues:        12
  Reviews:       30
  Repos Created: 5

Code Changes (from default branch history):
  Commits:   1081
  Additions: +738054
  Deletions: -294379
  Net:       +443675

Repositories Contributed To: 71

Activity:
  Days with contributions: 280
  Longest streak:          45 days
  Current streak:          12 days

Top Repositories by Commits:
  1. grokify/mogo: 150 commits (+25000/-5000)
  2. grokify/gogithub: 120 commits (+18000/-3000)
  ...
```

**JSON output:**

```bash
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 --format json
```

**Save to file:**

```bash
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 -o profile.txt
```

**Generate both raw and aggregate JSON:**

```bash
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 \
    --output-raw raw.json --output-aggregate aggregate.json
```

The raw JSON contains all per-repository data and can be used to regenerate aggregates without making API calls:

```bash
gogithub profile --input raw.json --output aggregate.json
```

**Include release counts:**

```bash
gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 \
    --include-releases --output-raw raw.json
```

This fetches release data for all contributed repositories and aggregates by month based on each release's publish date. See [Release Data](#release-data) for details on API overhead.

#### Output Formats

**Summary** (default): Human-readable text with sections for contributions, code changes, activity streaks, top repositories, and monthly breakdown.

**JSON**: Structured data including:

```json
{
  "username": "grokify",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-12-31T23:59:59Z",
  "total_commits": 1125,
  "commits_default_branch": 1081,
  "total_prs": 45,
  "total_issues": 12,
  "total_reviews": 30,
  "total_additions": 738054,
  "total_deletions": 294379,
  "net_additions": 443675,
  "total_releases": 8,
  "repos_contributed_to": 71,
  "calendar": {
    "total_contributions": 1500,
    "days_with_contributions": 280,
    "longest_streak": 45,
    "current_streak": 12
  },
  "monthly": [
    {
      "year": 2024,
      "month": 1,
      "month_name": "January",
      "commits": 95,
      "issues": 2,
      "prs": 5,
      "reviews": 3,
      "releases": 2,
      "additions": 50000,
      "deletions": 20000
    }
  ]
}
```

Note: `total_releases` and monthly `releases` are only populated when using `--include-releases`.

**Raw JSON** (`--output-raw`): Complete data including per-repository details and full calendar data. Use this for archival or to regenerate aggregates later.

#### Commit Count Clarification

The output shows two commit counts:

| Field | Description |
|-------|-------------|
| `total_commits` | GitHub's official count (shown on profile page) |
| `commits_default_branch` | Commits found traversing default branch history |

These may differ because `total_commits` includes all branches while `commits_default_branch` only traverses default branches (but provides additions/deletions data).

#### Release Data

The `--include-releases` flag fetches release counts for all repositories you contributed to during the specified period. Releases are aggregated by month based on their `published_at` date.

**API overhead:**

| Without `--include-releases` | With `--include-releases` |
|------------------------------|---------------------------|
| 2-4 GraphQL calls total | 2-4 GraphQL calls + 1 REST call per repository |

For example, if you contributed to 71 repositories, enabling releases adds 71+ additional API calls. Each repository with more than 100 releases requires additional paginated calls.

**Recommendations:**

- Use `--output-raw` to cache the data and avoid repeated API calls
- Regenerate aggregates from cached raw data using `--input` (no API calls needed)
- Only enable `--include-releases` when you need release statistics

### search-prs

Search for open pull requests by user.

```bash
gogithub search-prs --accounts <users> [flags]
```

#### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--accounts` | `-a` | GitHub accounts to search (comma-separated) | (required) |
| `--outfile` | `-o` | Output Excel file | `githubissues.xlsx` |

#### Examples

```bash
# Search for PRs by multiple users
gogithub search-prs --accounts grokify,octocat --outfile prs.xlsx

# Single user
gogithub search-prs -a grokify
```

## Progress Display

Long-running commands show real-time progress with:

- Stage indicators: `[1/4]`, `[2/4]`, etc.
- Visual progress bar with Unicode characters
- Percentage completion

```
[1/4] Fetching contribution statistics   [████████████████████] 100%
[2/4] Fetching commit details            [████████████████████] 100%
[3/4] Processing repositories            [██████████░░░░░░░░░░]  50%
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (missing token, API error, invalid arguments) |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub personal access token (required for API calls) |

## Tips

1. **Offline aggregate generation**: Fetch raw data once, then regenerate aggregates without API calls:

   ```bash
   # Fetch once (makes API calls)
   gogithub profile --user grokify --from 2024-01-01 --to 2024-12-31 \
       --include-releases --output-raw data.json

   # Regenerate anytime (no API calls)
   gogithub profile --input data.json --format json
   ```

2. **Date ranges**: Both `--from` and `--to` dates are inclusive.

3. **Large date ranges**: For ranges over 1 year, the tool automatically splits queries to work within GitHub's API limits.

4. **Rate limiting**: The tool respects GitHub's rate limits. For large queries, consider using a token with higher limits.

5. **Release data caching**: Since `--include-releases` adds significant API overhead, always use `--output-raw` to cache the results. The raw JSON preserves all release data for future aggregate generation.
