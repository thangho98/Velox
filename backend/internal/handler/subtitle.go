package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/pkg/subtitle"
)

// SubtitleHandler handles subtitle HTTP requests
type SubtitleHandler struct {
	svc           *service.SubtitleService
	mediaFileRepo *repository.MediaFileRepo // resolves video path for embedded sub extraction
	settingsRepo  *repository.AppSettingsRepo
	subtitleCache string // base dir for extracted VTT files; e.g. ~/.velox/subtitles
}

func NewSubtitleHandler(svc *service.SubtitleService, mediaFileRepo *repository.MediaFileRepo, settingsRepo *repository.AppSettingsRepo, subtitleCache string) *SubtitleHandler {
	return &SubtitleHandler{
		svc:           svc,
		mediaFileRepo: mediaFileRepo,
		settingsRepo:  settingsRepo,
		subtitleCache: subtitleCache,
	}
}

// ListByMediaFile returns all subtitles for a media file
func (h *SubtitleHandler) ListByMediaFile(w http.ResponseWriter, r *http.Request) {
	mediaFileID, err := parseID(r, "media_file_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media_file_id")
		return
	}

	subtitles, err := h.svc.ListByMediaFile(r.Context(), mediaFileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, subtitles)
}

// Get returns a subtitle by ID
func (h *SubtitleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	subtitle, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "subtitle not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, subtitle)
}

// CreateRequest represents a subtitle creation request
type CreateSubtitleRequest struct {
	MediaFileID int64  `json:"media_file_id"`
	Language    string `json:"language"`
	Codec       string `json:"codec"`
	Title       string `json:"title"`
	IsEmbedded  bool   `json:"is_embedded"`
	StreamIndex int    `json:"stream_index"`
	FilePath    string `json:"file_path"`
	IsForced    bool   `json:"is_forced"`
	IsDefault   bool   `json:"is_default"`
	IsSDH       bool   `json:"is_sdh"`
}

// Create creates a new subtitle
func (h *SubtitleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubtitleRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subtitle := &model.Subtitle{
		MediaFileID: req.MediaFileID,
		Language:    req.Language,
		Codec:       req.Codec,
		Title:       req.Title,
		IsEmbedded:  req.IsEmbedded,
		StreamIndex: req.StreamIndex,
		FilePath:    req.FilePath,
		IsForced:    req.IsForced,
		IsDefault:   req.IsDefault,
		IsSDH:       req.IsSDH,
	}

	if err := h.svc.Create(r.Context(), subtitle); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, subtitle)
}

// UpdateRequest represents a subtitle update request
type UpdateSubtitleRequest struct {
	Language    string `json:"language"`
	Codec       string `json:"codec"`
	Title       string `json:"title"`
	IsEmbedded  bool   `json:"is_embedded"`
	StreamIndex int    `json:"stream_index"`
	FilePath    string `json:"file_path"`
	IsForced    bool   `json:"is_forced"`
	IsDefault   bool   `json:"is_default"`
	IsSDH       bool   `json:"is_sdh"`
}

// Update updates a subtitle
func (h *SubtitleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	subtitle, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "subtitle not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req UpdateSubtitleRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subtitle.Language = req.Language
	subtitle.Codec = req.Codec
	subtitle.Title = req.Title
	subtitle.IsEmbedded = req.IsEmbedded
	subtitle.StreamIndex = req.StreamIndex
	subtitle.FilePath = req.FilePath
	subtitle.IsForced = req.IsForced
	subtitle.IsDefault = req.IsDefault
	subtitle.IsSDH = req.IsSDH

	if err := h.svc.Update(r.Context(), subtitle); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, subtitle)
}

// Delete deletes a subtitle
func (h *SubtitleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

// Serve serves a subtitle file as WebVTT.
// External .vtt → served directly; external .srt → converted to VTT on the fly.
// Embedded subtitles return 501 until the Phase 03 extractor is implemented.
func (h *SubtitleHandler) Serve(w http.ResponseWriter, r *http.Request) {
	subtitleID, err := parseID(r, "subtitle_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtitle_id")
		return
	}

	sub, err := h.svc.Get(r.Context(), subtitleID)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "subtitle not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if sub.IsEmbedded {
		// Look up the video file to get its path, then extract the subtitle stream.
		mf, err := h.mediaFileRepo.GetByID(r.Context(), sub.MediaFileID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("looking up media file: %v", err))
			return
		}

		cacheDir := filepath.Join(h.subtitleCache, fmt.Sprintf("%d", sub.MediaFileID))
		vttPath, err := subtitle.ExtractSubtitle(mf.FilePath, sub.StreamIndex, cacheDir)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("extracting subtitle: %v", err))
			return
		}

		data, err := os.ReadFile(vttPath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "reading extracted subtitle")
			return
		}
		_, _ = w.Write(data)
		return
	}

	if sub.FilePath == "" {
		respondError(w, http.StatusNotFound, "subtitle file path not set")
		return
	}

	data, err := os.ReadFile(sub.FilePath)
	if err != nil {
		respondError(w, http.StatusNotFound, "subtitle file not found on disk")
		return
	}

	codec := strings.ToLower(sub.Codec)
	if codec == "subrip" || codec == "srt" {
		_, _ = w.Write(subtitle.SRTToVTT(data))
		return
	}
	_, _ = w.Write(data)
}

