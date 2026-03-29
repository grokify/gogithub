package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v84/github"
)

// AppConfig holds GitHub App configuration.
type AppConfig struct {
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
	PrivateKeyPath string `json:"private_key_path"`
	PrivateKey     []byte `json:"-"` // Can be set directly instead of via path
}

// LoadAppConfig loads GitHub App configuration from environment variables or config file.
// Environment variables take precedence over config file values.
//
// Environment variables:
//   - GITHUB_APP_ID: The GitHub App ID
//   - GITHUB_INSTALLATION_ID: The installation ID for the target org/repo
//   - GITHUB_PRIVATE_KEY_PATH: Path to the private key PEM file
//   - GITHUB_PRIVATE_KEY: The private key PEM content (alternative to path)
//
// Config file location (in order of precedence):
//   - $XDG_CONFIG_HOME/gogithub/app.json
//   - ~/.config/gogithub/app.json
func LoadAppConfig() (*AppConfig, error) {
	cfg := &AppConfig{}

	// Try environment variables first
	if appID := os.Getenv("GITHUB_APP_ID"); appID != "" {
		id, err := strconv.ParseInt(appID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid GITHUB_APP_ID: %w", err)
		}
		cfg.AppID = id
	}

	if installID := os.Getenv("GITHUB_INSTALLATION_ID"); installID != "" {
		id, err := strconv.ParseInt(installID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid GITHUB_INSTALLATION_ID: %w", err)
		}
		cfg.InstallationID = id
	}

	if keyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH"); keyPath != "" {
		cfg.PrivateKeyPath = expandPath(keyPath)
	}

	if key := os.Getenv("GITHUB_PRIVATE_KEY"); key != "" {
		cfg.PrivateKey = []byte(key)
	}

	// If we have all required values from env vars, return
	if cfg.AppID != 0 && cfg.InstallationID != 0 && (cfg.PrivateKeyPath != "" || len(cfg.PrivateKey) > 0) {
		return cfg, nil
	}

	// Try config file
	configPath := getAppConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fileCfg, err := loadAppConfigFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}

		// Merge with env vars (env vars take precedence)
		if cfg.AppID == 0 {
			cfg.AppID = fileCfg.AppID
		}
		if cfg.InstallationID == 0 {
			cfg.InstallationID = fileCfg.InstallationID
		}
		if cfg.PrivateKeyPath == "" && len(cfg.PrivateKey) == 0 {
			cfg.PrivateKeyPath = expandPath(fileCfg.PrivateKeyPath)
		}
	}

	// Validate required fields
	if cfg.AppID == 0 {
		return nil, fmt.Errorf("missing app_id: set GITHUB_APP_ID or add to config file")
	}
	if cfg.InstallationID == 0 {
		return nil, fmt.Errorf("missing installation_id: set GITHUB_INSTALLATION_ID or add to config file")
	}
	if cfg.PrivateKeyPath == "" && len(cfg.PrivateKey) == 0 {
		return nil, fmt.Errorf("missing private key: set GITHUB_PRIVATE_KEY_PATH, GITHUB_PRIVATE_KEY, or add to config file")
	}

	return cfg, nil
}

// LoadAppConfigFromFile loads GitHub App configuration from a specific file path.
func LoadAppConfigFromFile(path string) (*AppConfig, error) {
	return loadAppConfigFile(path)
}

func loadAppConfigFile(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getAppConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "gogithub", "app.json")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gogithub", "app.json")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// NewAppClient creates a GitHub client authenticated as a GitHub App installation.
// The client uses an installation access token that is valid for 1 hour.
func NewAppClient(ctx context.Context, cfg *AppConfig) (*github.Client, error) {
	privateKey := cfg.PrivateKey
	if len(privateKey) == 0 {
		var err error
		privateKey, err = os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading private key: %w", err)
		}
	}

	// Create JWT for App authentication
	token, err := createAppJWT(cfg.AppID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("creating JWT: %w", err)
	}

	// Create client with JWT auth to get installation token
	jwtClient := github.NewClient(nil).WithAuthToken(token)

	// Get installation access token
	installToken, _, err := jwtClient.Apps.CreateInstallationToken(ctx, cfg.InstallationID, nil)
	if err != nil {
		return nil, fmt.Errorf("creating installation token: %w", err)
	}

	// Create client with installation token
	client := github.NewClient(nil).WithAuthToken(installToken.GetToken())

	return client, nil
}

// createAppJWT creates a JWT for GitHub App authentication.
// The JWT is valid for 10 minutes per GitHub's requirements.
func createAppJWT(appID int64, privateKey []byte) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", fmt.Errorf("parsing private key: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

// AppInstallation contains information about a GitHub App installation.
type AppInstallation struct {
	ID      int64
	Account string
	Type    string // "User" or "Organization"
}

// ListAppInstallations lists all installations of the GitHub App.
// This requires authenticating as the App (using JWT), not as an installation.
func ListAppInstallations(ctx context.Context, cfg *AppConfig) ([]AppInstallation, error) {
	privateKey := cfg.PrivateKey
	if len(privateKey) == 0 {
		var err error
		privateKey, err = os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading private key: %w", err)
		}
	}

	token, err := createAppJWT(cfg.AppID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("creating JWT: %w", err)
	}

	client := github.NewClient(nil).WithAuthToken(token)

	installations, _, err := client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing installations: %w", err)
	}

	result := make([]AppInstallation, len(installations))
	for i, inst := range installations {
		result[i] = AppInstallation{
			ID:      inst.GetID(),
			Account: inst.GetAccount().GetLogin(),
			Type:    inst.GetAccount().GetType(),
		}
	}

	return result, nil
}
