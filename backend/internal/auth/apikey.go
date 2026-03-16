package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// APIKeyEntry represents an issued API key with expiry and user info.
type APIKeyEntry struct {
	UserID    int64
	IsAdmin   bool
	ExpiresAt time.Time
}

// APIKeyStore manages short-lived API keys for external player URLs.
// Keys are stored in memory and expire after StreamTokenExpiry (2 hours).
type APIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]APIKeyEntry
}

// NewAPIKeyStore creates a new in-memory API key store.
func NewAPIKeyStore() *APIKeyStore {
	s := &APIKeyStore{keys: make(map[string]APIKeyEntry)}
	go s.cleanupLoop()
	return s
}

// Generate creates a new API key for the given user. Returns a 32-char hex string.
func (s *APIKeyStore) Generate(userID int64, isAdmin bool) string {
	b := make([]byte, 16)
	rand.Read(b)
	key := hex.EncodeToString(b)

	s.mu.Lock()
	s.keys[key] = APIKeyEntry{
		UserID:    userID,
		IsAdmin:   isAdmin,
		ExpiresAt: time.Now().Add(StreamTokenExpiry),
	}
	s.mu.Unlock()

	return key
}

// Validate checks if an API key is valid and not expired.
// Returns user info if valid, nil otherwise.
func (s *APIKeyStore) Validate(key string) *APIKeyEntry {
	s.mu.RLock()
	entry, ok := s.keys[key]
	s.mu.RUnlock()

	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return &entry
}

// cleanupLoop removes expired keys every 10 minutes.
func (s *APIKeyStore) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.keys {
			if now.After(v.ExpiresAt) {
				delete(s.keys, k)
			}
		}
		s.mu.Unlock()
	}
}
