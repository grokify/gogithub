# Authentication

GoGitHub provides flexible authentication options for both GitHub.com and GitHub Enterprise.

## REST API Client

### Token Authentication

The simplest way to authenticate is with a personal access token:

```go
import "github.com/grokify/gogithub/auth"

ctx := context.Background()
gh := auth.NewGitHubClient(ctx, "your-github-token")
```

### HTTP Client

If you need the underlying HTTP client (e.g., for custom transports):

```go
httpClient := auth.NewTokenClient(ctx, "your-github-token")
```

### Get Authenticated User

Verify authentication and get the current user:

```go
username, err := auth.GetAuthenticatedUser(ctx, gh)
if err != nil {
    // Handle authentication error
}
fmt.Printf("Authenticated as: %s\n", username)
```

### Get User Information

Retrieve information about any user:

```go
user, err := auth.GetUser(ctx, gh, "octocat")
if err != nil {
    // Handle error
}
fmt.Printf("Name: %s\n", user.GetName())
```

## Configuration

The `config` package provides a structured way to manage GitHub configuration:

### From Environment Variables

```go
import "github.com/grokify/gogithub/config"

cfg, err := config.FromEnv()
if err != nil {
    panic(err)
}

gh, err := cfg.NewClient(ctx)
if err != nil {
    panic(err)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_TOKEN` | Personal access token | (required) |
| `GITHUB_OWNER` | Default repository owner | - |
| `GITHUB_REPO` | Default repository name | - |
| `GITHUB_BRANCH` | Default branch | `main` |
| `GITHUB_BASE_URL` | API base URL | `https://api.github.com/` |
| `GITHUB_UPLOAD_URL` | Upload URL | `https://uploads.github.com/` |

### Manual Configuration

```go
cfg := &config.Config{
    Token:  "your-token",
    Owner:  "grokify",
    Repo:   "gogithub",
    Branch: "main",
}

if err := cfg.Validate(); err != nil {
    panic(err)
}

gh, err := cfg.NewClient(ctx)
```

### GitHub Enterprise

For GitHub Enterprise Server, specify custom URLs:

```go
cfg := &config.Config{
    Token:     "your-token",
    BaseURL:   "https://github.mycompany.com/api/v3",
    UploadURL: "https://github.mycompany.com/api/uploads",
}

gh, err := cfg.NewClient(ctx)
```

Check if a configuration is for GitHub Enterprise:

```go
if cfg.IsEnterprise() {
    fmt.Println("Using GitHub Enterprise")
}
```

## GraphQL API Client

For the GraphQL API (used for contribution statistics), use a separate client:

```go
import "github.com/grokify/gogithub/graphql"

client := graphql.NewClient(ctx, "your-github-token")
```

For GitHub Enterprise:

```go
client := graphql.NewEnterpriseClient(ctx, "your-token", "https://github.mycompany.com/api/graphql")
```

!!! note "GraphQL Requires Authentication"
    Unlike the REST API which allows limited unauthenticated access, the GraphQL API always requires a token. Even for public data, you need a token (with no special scopes).

## Token Scopes

### Basic Scopes

| Scope | Required For |
|-------|--------------|
| (none) | Read public data, GraphQL queries on public repos |
| `public_repo` | Write to public repositories |
| `repo` | Full access to private repositories |
| `read:user` | Read user profile data, contribution statistics |
| `read:org` | Read organization membership |

### Token Requirements by Use Case

#### Public Data Only (Recommended Starting Point)

For reading public user statistics, contribution data, and public repository information, you need **minimal permissions**:

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Repository access: "Public Repositories (read-only)", no other permissions |
| Classic PAT | No scopes required (just a valid token) |
| GitHub CLI | Default `gh auth login` works |

This is sufficient for:

- User contribution statistics (`profile`, `graphql` packages)
- Public repository data (commits, releases, contributors)
- Search across public repositories

#### Private Repository Access

For accessing private repositories, you need additional permissions:

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Repository access: select specific repos, Contents: Read |
| Classic PAT | `repo` scope |

#### Write Operations

For creating commits, PRs, releases, etc.:

| Token Type | Configuration |
|------------|---------------|
| Fine-grained PAT | Repository access: select repos, Contents: Read and write |
| Classic PAT | `repo` (private) or `public_repo` (public only) |

### Creating a Token

=== "Fine-Grained Token (Recommended)"

    Fine-grained tokens provide the most secure, minimal-permission approach.

    **For public data only:**

    1. Go to [GitHub Settings > Fine-grained tokens](https://github.com/settings/tokens?type=beta)
    2. Click "Generate new token"
    3. Set **Repository access** to "Public Repositories (read-only)"
    4. Leave all **Permissions** sections empty (no permissions needed)
    5. Click "Generate token" and copy it

    **For private repository access:**

    1. Go to [GitHub Settings > Fine-grained tokens](https://github.com/settings/tokens?type=beta)
    2. Click "Generate new token"
    3. Set **Repository access** to "Only select repositories" and choose your repos
    4. Under **Repository permissions**, set "Contents" to "Read-only"
    5. Click "Generate token" and copy it

=== "GitHub CLI"

    ```bash
    # Login with default scopes (sufficient for most read operations)
    gh auth login

    # Get current token for use in your code
    gh auth token
    ```

=== "Classic PAT"

    1. Go to [GitHub Settings > Personal access tokens (classic)](https://github.com/settings/tokens)
    2. Click "Generate new token (classic)"
    3. For public data: no scopes needed
    4. For private repos: select `repo` scope
    5. Copy the token

!!! tip "Start Minimal"
    Start with a fine-grained token with "Public Repositories (read-only)" access and no permissions. This is sufficient for most read operations on public data. Add permissions only when needed.

## Bot User Detection

GoGitHub provides constants for known bot users:

```go
import "github.com/grokify/gogithub/auth"

if issue.User.GetLogin() == auth.UsernameDependabot {
    fmt.Println("This is a Dependabot PR")
}

// Or check by ID
if issue.User.GetID() == auth.UserIDDependabot {
    fmt.Println("This is a Dependabot PR")
}
```

## API Reference

See [pkg.go.dev/github.com/grokify/gogithub/auth](https://pkg.go.dev/github.com/grokify/gogithub/auth) for complete API documentation.
