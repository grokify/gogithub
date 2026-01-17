# Changelog

This page summarizes changes for each release. For detailed release notes, see the [GitHub releases page](https://github.com/grokify/gogithub/releases).

## [v0.6.0] - 2026-01-17

### Highlights

- Added `graphql` package for GitHub GraphQL API support with user contribution statistics
- Added MkDocs documentation site with Material theme, auto-deployed to GitHub Pages
- Two approaches for commit statistics: quick stats (profile view) and detailed stats (additions/deletions)

### Added

- New `graphql` package for GitHub GraphQL API
- `GetContributionStats` - Quick contribution statistics (like GitHub profile)
- `GetCommitStats` - Detailed commit stats with additions/deletions
- `GetCommitStatsByVisibility` - Stats split by public/private repositories
- `Visibility` type with `VisibilityAll`, `VisibilityPublic`, `VisibilityPrivate`
- MkDocs documentation site at [grokify.github.io/gogithub](https://grokify.github.io/gogithub/)
- Documentation guides for all packages
- GitHub Actions workflow for automatic gh-pages deployment

### Dependencies

- Added `github.com/shurcooL/githubv4` for GitHub GraphQL API support

## [v0.5.0] - 2026-01-17

### Added

- `QueryBuilder` for type-safe search query construction
- Fluent interface methods: `User()`, `Org()`, `Repo()`, `State()`, `Is()`, `Author()`, `Assignee()`, `Label()`, `Mentions()`, `Involves()`

### Changed

- Search package now supports both `Query` map and `QueryBuilder` approaches

## [v0.4.0] - 2026-01-17

### Added

- `errors` package with typed API errors
- Helper functions: `IsNotFound()`, `IsRateLimited()`, `IsPermissionDenied()`, etc.
- `pathutil` package for path validation and normalization
- `repo.Batch` for atomic multi-file commits

### Changed

- Improved error handling across all packages
- Better path validation for commit operations

## [v0.3.0] - 2026-01-17

### Added

- Package restructuring by operation type
- `auth` package for authentication utilities
- `config` package for environment-based configuration
- `search` package for Search API operations
- `repo` package for repository operations
- `pr` package for pull request operations
- `release` package for release operations
- GitHub Enterprise support

### Changed

- Root package now provides backward-compatible re-exports
- Migrated to go-github v81

## Links

- [Full Changelog](https://github.com/grokify/gogithub/blob/main/CHANGELOG.md)
- [Release Notes](https://github.com/grokify/gogithub/releases)
- [Roadmap](https://github.com/grokify/gogithub/blob/main/ROADMAP.md)
