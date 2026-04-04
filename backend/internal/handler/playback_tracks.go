package handler

import (
	"fmt"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
)

func normalizeCapabilityValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeLanguageCode(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	switch language {
	case "en", "eng":
		return "eng"
	case "vi", "vie":
		return "vie"
	case "zh", "zho", "chi":
		return "zho"
	case "ja", "jpn":
		return "jpn"
	case "ko", "kor":
		return "kor"
	case "fr", "fra", "fre":
		return "fra"
	case "de", "deu", "ger", "dut":
		return "deu"
	case "es", "spa":
		return "spa"
	case "pt", "por":
		return "por"
	default:
		return language
	}
}

func normalizeContainerValue(container string) string {
	container = strings.TrimSpace(strings.ToLower(container))
	switch container {
	case "mp4", "mpeg4", "m4v":
		return playback.ContainerMP4
	case "webm":
		return playback.ContainerWebM
	case "mkv", "matroska", "matroska,webm":
		return playback.ContainerMKV
	case "mov", "qt":
		return playback.ContainerMOV
	default:
		return container
	}
}

func languageMatches(lhs, rhs string) bool {
	if lhs == "" || rhs == "" {
		return false
	}
	return normalizeLanguageCode(lhs) == normalizeLanguageCode(rhs)
}

func findSubtitleByLanguage(subtitles []model.Subtitle, language string) *model.Subtitle {
	if language == "" || language == "off" {
		return nil
	}

	var imageMatch *model.Subtitle
	for i := range subtitles {
		if !languageMatches(subtitles[i].Language, language) {
			continue
		}

		normalized := playback.NormalizeSubtitleCodec(subtitles[i].Codec)
		if normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub {
			if imageMatch == nil {
				imageMatch = &subtitles[i]
			}
			continue
		}

		return &subtitles[i]
	}

	if imageMatch != nil {
		return imageMatch
	}

	for i := range subtitles {
		if languageMatches(subtitles[i].Language, language) {
			return &subtitles[i]
		}
	}

	return nil
}

func findSubtitleByID(subtitles []model.Subtitle, subtitleID int) *model.Subtitle {
	if subtitleID <= 0 {
		return nil
	}
	for i := range subtitles {
		if int(subtitles[i].ID) == subtitleID {
			return &subtitles[i]
		}
	}
	return nil
}

func filterPlayableSubtitles(subtitles []model.Subtitle, burnInSupported bool) []model.Subtitle {
	if burnInSupported {
		return subtitles
	}

	filtered := make([]model.Subtitle, 0, len(subtitles))
	for _, sub := range subtitles {
		normalized := playback.NormalizeSubtitleCodec(sub.Codec)
		if normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub {
			continue
		}
		filtered = append(filtered, sub)
	}
	return filtered
}

func findAudioTrackByID(audioTracks []model.AudioTrack, trackID int) *model.AudioTrack {
	if trackID <= 0 {
		return nil
	}
	for i := range audioTracks {
		if int(audioTracks[i].ID) == trackID {
			return &audioTracks[i]
		}
	}
	return nil
}

func defaultAudioTrack(audioTracks []model.AudioTrack) *model.AudioTrack {
	if len(audioTracks) == 0 {
		return nil
	}
	for i := range audioTracks {
		if audioTracks[i].IsDefault {
			return &audioTracks[i]
		}
	}
	return &audioTracks[0]
}

func audioTrackPlayable(track model.AudioTrack, profile *playback.DeviceProfile) bool {
	if profile == nil {
		return true
	}
	return profile.SupportsAudioCodec(playback.NormalizeCodec(track.Codec))
}

func findCompatibleAudioTrack(
	audioTracks []model.AudioTrack,
	profile *playback.DeviceProfile,
	preferredLanguage string,
	fallbackLanguage string,
) *model.AudioTrack {
	for _, language := range []string{preferredLanguage, fallbackLanguage} {
		if language == "" {
			continue
		}
		for i := range audioTracks {
			if !languageMatches(audioTracks[i].Language, language) {
				continue
			}
			if !audioTrackPlayable(audioTracks[i], profile) {
				continue
			}
			return &audioTracks[i]
		}
	}

	for i := range audioTracks {
		if audioTrackPlayable(audioTracks[i], profile) {
			return &audioTracks[i]
		}
	}

	return nil
}

func resolvePlaybackAudioTrack(
	requestedID int,
	preferredLanguage string,
	audioTracks []model.AudioTrack,
	profile *playback.DeviceProfile,
) (*model.AudioTrack, int, bool) {
	requested := findAudioTrackByID(audioTracks, requestedID)
	if requested != nil {
		if requested.IsDefault {
			return requested, 0, false
		}
		return requested, int(requested.ID), false
	}

	base := defaultAudioTrack(audioTracks)
	if base == nil {
		return nil, 0, false
	}
	if audioTrackPlayable(*base, profile) {
		return base, 0, false
	}

	compatible := findCompatibleAudioTrack(audioTracks, profile, preferredLanguage, base.Language)
	if compatible == nil {
		return base, 0, false
	}

	effectiveID := 0
	if !compatible.IsDefault {
		effectiveID = int(compatible.ID)
	}
	return compatible, effectiveID, compatible.ID != base.ID
}

func adjustPlaybackDecisionForSelectedAudioTrack(
	decision playback.PlaybackDecision,
	selectedTrack *model.AudioTrack,
	effectiveAudioTrackID int,
	autoSelected bool,
) playback.PlaybackDecision {
	if selectedTrack == nil || effectiveAudioTrackID <= 0 {
		return decision
	}

	label := selectedTrack.Title
	if label == "" {
		label = selectedTrack.Language
	}
	if label == "" {
		label = fmt.Sprintf("track %d", selectedTrack.ID)
	}

	if decision.Method == playback.MethodDirectPlay {
		decision.Method = playback.MethodDirectStream
		decision.Container = playback.ContainerMP4
	}

	if decision.Method == playback.MethodDirectStream {
		if autoSelected {
			decision.Reason += fmt.Sprintf(" + auto-selected compatible audio track %s", label)
		} else {
			decision.Reason += fmt.Sprintf(" + selected audio track %s requires remuxed direct stream", label)
		}
	}

	return decision
}

func playbackModeQuery(method playback.PlaybackMethod) string {
	switch method {
	case playback.MethodDirectStream:
		return "directstream"
	default:
		return "direct"
	}
}
