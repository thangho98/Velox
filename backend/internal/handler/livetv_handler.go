package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/middleware"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

// watchMarkCooldown limits DB writes when the player re-fetches the manifest
// (HLS manifests reload every few seconds). One bump per user+channel within
// this window is plenty to keep the "recently watched" order fresh.
const watchMarkCooldown = 5 * time.Minute

type LiveTVHandler struct {
	repo        *repository.LiveTVRepo
	svc         *service.LiveTVService
	authSvc     *service.AuthService
	proxySecret []byte
	proxyClient *http.Client

	watchMarkMu   sync.Mutex
	watchMarkSeen map[watchMarkKey]time.Time
}

type watchMarkKey struct {
	userID    int64
	channelID int64
}

func NewLiveTVHandler(repo *repository.LiveTVRepo, svc *service.LiveTVService, authSvc *service.AuthService, proxySecret []byte) *LiveTVHandler {
	return &LiveTVHandler{
		repo:          repo,
		svc:           svc,
		authSvc:       authSvc,
		proxySecret:   proxySecret,
		proxyClient:   &http.Client{Timeout: 30 * time.Second},
		watchMarkSeen: make(map[watchMarkKey]time.Time),
	}
}

// shouldMarkWatched returns true when the (user, channel) pair hasn't been
// recorded within the cooldown window. Updates the seen timestamp atomically.
func (h *LiveTVHandler) shouldMarkWatched(userID, channelID int64) bool {
	key := watchMarkKey{userID: userID, channelID: channelID}
	now := time.Now()
	h.watchMarkMu.Lock()
	defer h.watchMarkMu.Unlock()
	if last, ok := h.watchMarkSeen[key]; ok && now.Sub(last) < watchMarkCooldown {
		return false
	}
	h.watchMarkSeen[key] = now
	// Opportunistic cleanup: prune entries older than 2× cooldown to avoid
	// unbounded growth over long uptimes.
	for k, t := range h.watchMarkSeen {
		if now.Sub(t) > 2*watchMarkCooldown {
			delete(h.watchMarkSeen, k)
		}
	}
	return true
}

func (h *LiveTVHandler) RegisterRoutes(mux *http.ServeMux) {
	// Admin routes
	mux.Handle("GET /api/livetv/playlists", middleware.RequireAdmin(http.HandlerFunc(h.handleGetPlaylists)))
	mux.Handle("POST /api/livetv/playlists", middleware.RequireAdmin(http.HandlerFunc(h.handleCreatePlaylist)))
	mux.Handle("PUT /api/livetv/playlists/{id}", middleware.RequireAdmin(http.HandlerFunc(h.handleUpdatePlaylist)))
	mux.Handle("DELETE /api/livetv/playlists/{id}", middleware.RequireAdmin(http.HandlerFunc(h.handleDeletePlaylist)))
	mux.Handle("POST /api/livetv/playlists/{id}/sync", middleware.RequireAdmin(http.HandlerFunc(h.handleSyncPlaylist)))

	// User routes
	mux.HandleFunc("GET /api/livetv/channels", h.handleGetChannels)
	mux.HandleFunc("GET /api/livetv/channels/recent", h.handleGetRecentChannels)
	mux.HandleFunc("GET /api/livetv/channels/search", h.handleSearchChannels)
	mux.HandleFunc("GET /api/livetv/channels/{id}", h.handleGetChannel)
	mux.HandleFunc("GET /api/livetv/channels/{id}/epg", h.handleGetChannelEPG)
	mux.Handle("POST /api/livetv/channels/{id}/toggle-hidden", middleware.RequireAdmin(http.HandlerFunc(h.handleToggleChannelHidden)))
	mux.HandleFunc("GET /api/livetv/groups", h.handleGetGroups)

	// Public stream proxy — the <video> element loads the entry URL, which then
	// fetches the upstream manifest and rewrites every child URL to route back
	// through /api/livetv/proxy with an HMAC signature.
	mux.HandleFunc("GET /api/livetv/stream/{channelId}", h.handleStreamEntry)
	mux.HandleFunc("GET /api/livetv/proxy", h.handleProxy)
}

func (h *LiveTVHandler) handleGetChannel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	channel, err := h.repo.GetChannelByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

func (h *LiveTVHandler) handleGetPlaylists(w http.ResponseWriter, r *http.Request) {
	playlists, err := h.repo.GetPlaylists(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, playlists)
}

func (h *LiveTVHandler) handleCreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var name, urlStr string
	var syncInterval int

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}
		name = r.FormValue("name")
		syncInterval, _ = strconv.Atoi(r.FormValue("sync_interval_hours"))

		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			dir := filepath.Join("data", "livetv_playlists")
			os.MkdirAll(dir, 0755)

			destPath := filepath.Join(dir, fmt.Sprintf("playlist_%d_%s", time.Now().Unix(), header.Filename))
			destFile, err := os.Create(destPath)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "failed to save file")
				return
			}
			defer destFile.Close()
			io.Copy(destFile, file)

			absPath, _ := filepath.Abs(destPath)
			urlStr = absPath
		} else {
			urlStr = r.FormValue("url")
		}
	} else {
		var payload struct {
			Name              string `json:"name"`
			URL               string `json:"url"`
			SyncIntervalHours int    `json:"sync_interval_hours"`
		}
		if err := parseJSON(r, &payload); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		name = payload.Name
		urlStr = payload.URL
		syncInterval = payload.SyncIntervalHours
	}

	playlist := &model.LivePlaylist{
		Name:              name,
		URL:               urlStr,
		SyncIntervalHours: syncInterval,
		IsActive:          true,
	}

	if err := h.repo.CreatePlaylist(r.Context(), playlist); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Trigger async sync right after creation
	go h.svc.SyncPlaylist(contextBackgroundFrom(r), playlist.ID)

	respondJSON(w, http.StatusOK, playlist)
}

