// Package fshare adapts pkg/fshare into the cloudstorage.Provider interface.
// The Driver is registered with cloudstorage.Registry at startup; callers
// should rely only on the cloudstorage abstraction, not this package directly.
package fshare

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thawng/velox/internal/cloudstorage"
	pkgfshare "github.com/thawng/velox/pkg/fshare"
)

// sessionLifetime is how long a freshly-issued fshare session is treated as
// valid before the scheduler refreshes it. fshare's server-side TTL is ~30min;
// 25min gives a 5-minute safety buffer for the refresh job to run.
const sessionLifetime = 25 * time.Minute

// Config tunes the fshare driver. All fields are optional — pkg/fshare
// provides working defaults verified live in 2026-04.
type Config struct {
	// AppKey overrides pkg/fshare.DefaultAppKey.
	AppKey string
	// UserAgent overrides pkg/fshare's default iOS UA.
	UserAgent string
}

// Driver implements cloudstorage.Driver + cloudstorage.PasswordAuthDriver.
type Driver struct {
	cfg Config
}

// NewDriver builds a fshare driver. The same driver instance can be registered
// with a Registry and used to construct unlimited Provider instances.
func NewDriver(cfg Config) *Driver {
	return &Driver{cfg: cfg}
}

// Type returns the registry key.
func (d *Driver) Type() string { return "fshare" }

// DisplayName is the human-readable label shown in admin UIs.
func (d *Driver) DisplayName() string { return "Fshare" }

// AuthFlow returns AuthFlowPassword — fshare has no OAuth support.
func (d *Driver) AuthFlow() cloudstorage.AuthFlow { return cloudstorage.AuthFlowPassword }

// NewProvider hydrates a Provider from stored credentials. The client gets a
// fresh cookie jar per call; existing session state is restored via
// RestoreSession so subsequent API calls skip re-login unless the server
// reports code:201.
func (d *Driver) NewProvider(creds *cloudstorage.Credentials) (cloudstorage.Provider, error) {
	client, err := d.newClient()
	if err != nil {
		return nil, err
	}
	if creds.AccessToken != "" && creds.SessionID != "" {
		client.RestoreSession(creds.AccessToken, creds.SessionID)
	}
	client.SetCredentials(creds.Email, creds.Password)
	return &provider{
		client: client,
		creds:  snapshot(creds),
	}, nil
}

// AuthenticatePassword implements cloudstorage.PasswordAuthDriver. Returns a
// freshly-issued Credentials snapshot ready for encryption + persistence.
func (d *Driver) AuthenticatePassword(ctx context.Context, email, password string) (*cloudstorage.Credentials, error) {
	client, err := d.newClient()
	if err != nil {
		return nil, err
	}
	sess, err := client.Login(ctx, email, password)
	if err != nil {
		return nil, mapError(err)
	}
	return &cloudstorage.Credentials{
		ProviderType: "fshare",
		Email:        email,
		Password:     password,
		AccessToken:  sess.Token,
		SessionID:    sess.SessionID,
		ExpiresAt:    time.Now().Add(sessionLifetime),
	}, nil
}

func (d *Driver) newClient() (*pkgfshare.Client, error) {
	client, err := pkgfshare.NewClient(pkgfshare.Options{
		AppKey:    d.cfg.AppKey,
		UserAgent: d.cfg.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("fshare driver: new client: %w", err)
	}
	return client, nil
}

// snapshot returns a defensive copy of creds to avoid aliasing between the
// caller's decrypted Credentials and the provider's private Session state.
func snapshot(c *cloudstorage.Credentials) cloudstorage.Credentials {
	out := *c
	if c.Meta != nil {
		out.Meta = make(map[string]string, len(c.Meta))
		for k, v := range c.Meta {
			out.Meta[k] = v
		}
	}
	return out
}

// mapError translates pkg/fshare errors into cloudstorage errors so the
// resolver/handler layer depends only on cloudstorage.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, pkgfshare.ErrSessionExpired):
		return cloudstorage.ErrSessionExpired
	case errors.Is(err, pkgfshare.ErrInvalidCredentials):
		return cloudstorage.ErrInvalidCredentials
	case errors.Is(err, pkgfshare.ErrAppKeyInvalid):
		return fmt.Errorf("%w: fshare app_key revoked", cloudstorage.ErrInvalidCredentials)
	case errors.Is(err, pkgfshare.ErrCaptchaRequired):
		return fmt.Errorf("%w: captcha required — use cookie paste fallback", cloudstorage.ErrInvalidCredentials)
	case errors.Is(err, pkgfshare.ErrRateLimit):
		return cloudstorage.ErrRateLimit
	case errors.Is(err, pkgfshare.ErrLinkDead), errors.Is(err, pkgfshare.ErrFilePassword):
		return cloudstorage.ErrFileNotFound
	case errors.Is(err, pkgfshare.ErrNotVIP):
		return cloudstorage.ErrAccountNotPremium
	case errors.Is(err, pkgfshare.ErrNotLoggedIn):
		return cloudstorage.ErrSessionExpired
	default:
		return err
	}
}
