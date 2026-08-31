package crypto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// DefaultWindowsFileName is the name of the encrypted credential file on Windows.
	DefaultWindowsFileName = "SophosVPN_Cred.bin"
	// DefaultLinuxSubDir is the subdirectory inside ~/.config on Unix-like systems.
	DefaultLinuxSubDir = "sophosquick"
	// DefaultLinuxFileName is the name of the credential file on Unix-like systems.
	DefaultLinuxFileName = "cred.bin"
)

// GetDefaultCredentialPath resolves the default path for storing encrypted credentials.
// On Windows: %LOCALAPPDATA%\SophosVPN_Cred.bin
// On Linux/macOS: ~/.config/sophosquick/cred.bin
func GetDefaultCredentialPath() (string, error) {
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
		return filepath.Join(localAppData, DefaultWindowsFileName), nil
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

	return filepath.Join(configDir, DefaultLinuxSubDir, DefaultLinuxFileName), nil
}

// resolvePath returns the provided path if non-empty, otherwise resolves the default path.
func resolvePath(customPath ...string) (string, error) {
	if len(customPath) > 0 && strings.TrimSpace(customPath[0]) != "" {
		return customPath[0], nil
	}
	return GetDefaultCredentialPath()
}

// SavePassword encrypts the given plain-text password using DPAPI (or mock DPAPI on Unix)
// and writes the encrypted payload to disk.
func SavePassword(password string, customPath ...string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	targetPath, err := resolvePath(customPath...)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", parentDir, err)
	}

	// Encrypt the password
	encryptedData, err := Encrypt([]byte(password))
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Write encrypted bytes to disk with restricted permissions (0600)
	if err := os.WriteFile(targetPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write credential file %q: %w", targetPath, err)
	}

	return nil
}

// LoadPassword reads the encrypted credential file from disk and decrypts the password.
func LoadPassword(customPath ...string) (string, error) {
	targetPath, err := resolvePath(customPath...)
	if err != nil {
		return "", err
	}

	encryptedData, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("credential file not found at %q: %w", targetPath, err)
		}
		return "", fmt.Errorf("failed to read credential file %q: %w", targetPath, err)
	}

	if len(encryptedData) == 0 {
		return "", errors.New("credential file is empty")
	}

	decryptedBytes, err := Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password from %q: %w", targetPath, err)
	}

	return string(decryptedBytes), nil
}

// HasSavedCredential checks whether the encrypted credential file exists on disk.
func HasSavedCredential(customPath ...string) bool {
	targetPath, err := resolvePath(customPath...)
	if err != nil {
		return false
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return false
	}

	return !info.IsDir() && info.Size() > 0
}

// DeleteCredential removes the saved encrypted credential file from disk.
func DeleteCredential(customPath ...string) error {
	targetPath, err := resolvePath(customPath...)
	if err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credential file %q: %w", targetPath, err)
	}

	return nil
}
