package sophos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanConnectionsDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create dummy profile files
	filesToCreate := []string{
		"vpn.company.com.tgb",
		"vpn_contingencia.company.com.ovpn",
		"office_remote.apx",
		"legacy.conf",
		"custom_connection.json",
		"ignored_notes.txt",    // Should be ignored
		"sophos_log.log",       // Should be ignored
	}

	for _, f := range filesToCreate {
		fullPath := filepath.Join(tempDir, f)
		if err := os.WriteFile(fullPath, []byte("dummy data"), 0644); err != nil {
			t.Fatalf("Failed to create test file %q: %v", f, err)
		}
	}

	// Create a subdirectory (should be ignored)
	if err := os.Mkdir(filepath.Join(tempDir, "subfolder.tgb"), 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	conns, err := ScanConnectionsDir(tempDir)
	if err != nil {
		t.Fatalf("ScanConnectionsDir failed: %v", err)
	}

	expectedFound := map[string]bool{
		"vpn.company.com":               true,
		"vpn_contingencia.company.com": true,
		"office_remote":                   true,
		"legacy":                          true,
		"custom_connection":               true,
	}

	if len(conns) != len(expectedFound) {
		t.Fatalf("Expected %d connections, got %d (%v)", len(expectedFound), len(conns), conns)
	}

	for _, conn := range conns {
		if !expectedFound[conn] {
			t.Errorf("Unexpected connection discovered: %q", conn)
		}
	}
}

func TestDiscoverConnectionsMergeAndFallback(t *testing.T) {
	tempDir := t.TempDir()

	// Create profile in directory
	profilePath := filepath.Join(tempDir, "scanned_profile.ovpn")
	if err := os.WriteFile(profilePath, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	fallbacks := []string{
		"scanned_profile",               // Duplicate of scanned, should deduplicate
		"fallback.company.com",
		"vpn_backup.company.com",
	}

	// Mock client with non-existent path
	client := NewClient("non_existent_binary")

	discovered, err := DiscoverConnections(client, fallbacks, tempDir)
	if err != nil {
		t.Fatalf("DiscoverConnections failed: %v", err)
	}

	expected := []string{
		"scanned_profile",
		"fallback.company.com",
		"vpn_backup.company.com",
	}

	if len(discovered) != len(expected) {
		t.Fatalf("Expected %d connections, got %d (%v)", len(expected), len(discovered), discovered)
	}

	for i, name := range expected {
		if discovered[i] != name {
			t.Errorf("Index %d mismatch: got %q, want %q", i, discovered[i], name)
		}
	}
}
