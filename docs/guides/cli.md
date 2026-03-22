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
| `--output-monthly` | | Output monthly JSON file (merges with existing) | |
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

**Generate monthly JSON with auto-merge:**

```bash
# First fetch - creates the file
gogithub profile --user grokify --from 2024-01-01 --to 2024-01-31 \
    --output-monthly monthly.json

# Later - add more months (automatically merges with existing data)
gogithub profile --user grokify --from 2024-02-01 --to 2024-03-31 \
    --output-monthly monthly.json
```

The `--output-monthly` flag:

- Creates a new file if it doesn't exist
- Merges with existing data if the file exists (new data overwrites same month)
- Keeps months sorted in descending chronological order (newest first)
- Outputs a focused format with just username, timestamp, and monthly array

This is useful for incrementally building a history of contributions over time.

#### Output Formats

**Summary** (default): Human-readable text with sections for contributions, code changes, activity streaks, top repositories, and monthly breakdown.

**JSON**: Structured data including:

```json
{
  "username": "grokify",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-12-31T23:59:59Z",
  "generatedAt": "2024-12-31T12:00:00Z",
  "totalCommits": 1125,
  "commitsDefaultBranch": 1081,
  "totalPrs": 45,
  "totalIssues": 12,
  "totalReviews": 30,
  "totalAdditions": 738054,
  "totalDeletions": 294379,
  "netAdditions": 443675,
  "totalReleases": 8,
  "reposContributedTo": 71,
  "calendar": {
    "totalContributions": 1500,
    "daysWithContributions": 280,
    "longestStreak": 45,
    "currentStreak": 12
  },
  "monthly": [
    {
      "year": 2024,
      "month": 1,
      "monthName": "January",
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

Note: `totalReleases` and monthly `releases` are only populated when using `--include-releases`.

**Monthly JSON** (`--output-monthly`): Focused format for tracking monthly contributions over time:

```json
{
  "username": "grokify",
  "generatedAt": "2024-12-31T12:00:00Z",
  "months": [
    {
      "year": 2024,
      "month": 3,
      "monthName": "March",
      "commits": 120,
      "issues": 5,
      "prs": 10,
      "reviews": 15,
      "releases": 1,
      "additions": 8000,
      "deletions": 3000
    },
    {
      "year": 2024,
      "month": 2,
      "monthName": "February",
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

Months are sorted in descending order (newest first). When merging, existing months are updated with new data.

**Raw JSON** (`--output-raw`): Complete data including per-repository details and full calendar data. Use this for archival or to regenerate aggregates later.

#### Commit Count Clarification

The output shows two commit counts:

| Field | Description |
|-------|-------------|
| `totalCommits` | GitHub's official count (shown on profile page) |
| `commitsDefaultBranch` | Commits found traversing default branch history |

These may differ because `totalCommits` includes all branches while `commitsDefaultBranch` only traverses default branches (but provides additions/deletions data).

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

6. **Incremental monthly tracking**: Use `--output-monthly` to build a running history of contributions:

   ```bash
   # Run monthly to accumulate data
   gogithub profile --user grokify --from 2024-03-01 --to 2024-03-31 \
       --output-monthly ~/contributions.json
   ```

   The file automatically merges new months with existing data, keeping everything sorted chronologically.
