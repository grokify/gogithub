# SVG Stats Card Generation

The `gogithub profile` command can generate SVG stats cards similar to [github-readme-stats](https://github.com/anuraghazra/github-readme-stats), but calculated from your actual contribution data.

## Quick Start

Generate an SVG stats card from the GitHub API:

```bash
gogithub profile --user YOUR_USERNAME --output-svg stats.svg
```

Or generate from an existing raw JSON file (no API calls needed):

```bash
gogithub profile --input profile.json --output-svg stats.svg
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--output-svg` | Output SVG file path | (none) |
| `--svg-theme` | Color theme name | `default` |
| `--svg-title` | Custom card title | `{username}'s GitHub Stats` |

## Available Themes

The following themes are available:

| Theme | Description |
|-------|-------------|
| `default` | Light theme with blue accents |
| `dark` | GitHub dark mode colors |
| `radical` | Vibrant pink and cyan on dark purple |
| `tokyonight` | Blue and teal on dark blue |
| `gruvbox` | Warm retro colors |
| `dracula` | Pink and purple on dark gray |
| `nord` | Cool blue tones |
| `catppuccin` | Soft pastel colors |

## Examples

### Basic usage

```bash
# Default theme
gogithub profile --user grokify --output-svg stats.svg

# Dark theme
gogithub profile --user grokify --output-svg stats.svg --svg-theme dark

# Custom title
gogithub profile --user grokify --output-svg stats.svg --svg-title "My GitHub Journey"
```

### Generate from existing JSON

If you've already fetched profile data, you can generate SVG without API calls:

```bash
# First, fetch and save raw data
gogithub profile --user grokify --output-raw profile.json

# Then generate SVG from the saved data
gogithub profile --input profile.json --output-svg stats.svg --svg-theme tokyonight
```

### Generate multiple outputs

```bash
# Generate JSON, README, and SVG together
gogithub profile --user grokify \
  --output-raw profile.json \
  --output-aggregate summary.json \
  --output-readme README.md \
  --output-svg stats.svg
```

## Stats Displayed

The SVG card shows the following statistics:

- **Total Commits** - All commits contributed (GitHub's official count)
- **Pull Requests** - PRs opened
- **Issues** - Issues created
- **Code Reviews** - PR reviews submitted
- **Repos Contributed To** - Number of repositories with contributions
- **Lines Changed** - Code additions and deletions with net change

## Using in GitHub Profile README

Add the SVG to your profile README:

```markdown
![GitHub Stats](./stats.svg)
```

## Automated Updates with GitHub Actions

You can automatically update your stats card using GitHub Actions. Copy the workflow template from:

```
.github/workflows/update-profile-readme.yml
```

### Setup Instructions

1. Copy the workflow file to your profile repository (username/username)
2. The workflow runs weekly by default, or you can trigger it manually
3. Reference the generated SVG in your README.md

### Workflow Features

- Runs on schedule (weekly) or manually
- Supports custom themes via workflow_dispatch inputs
- Only commits when there are actual changes
- Uses the built-in `GITHUB_TOKEN` for authentication

### Example Workflow

```yaml
name: Update Profile README

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  workflow_dispatch:
    inputs:
      theme:
        description: 'SVG theme'
        default: 'default'

jobs:
  update-readme:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install gogithub
        run: go install github.com/grokify/gogithub/cmd/gogithub@latest

      - name: Generate Stats SVG
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gogithub profile --user ${{ github.repository_owner }} \
            --output-svg stats.svg \
            --svg-theme ${{ github.event.inputs.theme || 'default' }}

      - name: Commit and Push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add stats.svg
          git diff --staged --quiet || git commit -m "chore: update profile stats"
          git push
```

## Programmatic Usage

The SVG generation is also available as a Go package:

```go
import (
    "github.com/grokify/gogithub/profile"
    "github.com/grokify/gogithub/profile/svg"
)

// Generate SVG from a UserProfile
func generateStatsCard(p *profile.UserProfile) string {
    return svg.GenerateSVG(p, "dark", "")
}

// With custom options
func generateCustomCard(p *profile.UserProfile) string {
    opts := svg.StatsCardOptions{
        Theme:        "tokyonight",
        Title:        "My GitHub Stats",
        ExcludeStats: []string{"Issues"},
        Width:        400,
    }
    sc := svg.NewStatsCardWithOptions(p, opts)
    return sc.Render()
}
```

## Customization

### Custom Colors

When using the Go package directly, you can override individual colors:

```go
opts := svg.StatsCardOptions{
    Theme:      "default",
    BgColor:    "#1a1b27",
    TitleColor: "#70a5fd",
    TextColor:  "#38bdae",
    IconColor:  "#bf91f3",
}
```

### Excluding Stats

To hide certain stats, use the `ExcludeStats` option:

```go
opts := svg.StatsCardOptions{
    ExcludeStats: []string{"Issues", "Code Reviews"},
}
```

## Comparison with github-readme-stats

| Feature | gogithub | github-readme-stats |
|---------|----------|---------------------|
| Data Source | Your actual contributions | GitHub API estimates |
| Lines of Code | Exact from commit history | Not available |
| Repos Contributed | Full count | Limited |
| Offline Mode | Yes (from saved JSON) | No |
| Self-hosted | Yes | Yes (Vercel) |
| Rate Limits | Standard GitHub API | Shared service |
