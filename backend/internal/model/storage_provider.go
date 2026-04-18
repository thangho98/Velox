package model

import "time"

// Provider type constants. Drivers register under these.
const (
	ProviderTypeFshare      = "fshare"
	ProviderTypeGoogleDrive = "google_drive" // future
	ProviderTypeOneDrive    = "onedrive"     // future
)

// StorageProvider represents one cloud account configured by a user.
// Multiple Libraries can share the same StorageProvider (different folders).
type StorageProvider struct {
	ID                   int64      `json:"id"`
	UserID               int64      `json:"user_id"`
	ProviderType         string     `json:"provider_type"`
	DisplayName          string     `json:"display_name"`
	AccountEmail         string     `json:"account_email"`
	CredentialsEncrypted []byte     `json:"-"` // never expose to API
	AccountType          *string    `json:"account_type,omitempty"`
	QuotaBytes           *int64     `json:"quota_bytes,omitempty"`
	UsedBytes            *int64     `json:"used_bytes,omitempty"`
	AccountExpiresAt     *time.Time `json:"account_expires_at,omitempty"`
	TokenExpiresAt       *time.Time `json:"token_expires_at,omitempty"`
	LastRefreshAt        *time.Time `json:"last_refresh_at,omitempty"`
	LastCheckAt          *time.Time `json:"last_check_at,omitempty"`
	LastError            *string    `json:"last_error,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// HealthStatus is a derived health flag for the admin dashboard.
type HealthStatus string

const (
	HealthHealthy      HealthStatus = "healthy"
	HealthExpiringSoon HealthStatus = "expiring_soon"
	HealthExpired      HealthStatus = "expired"
	HealthError        HealthStatus = "error"
	HealthUnknown      HealthStatus = "unknown"
)

// ComputeHealth derives a health status from the persisted timestamps + error.
// Tuned for fshare's ~30min session TTL; also sensible for OAuth providers
// because token_expires_at is set after every refresh.
func (p *StorageProvider) ComputeHealth(now time.Time) HealthStatus {
	if p.LastError != nil && *p.LastError != "" {
		return HealthError
	}
	if p.TokenExpiresAt == nil {
		return HealthUnknown
	}
	if !p.TokenExpiresAt.After(now) {
		return HealthExpired
	}
	if p.TokenExpiresAt.Sub(now) < 5*time.Minute {
		return HealthExpiringSoon
	}
	return HealthHealthy
}
