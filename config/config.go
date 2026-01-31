// Package config provides configuration utilities for GitHub API clients.
package config

import (
	"context"
	"errors"
	"os"

	"github.com/google/go-github/v82/github"
	"golang.org/x/oauth2"
)

// Default values for GitHub configuration.
const (
	DefaultBranch    = "main"
	DefaultBaseURL   = "https://api.github.com/"
	DefaultUploadURL = "https://uploads.github.com/"
)

// Standard environment variable names.
const (
	EnvOwner     = "GITHUB_OWNER"
	EnvRepo      = "GITHUB_REPO"
	EnvBranch    = "GITHUB_BRANCH"
	EnvToken     = "GITHUB_TOKEN" //nolint:gosec // This is an env var name, not a credential
	EnvBaseURL   = "GITHUB_API_URL"
	EnvUploadURL = "GITHUB_UPLOAD_URL"
)

// Config errors.
var (
	ErrOwnerRequired = errors.New("owner is required")
	ErrRepoRequired  = errors.New("repo is required")
	ErrTokenRequired = errors.New("token is required")
)

// Config holds configuration for GitHub API operations.
type Config struct {
	// Owner is the repository owner (user or organization). Required.
	Owner string

	// Repo is the repository name. Required.
	Repo string

	// Branch is the branch to operate on. Default: "main".
	Branch string

	// Token is the GitHub personal access token. Required.
	// Needs "repo" scope for private repos, or "public_repo" for public repos.
	Token string

	// BaseURL is the GitHub API base URL. Default: "https://api.github.com/".
	// Set this for GitHub Enterprise (e.g., "https://github.example.com/api/v3/").
	BaseURL string

	// UploadURL is the GitHub upload URL. Default: "https://uploads.github.com/".
	// Set this for GitHub Enterprise.
	UploadURL string
}

// Default returns a Config with default values set.
func Default() Config {
	return Config{
		Branch:    DefaultBranch,
		BaseURL:   DefaultBaseURL,
		UploadURL: DefaultUploadURL,
	}
}

// Validate checks if the configuration has all required fields.
func (c *Config) Validate() error {
	if c.Owner == "" {
		return ErrOwnerRequired
	}
	if c.Repo == "" {
		return ErrRepoRequired
	}
	if c.Token == "" {
		return ErrTokenRequired
	}
	return nil
}

// ApplyDefaults fills in default values for empty fields.
func (c *Config) ApplyDefaults() {
	if c.Branch == "" {
		c.Branch = DefaultBranch
	}
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.UploadURL == "" {
		c.UploadURL = DefaultUploadURL
	}
}

// IsEnterprise returns true if the config is for GitHub Enterprise.
func (c *Config) IsEnterprise() bool {
	return c.BaseURL != "" && c.BaseURL != DefaultBaseURL
}

// FromMap creates a Config from a string map.
// Supported keys:
//   - owner: repository owner (required)
//   - repo: repository name (required)
//   - branch: branch name (default: "main")
//   - token: GitHub personal access token (required)
//   - base_url: GitHub API base URL (for GitHub Enterprise)
//   - upload_url: GitHub upload URL (for GitHub Enterprise)
func FromMap(m map[string]string) Config {
	cfg := Default()

	if v, ok := m["owner"]; ok {
		cfg.Owner = v
	}
	if v, ok := m["repo"]; ok {
		cfg.Repo = v
	}
	if v, ok := m["branch"]; ok && v != "" {
		cfg.Branch = v
	}
	if v, ok := m["token"]; ok {
		cfg.Token = v
	}
	if v, ok := m["base_url"]; ok && v != "" {
		cfg.BaseURL = v
	}
	if v, ok := m["upload_url"]; ok && v != "" {
		cfg.UploadURL = v
	}

	return cfg
}

// EnvConfig holds environment variable names for configuration.
// Use this to customize which environment variables are used.
type EnvConfig struct {
	Owner     string
	Repo      string
	Branch    string
	Token     string
	BaseURL   string
	UploadURL string
}

// DefaultEnvConfig returns the default environment variable names.
func DefaultEnvConfig() EnvConfig {
	return EnvConfig{
		Owner:     EnvOwner,
		Repo:      EnvRepo,
		Branch:    EnvBranch,
		Token:     EnvToken,
		BaseURL:   EnvBaseURL,
		UploadURL: EnvUploadURL,
	}
}

// FromEnv creates a Config from environment variables using default names.
func FromEnv() Config {
	return FromEnvWithConfig(DefaultEnvConfig())
}

// FromEnvWithConfig creates a Config from environment variables with custom names.
func FromEnvWithConfig(envCfg EnvConfig) Config {
	cfg := Default()

	if v := os.Getenv(envCfg.Owner); v != "" {
		cfg.Owner = v
	}
	if v := os.Getenv(envCfg.Repo); v != "" {
		cfg.Repo = v
	}
	if v := os.Getenv(envCfg.Branch); v != "" {
		cfg.Branch = v
	}
	if v := os.Getenv(envCfg.Token); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv(envCfg.BaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(envCfg.UploadURL); v != "" {
		cfg.UploadURL = v
	}

	return cfg
}

// FromEnvWithFallback creates a Config from environment variables,
// checking primary env var names first, then fallback names.
func FromEnvWithFallback(primary, fallback EnvConfig) Config {
	cfg := Default()

	// Owner
	if v := os.Getenv(primary.Owner); v != "" {
		cfg.Owner = v
	} else if v := os.Getenv(fallback.Owner); v != "" {
		cfg.Owner = v
	}

	// Repo
	if v := os.Getenv(primary.Repo); v != "" {
		cfg.Repo = v
	} else if v := os.Getenv(fallback.Repo); v != "" {
		cfg.Repo = v
	}

	// Branch
	if v := os.Getenv(primary.Branch); v != "" {
		cfg.Branch = v
	} else if v := os.Getenv(fallback.Branch); v != "" {
		cfg.Branch = v
	}

	// Token
	if v := os.Getenv(primary.Token); v != "" {
		cfg.Token = v
	} else if v := os.Getenv(fallback.Token); v != "" {
		cfg.Token = v
	}

	// BaseURL
	if v := os.Getenv(primary.BaseURL); v != "" {
		cfg.BaseURL = v
	} else if v := os.Getenv(fallback.BaseURL); v != "" {
		cfg.BaseURL = v
	}

	// UploadURL
	if v := os.Getenv(primary.UploadURL); v != "" {
		cfg.UploadURL = v
	} else if v := os.Getenv(fallback.UploadURL); v != "" {
		cfg.UploadURL = v
	}

	return cfg
}

// NewClient creates a GitHub client from the configuration.
// The config must be validated before calling this function.
func (c *Config) NewClient(ctx context.Context) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	if c.IsEnterprise() {
		return github.NewClient(tc).WithEnterpriseURLs(c.BaseURL, c.UploadURL)
	}

	return github.NewClient(tc), nil
}

// MustNewClient creates a GitHub client from the configuration.
// It panics if the config is invalid or client creation fails.
func (c *Config) MustNewClient(ctx context.Context) *github.Client {
	if err := c.Validate(); err != nil {
		panic(err)
	}

	client, err := c.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	return client
}
