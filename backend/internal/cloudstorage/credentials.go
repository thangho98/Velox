package cloudstorage

import (
	"encoding/json"
	"fmt"

	"github.com/thawng/velox/pkg/crypto"
)

// EncryptCredentials serializes creds to JSON then AES-256-GCM encrypts with key.
// The output is ready to be stored in storage_providers.credentials_encrypted.
func EncryptCredentials(key []byte, creds *Credentials) ([]byte, error) {
	b, err := json.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("cloudstorage: marshal creds: %w", err)
	}
	ct, err := crypto.Encrypt(key, b)
	if err != nil {
		return nil, fmt.Errorf("cloudstorage: encrypt creds: %w", err)
	}
	return ct, nil
}

// DecryptCredentials reverses EncryptCredentials. Returns a usable Credentials
// pointer; the caller can inject it into a fresh Provider via Driver.NewProvider.
func DecryptCredentials(key, ciphertext []byte) (*Credentials, error) {
	plain, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("cloudstorage: decrypt creds: %w", err)
	}
	var creds Credentials
	if err := json.Unmarshal(plain, &creds); err != nil {
		return nil, fmt.Errorf("cloudstorage: unmarshal creds: %w", err)
	}
	return &creds, nil
}
