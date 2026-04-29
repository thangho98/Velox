package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/cloudstorage"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/service"
)

// cloudStreamTTL is the conservative lifetime reported to clients for cloud
// direct URLs. Fshare direct URLs empirically expire in minutes (possibly
// single-use); 5 min keeps clients reactively refreshing without stale hits.
const cloudStreamTTL = 5 * time.Minute

type StreamURLHandler struct {
	apiKeyStore *auth.APIKeyStore
	streamSvc   *service.StreamService

	// Cloud support (all nil when VELOX_CLOUD_SOURCES_ENABLED=false).
	libraryRepo  *repository.LibraryRepo
	mediaRepo    *repository.MediaRepo
	providerRepo *repository.StorageProviderRepo
	registry     *cloudstorage.Registry
	cryptoKey    []byte
}

func NewStreamURLHandler(apiKeyStore *auth.APIKeyStore, streamSvc *service.StreamService) *StreamURLHandler {
	return &StreamURLHandler{
		apiKeyStore: apiKeyStore,
		streamSvc:   streamSvc,
	}
}

// WithCloud attaches cloud-storage dependencies. Called at startup when the
// feature flag is enabled; passing nil registry or cryptoKey disables the
// cloud path (falls through to local api_key flow).
func (h *StreamURLHandler) WithCloud(
	libraryRepo *repository.LibraryRepo,
	mediaRepo *repository.MediaRepo,
	providerRepo *repository.StorageProviderRepo,
	registry *cloudstorage.Registry,
	cryptoKey []byte,
) *StreamURLHandler {
	h.libraryRepo = libraryRepo
	h.mediaRepo = mediaRepo
	h.providerRepo = providerRepo
	h.registry = registry
	h.cryptoKey = cryptoKey
	return h
}

func (h *StreamURLHandler) Create(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Try cloud path first; falls through to local flow when the library
	// isn't cloud-backed or the cloud feature is disabled.
	if h.cloudEnabled() {
		served, err := h.tryCloudURL(w, r, mediaID)
		if served {
			return
		}
		if err != nil {
			// Non-recoverable cloud error (already responded). Bail.
			return
		}
	}

	h.serveLocalURL(w, r, mediaID)
}

func (h *StreamURLHandler) cloudEnabled() bool {
	return h.registry != nil && h.libraryRepo != nil && h.providerRepo != nil && h.mediaRepo != nil && h.cryptoKey != nil
}

// tryCloudURL returns (served=true, nil) when it produced a response. Returns
// (false, nil) when this media isn't cloud-backed (caller should fall through).
// Returns (true, err) when a cloud error was sent to the client.
func (h *StreamURLHandler) tryCloudURL(w http.ResponseWriter, r *http.Request, mediaID int64) (bool, error) {
	ctx := r.Context()

	media, err := h.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "media not found")
			return true, err
		}
		return false, nil // fall through to local flow
	}

	library, err := h.libraryRepo.GetByID(ctx, media.LibraryID)
	if err != nil || !library.IsCloudLibrary() {
		return false, nil
	}

	fileID, err := parseCloudMediaFileID(ctx, h.streamSvc, mediaID)
	if err != nil || fileID == "" {
		return false, nil
	}

	sp, err := h.providerRepo.GetByID(ctx, *library.StorageProviderID)
	if err != nil {
		respondError(w, http.StatusFailedDependency, "storage provider not found")
		return true, err
	}

	driver, err := h.registry.Get(sp.ProviderType)
	if err != nil {
		respondError(w, http.StatusFailedDependency, err.Error())
		return true, err
	}

	creds, err := cloudstorage.DecryptCredentials(h.cryptoKey, sp.CredentialsEncrypted)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return true, err
	}

	provider, err := driver.NewProvider(creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return true, err
	}

	preToken := creds.AccessToken
	downloadURL, err := provider.GetDownloadURL(ctx, fileID)
	if err != nil {
		statusForCloudError(w, err)
		return true, err
	}

	// Persist silently-relogged-in credentials so other code paths see the
	// latest session. Cheap: no-op if token unchanged.
	if updated := provider.Session(); updated != nil && updated.AccessToken != preToken {
		if blob, encErr := cloudstorage.EncryptCredentials(h.cryptoKey, updated); encErr == nil {
			_ = h.providerRepo.UpdateCredentials(ctx, sp.ID, blob, updated.ExpiresAt)
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"direct_url":    downloadURL,
		"provider_type": sp.ProviderType,
		"direct_cdn":    true,
		"expires_in":    int(cloudStreamTTL.Seconds()),
	})
	return true, nil
}