func (h *LiveTVHandler) handleUpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var payload struct {
		Name              string `json:"name"`
		URL               string `json:"url"`
		SyncIntervalHours int    `json:"sync_interval_hours"`
		IsActive          bool   `json:"is_active"`
	}
	if err := parseJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	playlist := &model.LivePlaylist{
		ID:                id,
		Name:              payload.Name,
		URL:               payload.URL,
		SyncIntervalHours: payload.SyncIntervalHours,
		IsActive:          payload.IsActive,
	}

	if err := h.repo.UpdatePlaylist(r.Context(), playlist); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "playlist not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, playlist)
}

func (h *LiveTVHandler) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.DeletePlaylist(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "playlist not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *LiveTVHandler) handleSyncPlaylist(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Verify the playlist exists first
	_, err = h.repo.GetPlaylist(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "playlist not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Run sync asynchronously
	go h.svc.SyncPlaylist(contextBackgroundFrom(r), id)

	respondJSON(w, http.StatusOK, map[string]string{"status": "sync_started"})
}

func (h *LiveTVHandler) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	limit := parseIntQuery(r, "limit", 100)
	offset := parseIntQuery(r, "offset", 0)
	group := r.URL.Query().Get("group")

	channels, err := h.repo.GetChannels(r.Context(), group, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, channels)
}

func (h *LiveTVHandler) handleGetRecentChannels(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit := parseIntQuery(r, "limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	channels, err := h.repo.ListRecentChannels(r.Context(), userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if channels == nil {
		channels = []*model.LiveChannel{}
	}
	respondJSON(w, http.StatusOK, channels)
}

func (h *LiveTVHandler) handleSearchChannels(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		respondJSON(w, http.StatusOK, []any{})
		return
	}
	limit := parseIntQuery(r, "limit", 50)

	channels, err := h.repo.SearchChannels(r.Context(), q, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, channels)
}

func (h *LiveTVHandler) handleGetGroups(w http.ResponseWriter, r *http.Request) {
	// TODO: GetGroups in Repo needs to be added
	groups, err := h.repo.GetChannelGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

// contextBackgroundFrom safely extracts a detached background context to run async tasks
func contextBackgroundFrom(r *http.Request) context.Context {
	return context.Background()
}

// mockEPGTemplates is a pool of 3-program blocks. channel_id picks one deterministically
// so different channels show different "Up Next" schedules until a real XMLTV parser lands.
var mockEPGTemplates = [][3]struct {
	title, description string
}{
	{
		{"La Liga — Barca vs Sevilla", "Live Spanish Football"},
		{"SportsCenter — Late Night", "Late night sports news and highlights."},
		{"Boxing: Fury vs. Usyk (Encore)", "Heavyweight unification bout replay."},
	},
	{
		{"Morning Cartoons", "Classic animated shorts to start the day."},
		{"Anime Spotlight", "Featured weekly anime episode."},
		{"Retro Saturday", "Vintage animated favorites."},
	},
	{
		{"Breaking News Live", "Latest headlines from around the world."},
		{"Market Watch", "Financial markets and business updates."},
		{"World Report", "In-depth international coverage."},
	},
	{
		{"Blockbuster Movie", "Tonight's featured film."},
		{"Cinema Classics", "An acclaimed classic from the archives."},
		{"Late Show Feature", "Independent and arthouse cinema."},
	},
	{
		{"Chart Toppers", "This week's biggest hits."},
		{"Acoustic Sessions", "Stripped-back live performances."},
		{"Electronic Nights", "DJ-curated electronic tracks."},
	},
	{
		{"Nature Uncovered", "Wildlife and ecosystem explorations."},
		{"History Files", "Deep dives into major historical events."},
		{"Science Frontier", "Latest discoveries and innovations."},
	},
	{
		{"Stand-Up Special", "Tonight's featured comedian."},
		{"Sitcom Marathon", "Back-to-back comedy favorites."},
		{"Late Night Laughs", "Monologues and sketches."},
	},
	{
		{"Kids Morning Block", "Age-appropriate shows for early risers."},
		{"Family Adventure", "Stories the whole family can enjoy."},
		{"Bedtime Stories", "Calming tales for little viewers."},
	},
}

func (h *LiveTVHandler) handleGetChannelEPG(w http.ResponseWriter, r *http.Request) {
	// Return empty EPG until real XMLTV parser lands
	// channelID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	// if err != nil {
	// 	respondError(w, http.StatusBadRequest, "invalid channel id")
	// 	return
	// }

	programs := []model.LiveProgram{}

	respondJSON(w, http.StatusOK, programs)
}

func (h *LiveTVHandler) handleToggleChannelHidden(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	isHidden, err := h.svc.ToggleChannelHidden(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"is_hidden": isHidden,
	})
}
