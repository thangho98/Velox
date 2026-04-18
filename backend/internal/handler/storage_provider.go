package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/cloudstorage"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// StorageProviderHandler exposes CRUD + auth flow for cloud storage providers.
// Disabled (nil registry) when VELOX_CLOUD_SOURCES_ENABLED=false.
type StorageProviderHandler struct {
	registry     *cloudstorage.Registry
	providerRepo *repository.StorageProviderRepo
	libraryRepo  *repository.LibraryRepo
	cryptoKey    []byte
}

func NewStorageProviderHandler(
	registry *cloudstorage.Registry,
	providerRepo *repository.StorageProviderRepo,
	libraryRepo *repository.LibraryRepo,
	cryptoKey []byte,
) *StorageProviderHandler {
	return &StorageProviderHandler{
		registry:     registry,
		providerRepo: providerRepo,
		libraryRepo:  libraryRepo,
		cryptoKey:    cryptoKey,
	}
}

// Enabled reports whether cloud storage is wired up.
func (h *StorageProviderHandler) Enabled() bool {
	return h != nil && h.registry != nil
}

// providerView is the outward shape returned to the admin UI. It deliberately
// omits credentials_encrypted and derives health status on the server so the
// frontend doesn't need the same time-math logic.
type providerView struct {
	ID               int64      `json:"id"`
	ProviderType     string     `json:"provider_type"`
	DisplayName      string     `json:"display_name"`
	AccountEmail     string     `json:"account_email"`
	AccountType      *string    `json:"account_type,omitempty"`
	QuotaBytes       *int64     `json:"quota_bytes,omitempty"`
	UsedBytes        *int64     `json:"used_bytes,omitempty"`
	AccountExpiresAt *time.Time `json:"account_expires_at,omitempty"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	LastRefreshAt    *time.Time `json:"last_refresh_at,omitempty"`
	LastError        *string    `json:"last_error,omitempty"`
	Health           string     `json:"health"`
	CreatedAt        time.Time  `json:"created_at"`
}

func toView(p *model.StorageProvider, now time.Time) providerView {
	return providerView{
		ID:               p.ID,
		ProviderType:     p.ProviderType,
		DisplayName:      p.DisplayName,
		AccountEmail:     p.AccountEmail,
		AccountType:      p.AccountType,
		QuotaBytes:       p.QuotaBytes,
		UsedBytes:        p.UsedBytes,
		AccountExpiresAt: p.AccountExpiresAt,
		TokenExpiresAt:   p.TokenExpiresAt,
		LastRefreshAt:    p.LastRefreshAt,
		LastError:        p.LastError,
		Health:           string(p.ComputeHealth(now)),
		CreatedAt:        p.CreatedAt,
	}
}

// --- Driver catalog ---

// ListDrivers returns registered drivers (for the admin UI type dropdown).
func (h *StorageProviderHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	respondJSON(w, http.StatusOK, h.registry.List())
}

// --- Provider CRUD ---

type createProviderRequest struct {
	ProviderType string `json:"provider_type"`
	DisplayName  string `json:"display_name"`
	Auth         struct {
		Flow     string `json:"flow"` // "password" | "oauth"
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
		Code     string `json:"code,omitempty"`         // OAuth only
		Redirect string `json:"redirect_uri,omitempty"` // OAuth only
	} `json:"auth"`
}

// Create logs in the provider via its driver, then persists an encrypted
// credentials row. Failing VIP accounts (fshare Fee tier) are rejected at
// create time rather than at first playback.
func (h *StorageProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.ProviderType == "" || req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "provider_type and display_name are required")
		return
	}

	driver, err := h.registry.Get(req.ProviderType)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	var creds *cloudstorage.Credentials
	switch req.Auth.Flow {
	case "password":
		pwDriver, ok := driver.(cloudstorage.PasswordAuthDriver)
		if !ok {
			respondError(w, http.StatusBadRequest, "driver does not support password auth")
			return
		}
		if req.Auth.Email == "" || req.Auth.Password == "" {
			respondError(w, http.StatusBadRequest, "email and password are required")
			return
		}
		creds, err = pwDriver.AuthenticatePassword(ctx, req.Auth.Email, req.Auth.Password)
	case "oauth":
		oauthDriver, ok := driver.(cloudstorage.OAuthDriver)
		if !ok {
			respondError(w, http.StatusBadRequest, "driver does not support oauth")
			return
		}
		creds, err = oauthDriver.ExchangeCode(ctx, req.Auth.Code, req.Auth.Redirect)
	default:
		respondError(w, http.StatusBadRequest, "auth.flow must be 'password' or 'oauth'")
		return
	}
	if err != nil {
		writeCloudError(w, err)
		return
	}

	// Fetch AccountInfo eagerly so we can gate non-VIP accounts + populate the
	// admin UI card without a second round-trip.
	provider, err := driver.NewProvider(creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	info, err := provider.GetAccountInfo(ctx)
	if err != nil {
		writeCloudError(w, err)
		return
	}
	if driver.Type() == "fshare" && info.AccountType != "Vip" {
		respondError(w, http.StatusPaymentRequired, "fshare VIP account required for cloud streaming")
		return
	}

	blob, err := cloudstorage.EncryptCredentials(h.cryptoKey, creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sp := &model.StorageProvider{
		UserID:               userID,
		ProviderType:         driver.Type(),
		DisplayName:          req.DisplayName,
		AccountEmail:         info.Email,
		CredentialsEncrypted: blob,
		AccountType:          &info.AccountType,
	}
	if info.QuotaBytes > 0 {
		q := info.QuotaBytes
		sp.QuotaBytes = &q
	}
	if !info.ExpiresAt.IsZero() {
		exp := info.ExpiresAt
		sp.AccountExpiresAt = &exp
	}
	if !creds.ExpiresAt.IsZero() {
		exp := creds.ExpiresAt
		sp.TokenExpiresAt = &exp
	}
	now := time.Now()
	sp.LastRefreshAt = &now
	sp.LastCheckAt = &now

	if err := h.providerRepo.Create(ctx, sp); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, toView(sp, now))
}

// List returns all providers for the calling user.
func (h *StorageProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	providers, err := h.providerRepo.ListByUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	now := time.Now()
	out := make([]providerView, 0, len(providers))
	for _, p := range providers {
		out = append(out, toView(p, now))
	}
	respondJSON(w, http.StatusOK, out)
}

// Delete removes a provider and nulls out storage_provider_id on its libraries.
func (h *StorageProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.providerRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "provider not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Refresh re-runs the auth flow and updates the stored credentials. Useful for
// manually kicking a stale provider without waiting for the scheduler.
func (h *StorageProviderHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctx := r.Context()

	sp, err := h.providerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "provider not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	driver, err := h.registry.Get(sp.ProviderType)
	if err != nil {
		respondError(w, http.StatusFailedDependency, err.Error())
		return
	}
	creds, err := cloudstorage.DecryptCredentials(h.cryptoKey, sp.CredentialsEncrypted)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	var fresh *cloudstorage.Credentials
	switch d := driver.(type) {
	case cloudstorage.PasswordAuthDriver:
		if creds.Email == "" || creds.Password == "" {
			respondError(w, http.StatusFailedDependency, "stored credentials missing email/password — re-add provider")
			return
		}
		fresh, err = d.AuthenticatePassword(ctx, creds.Email, creds.Password)
	case cloudstorage.OAuthDriver:
		fresh, err = d.RefreshToken(ctx, creds)
	default:
		respondError(w, http.StatusNotImplemented, "driver has no refresh mechanism")
		return
	}
	if err != nil {
		writeCloudError(w, err)
		return
	}

	blob, err := cloudstorage.EncryptCredentials(h.cryptoKey, fresh)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.providerRepo.UpdateCredentials(ctx, sp.ID, blob, fresh.ExpiresAt); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Re-fetch the refreshed row so the response reflects new token_expires_at.
	if updated, err := h.providerRepo.GetByID(ctx, sp.ID); err == nil {
		sp = updated
	}
	respondJSON(w, http.StatusOK, toView(sp, time.Now()))
}

// ValidateURL lets the admin UI test a folder URL against a saved provider.
// Returns the resolved folder ID + how many items are inside.
func (h *StorageProviderHandler) ValidateURL(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		respondError(w, http.StatusServiceUnavailable, "cloud storage feature disabled")
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	ctx := r.Context()
	sp, err := h.providerRepo.GetByID(ctx, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "provider not found")
		return
	}
	driver, err := h.registry.Get(sp.ProviderType)
	if err != nil {
		respondError(w, http.StatusFailedDependency, err.Error())
		return
	}
	creds, err := cloudstorage.DecryptCredentials(h.cryptoKey, sp.CredentialsEncrypted)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	provider, err := driver.NewProvider(creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	folderID, err := provider.ParseFolderURL(body.URL)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := provider.ListFolder(ctx, folderID)
	if err != nil {
		writeCloudError(w, err)
		return
	}

	// Show the first 3 names as a sanity check for the admin.
	sample := make([]string, 0, 3)
	for i := 0; i < len(items) && i < 3; i++ {
		sample = append(sample, items[i].Name)
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"folder_id":    folderID,
		"item_count":   len(items),
		"sample_names": sample,
	})
}

// writeCloudError translates cloudstorage errors to HTTP responses. Shared with
// stream_url.go for consistency.
func writeCloudError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, cloudstorage.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, cloudstorage.ErrSessionExpired):
		respondError(w, http.StatusUnauthorized, "cloud session expired")
	case errors.Is(err, cloudstorage.ErrAccountNotPremium):
		respondError(w, http.StatusPaymentRequired, err.Error())
	case errors.Is(err, cloudstorage.ErrRateLimit):
		respondError(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, cloudstorage.ErrInvalidURL):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, cloudstorage.ErrFileNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	default:
		respondError(w, http.StatusBadGateway, err.Error())
	}
}