func (h *StreamURLHandler) serveLocalURL(w http.ResponseWriter, r *http.Request, mediaID int64) {
	userID, isAdmin, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	streamTTL := auth.MinStreamAPIKeyTTL
	if h.streamSvc != nil {
		if primaryFile, err := h.streamSvc.GetPrimaryFile(r.Context(), mediaID, 0); err == nil && primaryFile != nil {
			streamTTL = auth.StreamAPIKeyTTL(primaryFile.Duration)
		}
	}

	apiKey := h.apiKeyStore.Generate(userID, isAdmin, streamTTL)
	streamSessionID := newStreamSessionID()
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := resolvePublicHost(r)

	respondJSON(w, http.StatusOK, map[string]any{
		"direct_url":        fmt.Sprintf("%s://%s/api/stream/%d?api_key=%s&%s=%s", scheme, host, mediaID, apiKey, streamSessionQueryKey, streamSessionID),
		"hls_url":           fmt.Sprintf("%s://%s/api/stream/%d/hls/master.m3u8?api_key=%s&%s=%s", scheme, host, mediaID, apiKey, streamSessionQueryKey, streamSessionID),
		"stream_session_id": streamSessionID,
		"api_key":           apiKey,
		"expires_in":        int(streamTTL.Seconds()),
		"provider_type":     "local",
		"direct_cdn":        false,
	})
}

// parseCloudMediaFileID resolves the primary media_file for a media row and
// extracts its cloud native ID (e.g. fshare linkcode). Returns "" when the
// primary file isn't a cloud path.
func parseCloudMediaFileID(ctx context.Context, svc *service.StreamService, mediaID int64) (string, error) {
	if svc == nil {
		return "", nil
	}
	primaryFile, err := svc.GetPrimaryFile(ctx, mediaID, 0)
	if err != nil || primaryFile == nil {
		return "", err
	}
	_, nativeID, ok := scanner.ParseCloudPath(primaryFile.FilePath)
	if !ok {
		return "", nil
	}
	return nativeID, nil
}

func statusForCloudError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, cloudstorage.ErrSessionExpired), errors.Is(err, cloudstorage.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, "cloud session expired — re-authenticate the storage provider")
	case errors.Is(err, cloudstorage.ErrAccountNotPremium):
		respondError(w, http.StatusPaymentRequired, "cloud account tier does not allow downloads (VIP required)")
	case errors.Is(err, cloudstorage.ErrRateLimit):
		respondError(w, http.StatusTooManyRequests, "cloud provider is rate-limiting; try again in a minute")
	case errors.Is(err, cloudstorage.ErrFileNotFound):
		respondError(w, http.StatusNotFound, "file no longer available on cloud provider")
	default:
		respondError(w, http.StatusBadGateway, err.Error())
	}
}

// ResolveCloudDirectURL translates a cloud native path (e.g. fshare://...) into a
// playable HTTP URL by invoking the cloud driver.
func ResolveCloudDirectURL(ctx context.Context, mediaID int64, mf *model.MediaFile, libraryRepo *repository.LibraryRepo, mediaRepo *repository.MediaRepo, providerRepo *repository.StorageProviderRepo, registry *cloudstorage.Registry, cryptoKey []byte) (string, error) {
	if registry == nil {
		return mf.FilePath, nil // cloud disabled, fallback to raw path
	}

	_, nativeID, ok := scanner.ParseCloudPath(mf.FilePath)
	if !ok {
		return mf.FilePath, nil // already local path or HTTP URL
	}

	media, err := mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return "", fmt.Errorf("fetch media: %w", err)
	}

	library, err := libraryRepo.GetByID(ctx, media.LibraryID)
	if err != nil || !library.IsCloudLibrary() {
		return "", errors.New("not a cloud library")
	}

	sp, err := providerRepo.GetByID(ctx, *library.StorageProviderID)
	if err != nil {
		return "", fmt.Errorf("provider match: %w", err)
	}

	driver, err := registry.Get(sp.ProviderType)
	if err != nil {
		return "", err
	}

	creds, err := cloudstorage.DecryptCredentials(cryptoKey, sp.CredentialsEncrypted)
	if err != nil {
		return "", err
	}

	provider, err := driver.NewProvider(creds)
	if err != nil {
		return "", err
	}

	preToken := creds.AccessToken
	downloadURL, err := provider.GetDownloadURL(ctx, nativeID)
	if err != nil {
		return "", err
	}

	if updated := provider.Session(); updated != nil && updated.AccessToken != preToken {
		if blob, encErr := cloudstorage.EncryptCredentials(cryptoKey, updated); encErr == nil {
			_ = providerRepo.UpdateCredentials(ctx, sp.ID, blob, updated.ExpiresAt)
		}
	}

	return downloadURL, nil
}

// CloudProbe runs ffprobe on a cloud media file and persists the discovered
// metadata (codec, resolution, audio tracks, embedded subtitles) into the DB.
// Requires admin. Responds 200 with the probe summary on success.
func (h *StreamURLHandler) CloudProbe(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if !h.cloudEnabled() {
		respondError(w, http.StatusNotFound, "cloud storage not configured")
		return
	}

	mf, err := h.streamSvc.GetPrimaryFile(r.Context(), mediaID, 0)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "media not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, _, ok := scanner.ParseCloudPath(mf.FilePath); !ok {
		respondError(w, http.StatusBadRequest, "not a cloud media file")
		return
	}

	if err := h.streamSvc.ProbeAndUpdateCloudMetadata(r.Context(), mf); err != nil {
		statusForCloudError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"media_file_id": mf.ID,
		"video_codec":   mf.VideoCodec,
		"width":         mf.Width,
		"height":        mf.Height,
		"duration":      mf.Duration,
	})
}
