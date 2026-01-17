package config

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Branch != DefaultBranch {
		t.Errorf("Branch = %q, want %q", cfg.Branch, DefaultBranch)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.UploadURL != DefaultUploadURL {
		t.Errorf("UploadURL = %q, want %q", cfg.UploadURL, DefaultUploadURL)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid",
			config: Config{
				Owner: "owner",
				Repo:  "repo",
				Token: "token",
			},
			wantErr: nil,
		},
		{
			name: "missing owner",
			config: Config{
				Repo:  "repo",
				Token: "token",
			},
			wantErr: ErrOwnerRequired,
		},
		{
			name: "missing repo",
			config: Config{
				Owner: "owner",
				Token: "token",
			},
			wantErr: ErrRepoRequired,
		},
		{
			name: "missing token",
			config: Config{
				Owner: "owner",
				Repo:  "repo",
			},
			wantErr: ErrTokenRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := Config{
		Owner: "owner",
		Repo:  "repo",
		Token: "token",
	}

	cfg.ApplyDefaults()

	if cfg.Branch != DefaultBranch {
		t.Errorf("Branch = %q, want %q", cfg.Branch, DefaultBranch)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.UploadURL != DefaultUploadURL {
		t.Errorf("UploadURL = %q, want %q", cfg.UploadURL, DefaultUploadURL)
	}
}

func TestApplyDefaultsPreservesValues(t *testing.T) {
	cfg := Config{
		Owner:     "owner",
		Repo:      "repo",
		Token:     "token",
		Branch:    "develop",
		BaseURL:   "https://enterprise.example.com/api/v3/",
		UploadURL: "https://enterprise.example.com/uploads/",
	}

	cfg.ApplyDefaults()

	if cfg.Branch != "develop" {
		t.Errorf("Branch = %q, want %q", cfg.Branch, "develop")
	}
	if cfg.BaseURL != "https://enterprise.example.com/api/v3/" {
		t.Errorf("BaseURL was overwritten")
	}
	if cfg.UploadURL != "https://enterprise.example.com/uploads/" {
		t.Errorf("UploadURL was overwritten")
	}
}

func TestIsEnterprise(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    bool
	}{
		{"empty", "", false},
		{"default", DefaultBaseURL, false},
		{"enterprise", "https://enterprise.example.com/api/v3/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{BaseURL: tt.baseURL}
			if got := cfg.IsEnterprise(); got != tt.want {
				t.Errorf("IsEnterprise() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromMap(t *testing.T) {
	m := map[string]string{
		"owner":      "test-owner",
		"repo":       "test-repo",
		"branch":     "develop",
		"token":      "test-token",
		"base_url":   "https://enterprise.example.com/api/v3/",
		"upload_url": "https://enterprise.example.com/uploads/",
	}

	cfg := FromMap(m)

	if cfg.Owner != "test-owner" {
		t.Errorf("Owner = %q, want %q", cfg.Owner, "test-owner")
	}
	if cfg.Repo != "test-repo" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "test-repo")
	}
	if cfg.Branch != "develop" {
		t.Errorf("Branch = %q, want %q", cfg.Branch, "develop")
	}
	if cfg.Token != "test-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "test-token")
	}
	if cfg.BaseURL != "https://enterprise.example.com/api/v3/" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://enterprise.example.com/api/v3/")
	}
	if cfg.UploadURL != "https://enterprise.example.com/uploads/" {
		t.Errorf("UploadURL = %q, want %q", cfg.UploadURL, "https://enterprise.example.com/uploads/")
	}
}

func TestFromMapDefaults(t *testing.T) {
	m := map[string]string{
		"owner": "test-owner",
		"repo":  "test-repo",
		"token": "test-token",
	}

	cfg := FromMap(m)

	if cfg.Branch != DefaultBranch {
		t.Errorf("Branch = %q, want %q", cfg.Branch, DefaultBranch)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.UploadURL != DefaultUploadURL {
		t.Errorf("UploadURL = %q, want %q", cfg.UploadURL, DefaultUploadURL)
	}
}

func TestFromMapEmptyValues(t *testing.T) {
	m := map[string]string{
		"owner":  "test-owner",
		"repo":   "test-repo",
		"token":  "test-token",
		"branch": "", // Empty should use default
	}

	cfg := FromMap(m)

	if cfg.Branch != DefaultBranch {
		t.Errorf("Branch = %q, want %q (empty string should use default)", cfg.Branch, DefaultBranch)
	}
}

func TestFromEnv(t *testing.T) {
	// Save and restore env vars
	defer func() {
		os.Unsetenv(EnvOwner)
		os.Unsetenv(EnvRepo)
		os.Unsetenv(EnvBranch)
		os.Unsetenv(EnvToken)
		os.Unsetenv(EnvBaseURL)
		os.Unsetenv(EnvUploadURL)
	}()

	os.Setenv(EnvOwner, "env-owner")
	os.Setenv(EnvRepo, "env-repo")
	os.Setenv(EnvBranch, "env-branch")
	os.Setenv(EnvToken, "env-token")
	os.Setenv(EnvBaseURL, "https://enterprise.example.com/api/v3/")
	os.Setenv(EnvUploadURL, "https://enterprise.example.com/uploads/")

	cfg := FromEnv()

	if cfg.Owner != "env-owner" {
		t.Errorf("Owner = %q, want %q", cfg.Owner, "env-owner")
	}
	if cfg.Repo != "env-repo" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "env-repo")
	}
	if cfg.Branch != "env-branch" {
		t.Errorf("Branch = %q, want %q", cfg.Branch, "env-branch")
	}
	if cfg.Token != "env-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "env-token")
	}
	if cfg.BaseURL != "https://enterprise.example.com/api/v3/" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://enterprise.example.com/api/v3/")
	}
	if cfg.UploadURL != "https://enterprise.example.com/uploads/" {
		t.Errorf("UploadURL = %q, want %q", cfg.UploadURL, "https://enterprise.example.com/uploads/")
	}
}

