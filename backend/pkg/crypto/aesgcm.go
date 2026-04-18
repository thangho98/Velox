// Package crypto provides AES-256-GCM authenticated encryption for at-rest
// secrets (cloud credentials, tokens). The package intentionally has zero
// dependencies beyond the standard library so it can be pulled into any part
// of Velox.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// KeySize is the required key length in bytes (AES-256).
	KeySize = 32
	// NonceSize is the GCM nonce length in bytes.
	NonceSize = 12
)

var (
	// ErrInvalidKey is returned when the key is not exactly KeySize bytes.
	ErrInvalidKey = errors.New("crypto: key must be 32 bytes for AES-256")
	// ErrNoKey is returned by LoadKeyFromEnv when VELOX_CLOUD_SECRET is unset.
	ErrNoKey = errors.New("crypto: VELOX_CLOUD_SECRET env is not set")
	// ErrTampered is returned when the ciphertext fails authentication.
	ErrTampered = errors.New("crypto: ciphertext authentication failed (tampered or wrong key)")
	// ErrShortCiphertext is returned for ciphertexts smaller than the GCM nonce + tag.
	ErrShortCiphertext = errors.New("crypto: ciphertext too short")
)

// Encrypt seals plaintext with AES-256-GCM and returns nonce || ciphertext || tag.
// The output is self-contained: pass it back to Decrypt with the same key.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: read nonce: %w", err)
	}

	// Seal appends ciphertext+tag to the dst (nonce prefix), giving us
	// a single self-contained slice.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt reverses Encrypt. Returns ErrTampered if the MAC check fails.
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrShortCiphertext
	}
	nonce, payload := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return nil, ErrTampered
	}
	return plaintext, nil
}

// LoadKeyFromEnv reads VELOX_CLOUD_SECRET as a hex-encoded 32-byte key.
// Returns ErrNoKey when unset and ErrInvalidKey when the decoded value is
// the wrong length or not valid hex.
//
// Generate a key via: openssl rand -hex 32
func LoadKeyFromEnv() ([]byte, error) {
	return parseHexKey(os.Getenv("VELOX_CLOUD_SECRET"))
}

// LoadOrGenerateKey returns a 32-byte AES key. Preference order:
//
//  1. envVar, if set — decoded as hex.
//  2. filePath, if the file exists — decoded as hex.
//  3. Generate a fresh random key, persist it to filePath (0600 perms), return it.
//
// This lets Velox self-configure on first boot while still honoring an
// explicit env override for users who want to manage the key externally
// (e.g. secrets manager, docker secrets).
//
// The file is written with 0600 mode and the parent directory created with
// 0700 mode — anyone with file-system access to the data dir already has
// access to velox.db, so these perms just prevent casual exposure.
func LoadOrGenerateKey(envVar, filePath string) ([]byte, Origin, error) {
	if raw := os.Getenv(envVar); raw != "" {
		key, err := parseHexKey(raw)
		if err != nil {
			return nil, OriginNone, fmt.Errorf("env %s: %w", envVar, err)
		}
		return key, OriginEnv, nil
	}

	if data, err := os.ReadFile(filePath); err == nil {
		key, err := parseHexKey(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, OriginNone, fmt.Errorf("key file %s: %w", filePath, err)
		}
		return key, OriginFile, nil
	}

	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, OriginNone, fmt.Errorf("generate key: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
		return nil, OriginNone, fmt.Errorf("mkdir %s: %w", filepath.Dir(filePath), err)
	}
	if err := os.WriteFile(filePath, []byte(hex.EncodeToString(key)), 0o600); err != nil {
		return nil, OriginNone, fmt.Errorf("write %s: %w", filePath, err)
	}
	return key, OriginGenerated, nil
}

// Origin tells the caller how LoadOrGenerateKey obtained the key, so startup
// logs can distinguish "loaded existing" from "minted fresh".
type Origin int

const (
	OriginNone Origin = iota
	OriginEnv
	OriginFile
	OriginGenerated
)

func (o Origin) String() string {
	switch o {
	case OriginEnv:
		return "env"
	case OriginFile:
		return "file"
	case OriginGenerated:
		return "generated"
	default:
		return "unknown"
	}
}

// parseHexKey decodes a hex-encoded 32-byte AES-256 key.
func parseHexKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, ErrNoKey
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: hex decode: %v", ErrInvalidKey, err)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: got %d bytes, want %d", ErrInvalidKey, len(key), KeySize)
	}
	return key, nil
}
