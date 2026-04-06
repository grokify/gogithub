package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Could not get user home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/config/app.json",
			expected: filepath.Join(home, "config/app.json"),
		},
		{
			name:     "no tilde",
			input:    "/absolute/path/app.json",
			expected: "/absolute/path/app.json",
		},
		{
			name:     "relative path",
			input:    "relative/path/app.json",
			expected: "relative/path/app.json",
		},
		{
			name:     "tilde in middle (not expanded)",
			input:    "/path/~/config",
			expected: "/path/~/config",
		},
		{
			name:     "just tilde",
			input:    "~/",
			expected: home, // filepath.Join removes trailing slash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAppConfigPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

		configDir := filepath.Join(string(filepath.Separator), "custom", "config")
		os.Setenv("XDG_CONFIG_HOME", configDir)
		result := getAppConfigPath()
		expected := filepath.Join(configDir, "gogithub", "app.json")
		if result != expected {
			t.Errorf("getAppConfigPath() = %q, want %q", result, expected)
		}
	})

	// Test without XDG_CONFIG_HOME (uses ~/.config)
	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

		os.Unsetenv("XDG_CONFIG_HOME")
		home, _ := os.UserHomeDir()
		result := getAppConfigPath()
		expected := filepath.Join(home, ".config", "gogithub", "app.json")
		if result != expected {
			t.Errorf("getAppConfigPath() = %q, want %q", result, expected)
		}
	})
}

func TestLoadAppConfigFromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.json")

	config := AppConfig{
		AppID:          123456,
		InstallationID: 789012,
		PrivateKeyPath: "/path/to/key.pem",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test loading
	loaded, err := LoadAppConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfigFromFile() error: %v", err)
	}

	if loaded.AppID != config.AppID {
		t.Errorf("AppID = %d, want %d", loaded.AppID, config.AppID)
	}
	if loaded.InstallationID != config.InstallationID {
		t.Errorf("InstallationID = %d, want %d", loaded.InstallationID, config.InstallationID)
	}
	if loaded.PrivateKeyPath != config.PrivateKeyPath {
		t.Errorf("PrivateKeyPath = %q, want %q", loaded.PrivateKeyPath, config.PrivateKeyPath)
	}
}

func TestLoadAppConfigFromFileNotFound(t *testing.T) {
	_, err := LoadAppConfigFromFile("/nonexistent/path/app.json")
	if err == nil {
		t.Error("LoadAppConfigFromFile() should return error for nonexistent file")
	}
}

func TestLoadAppConfigFromFileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := LoadAppConfigFromFile(configPath)
	if err == nil {
		t.Error("LoadAppConfigFromFile() should return error for invalid JSON")
	}
}

func TestLoadAppConfigFromEnv(t *testing.T) {
	// Save original env vars
	origAppID := os.Getenv("GITHUB_APP_ID")
	origInstallID := os.Getenv("GITHUB_INSTALLATION_ID")
	origKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	origKey := os.Getenv("GITHUB_PRIVATE_KEY")

	defer func() {
		os.Setenv("GITHUB_APP_ID", origAppID)
		os.Setenv("GITHUB_INSTALLATION_ID", origInstallID)
		os.Setenv("GITHUB_PRIVATE_KEY_PATH", origKeyPath)
		os.Setenv("GITHUB_PRIVATE_KEY", origKey)
	}()

	t.Run("all from env", func(t *testing.T) {
		os.Setenv("GITHUB_APP_ID", "123456")
		os.Setenv("GITHUB_INSTALLATION_ID", "789012")
		os.Setenv("GITHUB_PRIVATE_KEY", "fake-key-content")
		os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

		cfg, err := LoadAppConfig()
		if err != nil {
			t.Fatalf("LoadAppConfig() error: %v", err)
		}

		if cfg.AppID != 123456 {
			t.Errorf("AppID = %d, want %d", cfg.AppID, 123456)
		}
		if cfg.InstallationID != 789012 {
			t.Errorf("InstallationID = %d, want %d", cfg.InstallationID, 789012)
		}
		if string(cfg.PrivateKey) != "fake-key-content" {
			t.Errorf("PrivateKey = %q, want %q", string(cfg.PrivateKey), "fake-key-content")
		}
	})

	t.Run("invalid app ID", func(t *testing.T) {
		os.Setenv("GITHUB_APP_ID", "not-a-number")
		os.Setenv("GITHUB_INSTALLATION_ID", "789012")
		os.Setenv("GITHUB_PRIVATE_KEY", "fake-key")

		_, err := LoadAppConfig()
		if err == nil {
			t.Error("LoadAppConfig() should return error for invalid app ID")
		}
	})

	t.Run("invalid installation ID", func(t *testing.T) {
		os.Setenv("GITHUB_APP_ID", "123456")
		os.Setenv("GITHUB_INSTALLATION_ID", "not-a-number")
		os.Setenv("GITHUB_PRIVATE_KEY", "fake-key")

		_, err := LoadAppConfig()
		if err == nil {
			t.Error("LoadAppConfig() should return error for invalid installation ID")
		}
	})

	t.Run("missing app ID", func(t *testing.T) {
		os.Unsetenv("GITHUB_APP_ID")
		os.Setenv("GITHUB_INSTALLATION_ID", "789012")
		os.Setenv("GITHUB_PRIVATE_KEY", "fake-key")

		_, err := LoadAppConfig()
		if err == nil {
			t.Error("LoadAppConfig() should return error for missing app ID")
		}
	})
}

func TestAppConfigStruct(t *testing.T) {
	cfg := AppConfig{
		AppID:          12345,
		InstallationID: 67890,
		PrivateKeyPath: "/path/to/key.pem",
		PrivateKey:     []byte("test-key"),
	}

	if cfg.AppID != 12345 {
		t.Errorf("AppID = %d, want %d", cfg.AppID, 12345)
	}
	if cfg.InstallationID != 67890 {
		t.Errorf("InstallationID = %d, want %d", cfg.InstallationID, 67890)
	}
	if cfg.PrivateKeyPath != "/path/to/key.pem" {
		t.Errorf("PrivateKeyPath = %q, want %q", cfg.PrivateKeyPath, "/path/to/key.pem")
	}
	if string(cfg.PrivateKey) != "test-key" {
		t.Errorf("PrivateKey = %q, want %q", string(cfg.PrivateKey), "test-key")
	}
}

func TestAppInstallationStruct(t *testing.T) {
	inst := AppInstallation{
		ID:      12345,
		Account: "myorg",
		Type:    "Organization",
	}

	if inst.ID != 12345 {
		t.Errorf("ID = %d, want %d", inst.ID, 12345)
	}
	if inst.Account != "myorg" {
		t.Errorf("Account = %q, want %q", inst.Account, "myorg")
	}
	if inst.Type != "Organization" {
		t.Errorf("Type = %q, want %q", inst.Type, "Organization")
	}
}

func TestJWTExpiryConstant(t *testing.T) {
	// JWTExpiry should be 10 minutes per GitHub's requirements
	expectedMinutes := 10
	actualMinutes := int(JWTExpiry.Minutes())

	if actualMinutes != expectedMinutes {
		t.Errorf("JWTExpiry = %v (%d minutes), want %d minutes", JWTExpiry, actualMinutes, expectedMinutes)
	}
}
