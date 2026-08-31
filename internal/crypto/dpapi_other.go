//go:build !windows

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// mockKey is a static 256-bit key used for stub encryption on non-Windows environments (Linux/macOS).
var mockKey = []byte("SophosQuickMockDPAPIKey32Bytes!!")

// Encrypt provides a mock implementation of DPAPI encryption for non-Windows platforms (Linux/macOS).
// It uses AES-GCM to allow testing and development on non-Windows systems.
func Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	block, err := aes.NewCipher(mockKey)
	if err != nil {
		return nil, fmt.Errorf("mock encryption cipher initialization failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("mock GCM initialization failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Format: [Nonce][Ciphertext+Tag]
	sealed := gcm.Seal(nonce, nonce, data, nil)
	return sealed, nil
}

// Decrypt provides a mock implementation of DPAPI decryption for non-Windows platforms (Linux/macOS).
// It uses AES-GCM to allow testing and development on non-Windows systems.
func Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	block, err := aes.NewCipher(mockKey)
	if err != nil {
		return nil, fmt.Errorf("mock decryption cipher initialization failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("mock GCM initialization failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("invalid ciphertext: length is shorter than nonce")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("mock decryption failed: %w", err)
	}

	return plaintext, nil
}
