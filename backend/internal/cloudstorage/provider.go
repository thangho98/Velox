package cloudstorage

import "context"

// Provider operates on an authenticated cloud session. One Provider instance
// corresponds to one cloud account. Drivers construct Provider instances
// lazily via [Driver.NewProvider].
type Provider interface {
	// Type returns the provider's registered type (e.g. "fshare").
	Type() string

	// ParseFolderURL extracts the native folder ID from a user-pasted URL.
	// Returns ErrInvalidURL when the URL doesn't match the provider's format.
	ParseFolderURL(rawURL string) (folderID string, err error)

	// ParseFileURL extracts the native file ID from a user-pasted URL.
	// Returns ErrInvalidURL when the URL doesn't match the provider's format.
	ParseFileURL(rawURL string) (fileID string, err error)

	// ListFolder returns items inside folderID. Empty string means the account
	// root. Implementations handle pagination internally.
	ListFolder(ctx context.Context, folderID string) ([]Item, error)

	// GetDownloadURL returns a short-lived direct download URL for the file.
	// Callers MUST NOT cache the returned URL — providers commonly issue
	// single-use links.
	GetDownloadURL(ctx context.Context, fileID string) (string, error)

	// GetAccountInfo fetches account metadata (tier, quota, expiry). Also
	// works as a session health check — returns ErrSessionExpired when the
	// underlying session is stale.
	GetAccountInfo(ctx context.Context) (*AccountInfo, error)

	// CheckHealth asserts the session is valid without returning metadata.
	// Useful for scheduled ping tasks.
	CheckHealth(ctx context.Context) error

	// Session returns the current credentials snapshot for persistence. After
	// any Provider call, callers should diff this against the pre-call value
	// and save back to storage_providers if changed (auto-relogin case).
	Session() *Credentials
}

// Driver is the factory + auth metadata for a provider type. Drivers register
// themselves with a Registry at startup.
type Driver interface {
	// Type is the stable registry key (e.g. "fshare").
	Type() string
	// DisplayName is the human-readable label shown in the admin UI.
	DisplayName() string
	// AuthFlow selects which admin-UI form to render.
	AuthFlow() AuthFlow
	// NewProvider constructs a Provider instance from stored credentials.
	// Drivers must not share state between Provider instances — cookie jars
	// and session tokens are per-instance.
	NewProvider(creds *Credentials) (Provider, error)
}

// PasswordAuthDriver is implemented by drivers whose auth flow accepts an
// email + password pair (fshare today).
type PasswordAuthDriver interface {
	Driver
	AuthenticatePassword(ctx context.Context, email, password string) (*Credentials, error)
}

// OAuthDriver is implemented by drivers using the OAuth 2.0 authorization-code
// flow (GDrive, OneDrive in the future). MVP has no OAuth drivers; interface
// exists so the refresh scheduler can dispatch polymorphically without code
// changes when the first OAuth driver ships.
type OAuthDriver interface {
	Driver
	AuthURL(state, redirectURI string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (*Credentials, error)
	RefreshToken(ctx context.Context, creds *Credentials) (*Credentials, error)
}
