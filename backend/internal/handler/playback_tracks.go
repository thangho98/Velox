package handler

import (
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

func playbackModeQuery(method playback.PlaybackMethod) string {
	switch method {
	case playback.MethodDirectStream:
		return "directstream"
	default:
		return "direct"
	}
}
