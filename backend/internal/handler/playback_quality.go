package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
)

// QualityOption represents an available quality for the frontend quality selector
type QualityOption struct {
	Height      int    `json:"height"`                 // e.g. 1080, 720, 480
	Label       string `json:"label"`                  // e.g. "1080p", "720p"
	Instant     bool   `json:"instant"`                // true = pretranscode ready (⚡)
	Source      string `json:"source"`                 // "original", "pretranscode", "transcode"
	BitrateKbps int    `json:"bitrate_kbps,omitempty"` // estimated bitrate
}

// buildQualityOptions returns the list of selectable qualities for the frontend.
// Includes: Original, pretranscode profiles (⚡ instant), and standard transcode tiers.
// Uses a single JOIN query to avoid N+1 DB lookups.
func (h *PlaybackHandler) buildQualityOptions(ctx context.Context, file model.MediaFile) []QualityOption {
	sourceHeight := file.Height
	var options []QualityOption

	// 1. Original quality
	options = append(options, QualityOption{
		Height:      sourceHeight,
		Label:       fmt.Sprintf("Original (%dp)", sourceHeight),
		Instant:     true,
		Source:      "original",
		BitrateKbps: file.Bitrate / 1000,
	})

	// 2. Collect pretranscode ready profiles (single JOIN query)
	ptHeights := map[int]bool{}
	if results, err := h.streamSvc.FindAllPretranscodesWithProfiles(ctx, file.ID); err == nil {
		for _, r := range results {
			if r.Profile.Height <= sourceHeight {
				ptHeights[r.Profile.Height] = true
				if r.Profile.Height == sourceHeight {
					continue
				}
				options = append(options, QualityOption{
					Height:      r.Profile.Height,
					Label:       r.Profile.Name,
					Instant:     true,
					Source:      "pretranscode",
					BitrateKbps: r.Profile.VideoBitrate + r.Profile.AudioBitrate,
				})
			}
		}
	}

	// 3. Standard transcode tiers (only those below source and not covered by pretranscode)
	transcodeTiers := []struct {
		height  int
		bitrate int
	}{
		{2160, 40000}, {1440, 16000}, {1080, 8000}, {720, 4000}, {480, 1500}, {360, 800}, {240, 400},
	}
	for _, tier := range transcodeTiers {
		if tier.height >= sourceHeight || ptHeights[tier.height] {
			continue
		}
		options = append(options, QualityOption{
			Height:      tier.height,
			Label:       fmt.Sprintf("%dp", tier.height),
			Instant:     false,
			Source:      "transcode",
			BitrateKbps: tier.bitrate,
		})
	}

	// Sort: Original first, then by height descending
	if len(options) > 1 {
		sort.Slice(options[1:], func(i, j int) bool {
			return options[1+i].Height > options[1+j].Height
		})
	}

	return options
}

func resolveSelectedAudioTrackID(requestedID int, audioTracks []model.AudioTrack) int {
	if requestedID <= 0 {
		return 0
	}

	for _, track := range audioTracks {
		if int(track.ID) != requestedID {
			continue
		}
		if track.IsDefault {
			return 0
		}
		return requestedID
	}

	return 0
}

// GetClientCapabilities returns detected client capabilities
func (h *PlaybackHandler) GetClientCapabilities(w http.ResponseWriter, r *http.Request) {
	info := playback.GetClientInfo(r.UserAgent())

	resp := map[string]interface{}{
		"browser":   info.Browser,
		"platform":  info.Platform,
		"is_mobile": info.IsMobile,
	}

	if info.Profile != nil {
		resp["video_codecs"] = info.Profile.SupportedVideoCodecs
		resp["audio_codecs"] = info.Profile.SupportedAudioCodecs
		resp["containers"] = info.Profile.SupportedContainers
		resp["max_resolution"] = info.Profile.MaxHeight
		resp["max_bitrate"] = info.Profile.MaxBitrate
		resp["supports_hls"] = info.Profile.SupportsHLS
		resp["supports_webm"] = info.Profile.SupportsWebM
	}

	respondJSON(w, http.StatusOK, resp)
}

func applyClientCapabilityOverrides(profile *playback.DeviceProfile, clientCaps PlaybackInfoRequest) *playback.DeviceProfile {
	if profile == nil {
		return nil
	}

	clone := *profile
	if len(clientCaps.VideoCodecs) > 0 {
		clone.SupportedVideoCodecs = normalizeCapabilityValues(clientCaps.VideoCodecs)
	}
	if len(clientCaps.AudioCodecs) > 0 {
		clone.SupportedAudioCodecs = normalizeCapabilityValues(clientCaps.AudioCodecs)
	}
	if len(clientCaps.Containers) > 0 {
		clone.SupportedContainers = normalizeCapabilityValues(clientCaps.Containers)
	}
	if clientCaps.MaxHeight > 0 {
		clone.MaxHeight = clientCaps.MaxHeight
	}
	if clientCaps.PreferDirectPlay {
		clone.MaxBitrate = 0
		clone.MaxHeight = 0
		// Only the Velox Desktop (libmpv) client sets prefer_direct_play.
		// libmpv links jellyfin-ffmpeg with the tonemapx SIMD filter, so HDR10,
		// HLG, and DV P5 all reshape on the client. Flip both flags defensively
		// here in case a proxy strips the custom UA — DV P7/8 already passes
		// through the remove_dovi BSF and direct-plays without reshape.
		clone.SupportsHDR = true
		clone.SupportsDolbyVision = true
	}

	return &clone
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, list := range values {
		copied := make([]string, len(list))
		copy(copied, list)
		cloned[key] = copied
	}
	return cloned
}

func buildURLWithQuery(path string, values url.Values) string {
	if len(values) == 0 {
		return path
	}
	return path + "?" + values.Encode()
}
