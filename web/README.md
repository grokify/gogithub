# GitHub Stats Viewer

A static web application that visualizes GitHub contribution data from `gogithub profile` JSON output.

## Features

- **Stats Card** - Embeddable SVG card showing commits, PRs, issues, reviews, and code changes
- **Contribution Calendar** - GitHub-style green heatmap grid with tooltips
- **Monthly Activity Charts** - Bar/line charts for activity and code change trends
- **Top Repositories** - Sortable table of repositories by commits/additions
- **Dark Mode** - Toggle between light and dark themes
- **Export** - Download stats card as SVG or PNG

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Usage

1. Generate profile data with gogithub:

   ```bash
   gogithub profile --user USERNAME --output-raw profile.json
   ```

2. Open the app in your browser

3. Drag and drop the JSON file (or click to select)

4. View your GitHub stats visualizations

## Demo Mode

Click the "Load Demo" button to see sample data without generating a profile.

## Tech Stack

- **Vite** - Fast build tool and dev server
- **TypeScript** - Type-safe JavaScript
- **Chart.js** - Monthly activity charts
- **Pure SVG** - Stats card and contribution calendar
- **CSS Variables** - Light/dark theme support

## Directory Structure

```
web/
├── src/
│   ├── main.ts                 # Entry point
│   ├── types.ts                # TypeScript interfaces
│   ├── components/
│   │   ├── FileLoader.ts       # Drag-drop file input
│   │   ├── StatsCard.ts        # SVG stats card generator
│   │   ├── Calendar.ts         # Contribution heatmap
│   │   ├── MonthlyChart.ts     # Chart.js charts
│   │   └── RepoList.ts         # Repository table
│   ├── utils/
│   │   ├── colors.ts           # Theme colors
│   │   ├── format.ts           # Number formatting
│   │   └── download.ts         # SVG/PNG export
│   └── styles/
│       └── main.css            # All styles
├── public/
│   └── sample-data.json        # Demo data
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Data Format

The app expects JSON output from `gogithub profile --output-raw`. Key fields:

| Field | Description |
|-------|-------------|
| `username` | GitHub username |
| `total_commits` | Total commit count |
| `total_prs` | Pull request count |
| `total_issues` | Issue count |
| `total_reviews` | Code review count |
| `total_additions` | Lines added |
| `total_deletions` | Lines deleted |
| `repos` | Array of repository details |
| `monthly` | Monthly breakdown |
| `calendar` | Contribution calendar data |
