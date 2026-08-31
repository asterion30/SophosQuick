package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	testCases := []string{
		"SimplePassword123!",
		"Complex P@$$w0rd with spaces and utf8: áéíóú ñ 🚀",
		"a",
		"very_long_password_1234567890_abcdefghijklmnopqrstuvwxyz_!@#$%^&*()_+",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			encrypted, err := Encrypt([]byte(tc))
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if len(encrypted) == 0 {
				t.Fatalf("Encrypted output is empty")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(decrypted) != tc {
				t.Fatalf("Decrypted mismatch: got %q, want %q", string(decrypted), tc)
			}
		})
	}
}

func TestEncryptDecryptEmpty(t *testing.T) {
	enc, err := Encrypt([]byte{})
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}
	if len(enc) != 0 {
		t.Fatalf("Expected empty output for empty input, got %v", enc)
	}

	dec, err := Decrypt([]byte{})
	if err != nil {
		t.Fatalf("Decrypt empty failed: %v", err)
	}
	if len(dec) != 0 {
		t.Fatalf("Expected empty output for empty input, got %v", dec)
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	corrupted := []byte{0x01, 0x02, 0x03, 0x04}
	_, err := Decrypt(corrupted)
	if err == nil {
		t.Fatalf("Expected error when decrypting corrupted data, got nil")
	}
}

func TestDeterministicOrSafeMock(t *testing.T) {
	data := []byte("secret")
	enc1, err1 := Encrypt(data)
	enc2, err2 := Encrypt(data)
	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected encryption failure: %v, %v", err1, err2)
	}

	dec1, err := Decrypt(enc1)
	if err != nil || !bytes.Equal(dec1, data) {
		t.Fatalf("dec1 mismatch: %v, %s", err, string(dec1))
	}

	dec2, err := Decrypt(enc2)
	if err != nil || !bytes.Equal(dec2, data) {
		t.Fatalf("dec2 mismatch: %v, %s", err, string(dec2))
	}
}
