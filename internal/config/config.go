package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// ConfigFileName is the default configuration file name.
	ConfigFileName = "config.json"
	// AppDirName is the configuration sub-folder name.
	AppDirName = "SophosQuick"
)

// Config holds persistent settings for SophosQuick.
type Config struct {
	DefaultConnection   string   `json:"default_connection"`
	Username            string   `json:"username"`
	SccliPath           string   `json:"sccli_path,omitempty"`
	ConnectionsDir      string   `json:"connections_dir,omitempty"`
	FallbackConnections []string `json:"fallback_connections"`
	AutoConnect         bool     `json:"auto_connect"`
	SaveLastUsed        bool     `json:"save_last_used"`
}

// GetCurrentUsername attempts to resolve the current OS username.
func GetCurrentUsername() string {
	// 1. Check USERNAME (Windows standard)
	if u := os.Getenv("USERNAME"); strings.TrimSpace(u) != "" {
		return strings.TrimSpace(u)
	}

	// 2. Check USER (Unix/macOS standard)
	if u := os.Getenv("USER"); strings.TrimSpace(u) != "" {
		return strings.TrimSpace(u)
	}

	// 3. Fallback to os/user
	if u, err := user.Current(); err == nil && u.Username != "" {
		// Strip domain prefix if present (e.g. DOMAIN\user -> user)
		parts := strings.Split(u.Username, `\`)
		return parts[len(parts)-1]
	}

	return ""
}

// DefaultConfig returns a Config struct populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultConnection:   "",
		Username:            "",
		FallbackConnections: []string{},
		AutoConnect:         false,
		SaveLastUsed:        true,
	}
}

// GetDefaultConfigPath resolves the default path to config.json according to the operating system.
// Windows: %LOCALAPPDATA%\SophosQuick\config.json
// Linux/macOS: ~/.config/sophosquick/config.json
func GetDefaultConfigPath() (string, error) {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			userProfile := os.Getenv("USERPROFILE")
			if userProfile == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return "", fmt.Errorf("unable to determine user directory: %w", err)
				}
				localAppData = filepath.Join(homeDir, "AppData", "Local")
			} else {
				localAppData = filepath.Join(userProfile, "AppData", "Local")
			}
		}
		return filepath.Join(localAppData, AppDirName, ConfigFileName), nil
	}

	// Linux / macOS / Unix systems
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, strings.ToLower(AppDirName), ConfigFileName), nil
}

// resolvePath resolves the path to use, preferring customPath if provided.
func resolvePath(customPath ...string) (string, error) {
	if len(customPath) > 0 && strings.TrimSpace(customPath[0]) != "" {
		return customPath[0], nil
	}
	return GetDefaultConfigPath()
}

// LoadConfig loads configuration from disk. If the file does not exist,
// it returns the default configuration without error.
func LoadConfig(customPath ...string) (*Config, error) {
	configPath, err := resolvePath(customPath...)
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON in %q: %w", configPath, err)
	}

	// Ensure fallback connections list is not nil
	if cfg.FallbackConnections == nil {
		cfg.FallbackConnections = []string{}
	}

	return cfg, nil
}

// SaveConfig writes the given Config struct to disk formatted with indentation.
func SaveConfig(cfg *Config, customPath ...string) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	configPath, err := resolvePath(customPath...)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(configPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %q: %w", parentDir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write config file %q: %w", configPath, err)
	}

	return nil
}