// SetDefault sets a subtitle as default
func (h *SubtitleHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	mediaFileID, err := parseID(r, "media_file_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media_file_id")
		return
	}

	subtitleID, err := parseID(r, "subtitle_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtitle_id")
		return
	}

	if err := h.svc.SetDefault(r.Context(), mediaFileID, subtitleID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "default set"})
}

// Translate translates a subtitle to a target language using DeepL (primary) or Google (fallback).
// POST /api/subtitles/{id}/translate
func (h *SubtitleHandler) Translate(w http.ResponseWriter, r *http.Request) {
	subtitleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtitle id")
		return
	}

	var req struct {
		TargetLanguage string `json:"target_language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TargetLanguage == "" {
		respondError(w, http.StatusBadRequest, "target_language required")
		return
	}

	// Get DeepL API key from settings (optional — falls back to Google if empty)
	deeplKey, _ := h.settingsRepo.Get(r.Context(), model.SettingDeepLAPIKey)

	sub, err := h.svc.TranslateSubtitle(r.Context(), subtitleID, req.TargetLanguage, deeplKey, h.subtitleCache)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sub)
}

// AudioTrackHandler handles audio track HTTP requests
type AudioTrackHandler struct {
	svc *service.AudioTrackService
}

func NewAudioTrackHandler(svc *service.AudioTrackService) *AudioTrackHandler {
	return &AudioTrackHandler{svc: svc}
}

// ListByMediaFile returns all audio tracks for a media file
func (h *AudioTrackHandler) ListByMediaFile(w http.ResponseWriter, r *http.Request) {
	mediaFileID, err := parseID(r, "media_file_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media_file_id")
		return
	}

	tracks, err := h.svc.ListByMediaFile(r.Context(), mediaFileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tracks)
}

// Get returns an audio track by ID
func (h *AudioTrackHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	track, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "audio track not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, track)
}

// CreateAudioTrackRequest represents an audio track creation request
type CreateAudioTrackRequest struct {
	MediaFileID   int64  `json:"media_file_id"`
	StreamIndex   int    `json:"stream_index"`
	Codec         string `json:"codec"`
	Language      string `json:"language"`
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
	Bitrate       int    `json:"bitrate"`
	Title         string `json:"title"`
	IsDefault     bool   `json:"is_default"`
}

// Create creates a new audio track
func (h *AudioTrackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAudioTrackRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	track := &model.AudioTrack{
		MediaFileID:   req.MediaFileID,
		StreamIndex:   req.StreamIndex,
		Codec:         req.Codec,
		Language:      req.Language,
		Channels:      req.Channels,
		ChannelLayout: req.ChannelLayout,
		Bitrate:       req.Bitrate,
		Title:         req.Title,
		IsDefault:     req.IsDefault,
	}

	if err := h.svc.Create(r.Context(), track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, track)
}

// UpdateAudioTrackRequest represents an audio track update request
type UpdateAudioTrackRequest struct {
	Codec         string `json:"codec"`
	Language      string `json:"language"`
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
	Bitrate       int    `json:"bitrate"`
	Title         string `json:"title"`
	IsDefault     bool   `json:"is_default"`
}

// Update updates an audio track
func (h *AudioTrackHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	track, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "audio track not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req UpdateAudioTrackRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	track.Codec = req.Codec
	track.Language = req.Language
	track.Channels = req.Channels
	track.ChannelLayout = req.ChannelLayout
	track.Bitrate = req.Bitrate
	track.Title = req.Title
	track.IsDefault = req.IsDefault

	if err := h.svc.Update(r.Context(), track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, track)
}

// Delete deletes an audio track
func (h *AudioTrackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

// SetDefault sets an audio track as default
func (h *AudioTrackHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	mediaFileID, err := parseID(r, "media_file_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media_file_id")
		return
	}

	trackID, err := parseID(r, "track_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid track_id")
		return
	}

	if err := h.svc.SetDefault(r.Context(), mediaFileID, trackID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "default set"})
}
