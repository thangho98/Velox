// Package cloudstorage is the Velox-specific abstraction layer over
// third-party cloud providers (fshare, Google Drive, OneDrive, ...).
//
// Scanner and stream handler code depend only on this package; adding a new
// provider means implementing [Provider] + [Driver] in a drivers/* subpackage
// and registering the driver at startup.
package cloudstorage

import "time"

// AuthFlow is the authentication protocol a driver uses. The admin UI picks
// the form fields based on this enum.
type AuthFlow int

const (
	// AuthFlowPassword is email + password auth (fshare, FTP-like).
	AuthFlowPassword AuthFlow = iota
	// AuthFlowOAuth is the OAuth 2.0 authorization-code flow (GDrive, OneDrive).
	AuthFlowOAuth
)

// String returns the wire-format identifier used by the admin UI.
func (a AuthFlow) String() string {
	switch a {
	case AuthFlowPassword:
		return "password"
	case AuthFlowOAuth:
		return "oauth"
	default:
		return "unknown"
	}
}

// Item is the uniform representation of a folder or file across all providers.
// Drivers translate their native shape into this struct.
type Item struct {
	// ID is the provider-native identifier (fshare linkcode, gdrive fileId, …).
	ID       string
	Name     string
	IsFolder bool
	Size     int64
	Mimetype string
	Modified time.Time
	// URL is the canonical user-facing URL (for display or paste-back).
	URL string
}

// AccountInfo is the uniform account metadata across providers.
// Provider-specific fields beyond these should flow through Credentials.Meta.
type AccountInfo struct {
	Email       string
	DisplayName string
	// AccountType is the provider's tier label ("Vip" for fshare, "Paid"/"Free"
	// for gdrive). Consumers check this before gating cloud downloads.
	AccountType string
	QuotaBytes  int64
	UsedBytes   int64
	// ExpiresAt is the VIP/subscription expiry for the provider tier.
	// Zero time means the provider has no subscription concept.
	ExpiresAt time.Time
}

// Credentials is the encrypted-at-rest auth state for one provider instance.
// Different drivers populate different subsets:
//   - Password-auth (fshare): Email, Password, AccessToken, SessionID, ExpiresAt
//   - OAuth (GDrive, OneDrive): Email, AccessToken, RefreshToken, ExpiresAt
//
// Meta is a driver-defined bag for extra fields (e.g. fshare account_type cache).
type Credentials struct {
	ProviderType string            `json:"provider_type"`
	Email        string            `json:"email"`
	Password     string            `json:"password,omitempty"`
	AccessToken  string            `json:"access_token,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Meta         map[string]string `json:"meta,omitempty"`
}
