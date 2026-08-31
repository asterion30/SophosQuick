package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatalf("DefaultConfig returned nil")
	}

	if cfg.DefaultConnection != "vpn.company.com" {
		t.Errorf("Expected default connection 'vpn.company.com', got %q", cfg.DefaultConnection)
	}

	if len(cfg.FallbackConnections) < 2 {
		t.Errorf("Expected at least 2 fallback connections, got %d", len(cfg.FallbackConnections))
	}
}

func TestLoadNonExistentConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "does_not_exist.json")

	cfg, err := LoadConfig(nonExistentPath)
	if err != nil {
		t.Fatalf("LoadConfig should not fail for non-existent file, got: %v", err)
	}

	if cfg.DefaultConnection != "vpn.company.com" {
		t.Errorf("Expected default connection 'vpn.company.com', got %q", cfg.DefaultConnection)
	}
}

func TestSaveAndLoadConfigRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom_config.json")

	customCfg := &Config{
		DefaultConnection:   "custom.vpn.company.com",
		Username:            "testuser42",
		SccliPath:           "C:\\Custom\\sccli.exe",
		ConnectionsDir:      "C:\\Custom\\Connections",
		FallbackConnections: []string{"conn1", "conn2", "conn3"},
		AutoConnect:         true,
		SaveLastUsed:        false,
	}

	if err := SaveConfig(customCfg, configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.DefaultConnection != customCfg.DefaultConnection {
		t.Errorf("DefaultConnection mismatch: got %q, want %q", loaded.DefaultConnection, customCfg.DefaultConnection)
	}

	if loaded.Username != customCfg.Username {
		t.Errorf("Username mismatch: got %q, want %q", loaded.Username, customCfg.Username)
	}

	if loaded.SccliPath != customCfg.SccliPath {
		t.Errorf("SccliPath mismatch: got %q, want %q", loaded.SccliPath, customCfg.SccliPath)
	}

	if loaded.ConnectionsDir != customCfg.ConnectionsDir {
		t.Errorf("ConnectionsDir mismatch: got %q, want %q", loaded.ConnectionsDir, customCfg.ConnectionsDir)
	}

	if len(loaded.FallbackConnections) != len(customCfg.FallbackConnections) {
		t.Fatalf("FallbackConnections length mismatch: got %d, want %d", len(loaded.FallbackConnections), len(customCfg.FallbackConnections))
	}

	for i, conn := range customCfg.FallbackConnections {
		if loaded.FallbackConnections[i] != conn {
			t.Errorf("FallbackConnection[%d] mismatch: got %q, want %q", i, loaded.FallbackConnections[i], conn)
		}
	}

	if loaded.AutoConnect != customCfg.AutoConnect {
		t.Errorf("AutoConnect mismatch: got %v, want %v", loaded.AutoConnect, customCfg.AutoConnect)
	}

	if loaded.SaveLastUsed != customCfg.SaveLastUsed {
		t.Errorf("SaveLastUsed mismatch: got %v, want %v", loaded.SaveLastUsed, customCfg.SaveLastUsed)
	}
}
