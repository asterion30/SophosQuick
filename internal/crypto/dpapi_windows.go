//go:build windows

package crypto

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	dllCrypt32             = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData   = dllCrypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = dllCrypt32.NewProc("CryptUnprotectData")
	dllKernel32            = syscall.NewLazyDLL("kernel32.dll")
	procLocalFree          = dllKernel32.NewProc("LocalFree")
)

// dataBlob represents the Windows DATA_BLOB structure used in Crypt32 API calls.
type dataBlob struct {
	cbData uint32
	pbData *byte
}

// newBlob creates a new dataBlob pointing to the underlying byte slice data.
func newBlob(d []byte) *dataBlob {
	if len(d) == 0 {
		return &dataBlob{}
	}
	return &dataBlob{
		cbData: uint32(len(d)),
		pbData: &d[0],
	}
}

// toByteArray copies memory from the Windows-allocated buffer into a Go byte slice.
func (b *dataBlob) toByteArray() []byte {
	if b.cbData == 0 || b.pbData == nil {
		return []byte{}
	}
	out := make([]byte, b.cbData)
	copy(out, unsafe.Slice(b.pbData, b.cbData))
	return out
}

// Encrypt encrypts data using Windows DPAPI (CryptProtectData).
// The encrypted data is cryptographically bound to the current Windows user session.
func Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	inBlob := newBlob(data)
	var outBlob dataBlob

	r, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(inBlob)),
		0, // szDataDescr
		0, // pOptionalEntropy
		0, // pvReserved
		0, // pPromptStruct
		0, // dwFlags (0 = current user)
		uintptr(unsafe.Pointer(&outBlob)),
	)
	if r == 0 {
		return nil, fmt.Errorf("DPAPI CryptProtectData failed: %w", err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData)))

	return outBlob.toByteArray(), nil
}

// Decrypt decrypts data using Windows DPAPI (CryptUnprotectData).
// Only the Windows user account that encrypted the data can decrypt it.
func Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	inBlob := newBlob(data)
	var outBlob dataBlob

	r, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(inBlob)),
		0, // ppszDataDescr
		0, // pOptionalEntropy
		0, // pvReserved
		0, // pPromptStruct
		0, // dwFlags
		uintptr(unsafe.Pointer(&outBlob)),
	)
	if r == 0 {
		return nil, fmt.Errorf("DPAPI CryptUnprotectData failed: %w", err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData)))

	return outBlob.toByteArray(), nil
}
