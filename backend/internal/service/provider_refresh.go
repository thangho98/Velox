package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/thawng/velox/internal/cloudstorage"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// ProviderRefreshService keeps cloud storage sessions fresh. fshare sessions
// expire ~30min server-side; this service re-logs in at 25min to avoid
// surprise code:201s during user playback.
//
// Dispatches polymorphically by driver auth flow: password-auth drivers
// (fshare) re-login with stored credentials; OAuth drivers (GDrive, OneDrive
// in the future) exchange their refresh token.
type ProviderRefreshService struct {
	repo      *repository.StorageProviderRepo
	registry  *cloudstorage.Registry
	cryptoKey []byte
	logger    *slog.Logger
}

func NewProviderRefreshService(
	repo *repository.StorageProviderRepo,
	registry *cloudstorage.Registry,
	cryptoKey []byte,
	logger *slog.Logger,
) *ProviderRefreshService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProviderRefreshService{
		repo:      repo,
		registry:  registry,
		cryptoKey: cryptoKey,
		logger:    logger,
	}
}

// RefreshExpiring finds providers whose tokens are about to expire (within
// lookahead) and refreshes them. Per-provider failures are logged + persisted
// as last_error; one bad provider doesn't block the batch.
//
// Called by the scheduled-tasks system every 5 minutes.
func (s *ProviderRefreshService) RefreshExpiring(ctx context.Context, lookahead time.Duration) error {
	providers, err := s.repo.ListExpiringSoon(ctx, time.Now().Add(lookahead))
	if err != nil {
		return fmt.Errorf("list expiring providers: %w", err)
	}

	for _, p := range providers {
		if err := s.refreshOne(ctx, p); err != nil {
			msg := err.Error()
			p.LastError = &msg
			_ = s.repo.Update(ctx, p)
			s.logger.Warn("provider.refresh_failed",
				"id", p.ID,
				"type", p.ProviderType,
				"email", p.AccountEmail,
				"err", err,
			)
		} else {
			s.logger.Info("provider.refreshed",
				"id", p.ID,
				"type", p.ProviderType,
				"email", p.AccountEmail,
			)
		}
	}
	return nil
}

// RefreshOne refreshes a single provider. Exposed for manual "refresh now"
// buttons in the admin UI (see StorageProviderHandler.Refresh).
func (s *ProviderRefreshService) RefreshOne(ctx context.Context, providerID int64) error {
	p, err := s.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	return s.refreshOne(ctx, p)
}

func (s *ProviderRefreshService) refreshOne(ctx context.Context, p *model.StorageProvider) error {
	driver, err := s.registry.Get(p.ProviderType)
	if err != nil {
		return err
	}

	creds, err := cloudstorage.DecryptCredentials(s.cryptoKey, p.CredentialsEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt credentials: %w", err)
	}

	// Dispatch by auth flow — service code doesn't know or care whether the
	// driver is password-auth or OAuth.
	var fresh *cloudstorage.Credentials
	switch d := driver.(type) {
	case cloudstorage.PasswordAuthDriver:
		if creds.Email == "" || creds.Password == "" {
			return fmt.Errorf("stored credentials missing email/password — re-add provider")
		}
		fresh, err = d.AuthenticatePassword(ctx, creds.Email, creds.Password)
	case cloudstorage.OAuthDriver:
		fresh, err = d.RefreshToken(ctx, creds)
	default:
		return fmt.Errorf("driver %s has no refresh mechanism", driver.Type())
	}
	if err != nil {
		return err
	}

	blob, err := cloudstorage.EncryptCredentials(s.cryptoKey, fresh)
	if err != nil {
		return fmt.Errorf("encrypt fresh credentials: %w", err)
	}

	// Optionally refresh account info so the admin UI's quota/VIP labels stay
	// accurate. Ignore errors here — a failing GetAccountInfo doesn't
	// invalidate the new session.
	provider, pErr := driver.NewProvider(fresh)
	if pErr == nil {
		if info, infoErr := provider.GetAccountInfo(ctx); infoErr == nil {
			p.AccountType = &info.AccountType
			if info.QuotaBytes > 0 {
				q := info.QuotaBytes
				p.QuotaBytes = &q
			}
			if info.UsedBytes > 0 {
				u := info.UsedBytes
				p.UsedBytes = &u
			}
			if !info.ExpiresAt.IsZero() {
				exp := info.ExpiresAt
				p.AccountExpiresAt = &exp
			}
		}
	}

	now := time.Now()
	p.CredentialsEncrypted = blob
	p.TokenExpiresAt = &fresh.ExpiresAt
	p.LastRefreshAt = &now
	p.LastCheckAt = &now
	p.LastError = nil

	return s.repo.Update(ctx, p)
}
