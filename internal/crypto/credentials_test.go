package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "test_cred.bin")

	// 1. Initial state: HasSavedCredential should be false
	if HasSavedCredential(credPath) {
		t.Fatalf("Expected HasSavedCredential to be false before saving")
	}

	// 2. Load non-existent credential should return error
	_, err := LoadPassword(credPath)
	if err == nil {
		t.Fatalf("Expected error when loading non-existent credential, got nil")
	}

	// 3. Save password
	testPassword := "MySecretPassword!2026"
	if err := SavePassword(testPassword, credPath); err != nil {
		t.Fatalf("SavePassword failed: %v", err)
	}

	// 4. HasSavedCredential should now be true
	if !HasSavedCredential(credPath) {
		t.Fatalf("Expected HasSavedCredential to be true after saving")
	}

	// 5. Load password and verify
	loaded, err := LoadPassword(credPath)
	if err != nil {
		t.Fatalf("LoadPassword failed: %v", err)
	}
	if loaded != testPassword {
		t.Fatalf("Password mismatch: got %q, want %q", loaded, testPassword)
	}

	// 6. Delete credential
	if err := DeleteCredential(credPath); err != nil {
		t.Fatalf("DeleteCredential failed: %v", err)
	}

	// 7. Verify deletion
	if HasSavedCredential(credPath) {
		t.Fatalf("Expected HasSavedCredential to be false after deletion")
	}

	// 8. Delete non-existent credential should not fail
	if err := DeleteCredential(credPath); err != nil {
		t.Fatalf("DeleteCredential on missing file returned error: %v", err)
	}
}

func TestSaveEmptyPassword(t *testing.T) {
	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "empty_cred.bin")

	err := SavePassword("", credPath)
	if err == nil {
		t.Fatalf("Expected error when saving empty password, got nil")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "empty_file.bin")

	if err := os.WriteFile(credPath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := LoadPassword(credPath)
	if err == nil {
		t.Fatalf("Expected error when loading empty credential file, got nil")
	}
}