func TestFromEnvWithFallback(t *testing.T) {
	defer func() {
		os.Unsetenv("PRIMARY_TOKEN")
		os.Unsetenv("FALLBACK_TOKEN")
	}()

	primary := EnvConfig{
		Owner:     "PRIMARY_OWNER",
		Repo:      "PRIMARY_REPO",
		Branch:    "PRIMARY_BRANCH",
		Token:     "PRIMARY_TOKEN",
		BaseURL:   "PRIMARY_BASE_URL",
		UploadURL: "PRIMARY_UPLOAD_URL",
	}

	fallback := EnvConfig{
		Owner:     "FALLBACK_OWNER",
		Repo:      "FALLBACK_REPO",
		Branch:    "FALLBACK_BRANCH",
		Token:     "FALLBACK_TOKEN",
		BaseURL:   "FALLBACK_BASE_URL",
		UploadURL: "FALLBACK_UPLOAD_URL",
	}

	// Set only fallback
	os.Setenv("FALLBACK_TOKEN", "fallback-token-value")

	cfg := FromEnvWithFallback(primary, fallback)

	if cfg.Token != "fallback-token-value" {
		t.Errorf("Token = %q, want %q (should use fallback)", cfg.Token, "fallback-token-value")
	}

	// Now set primary
	os.Setenv("PRIMARY_TOKEN", "primary-token-value")

	cfg = FromEnvWithFallback(primary, fallback)

	if cfg.Token != "primary-token-value" {
		t.Errorf("Token = %q, want %q (should use primary)", cfg.Token, "primary-token-value")
	}
}

func TestDefaultEnvConfig(t *testing.T) {
	envCfg := DefaultEnvConfig()

	if envCfg.Owner != EnvOwner {
		t.Errorf("Owner = %q, want %q", envCfg.Owner, EnvOwner)
	}
	if envCfg.Repo != EnvRepo {
		t.Errorf("Repo = %q, want %q", envCfg.Repo, EnvRepo)
	}
	if envCfg.Token != EnvToken {
		t.Errorf("Token = %q, want %q", envCfg.Token, EnvToken)
	}
}

func TestNewClient(t *testing.T) {
	cfg := Config{
		Owner: "owner",
		Repo:  "repo",
		Token: "token",
	}
	cfg.ApplyDefaults()

	client, err := cfg.NewClient(context.Background())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestNewClientEnterprise(t *testing.T) {
	cfg := Config{
		Owner:     "owner",
		Repo:      "repo",
		Token:     "token",
		BaseURL:   "https://enterprise.example.com/api/v3/",
		UploadURL: "https://enterprise.example.com/uploads/",
	}

	client, err := cfg.NewClient(context.Background())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil client")
	}
}

func TestMustNewClientPanicsOnInvalidConfig(t *testing.T) {
	cfg := Config{} // Invalid - missing required fields

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewClient() should panic on invalid config")
		}
	}()

	cfg.MustNewClient(context.Background())
}
