package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
)

// SubtitleService handles subtitle business logic
type SubtitleService struct {
	subtitleRepo  *repository.SubtitleRepo
	mediaFileRepo *repository.MediaFileRepo
}

func NewSubtitleService(subtitleRepo *repository.SubtitleRepo, mediaFileRepo *repository.MediaFileRepo) *SubtitleService {
	return &SubtitleService{
		subtitleRepo:  subtitleRepo,
		mediaFileRepo: mediaFileRepo,
	}
}

// ListByMediaFile returns all subtitles for a media file
func (s *SubtitleService) ListByMediaFile(ctx context.Context, mediaFileID int64) ([]model.Subtitle, error) {
	subtitles, err := s.subtitleRepo.ListByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return nil, err
	}
	subtitles = filterMalformedExternalSubtitles(subtitles)
	if s.mediaFileRepo == nil || len(subtitles) < 2 {
		return subtitles, nil
	}

	mediaFile, err := s.mediaFileRepo.GetByID(ctx, mediaFileID)
	if err != nil {
		return subtitles, nil
	}
	return rankSubtitlesForMediaFile(subtitles, mediaFile), nil
}

// Get returns a subtitle by ID
func (s *SubtitleService) Get(ctx context.Context, id int64) (*model.Subtitle, error) {
	sub, err := s.subtitleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sub, nil
}

// Create creates a new subtitle
func (s *SubtitleService) Create(ctx context.Context, subtitle *model.Subtitle) error {
	return s.subtitleRepo.Create(ctx, subtitle)
}

// Update updates a subtitle
func (s *SubtitleService) Update(ctx context.Context, subtitle *model.Subtitle) error {
	return s.subtitleRepo.Update(ctx, subtitle)
}

// Delete deletes a subtitle
func (s *SubtitleService) Delete(ctx context.Context, id int64) error {
	return s.subtitleRepo.Delete(ctx, id)
}

// SetDefault sets a subtitle as default
func (s *SubtitleService) SetDefault(ctx context.Context, mediaFileID, subtitleID int64) error {
	return s.subtitleRepo.SetDefault(ctx, mediaFileID, subtitleID)
}

// AudioTrackService handles audio track business logic
type AudioTrackService struct {
	audioTrackRepo *repository.AudioTrackRepo
}

func NewAudioTrackService(audioTrackRepo *repository.AudioTrackRepo) *AudioTrackService {
	return &AudioTrackService{audioTrackRepo: audioTrackRepo}
}

// ListByMediaFile returns all audio tracks for a media file
func (s *AudioTrackService) ListByMediaFile(ctx context.Context, mediaFileID int64) ([]model.AudioTrack, error) {
	return s.audioTrackRepo.ListByMediaFileID(ctx, mediaFileID)
}

// Get returns an audio track by ID
func (s *AudioTrackService) Get(ctx context.Context, id int64) (*model.AudioTrack, error) {
	track, err := s.audioTrackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return track, nil
}

// Create creates a new audio track
func (s *AudioTrackService) Create(ctx context.Context, track *model.AudioTrack) error {
	return s.audioTrackRepo.Create(ctx, track)
}

// Update updates an audio track
func (s *AudioTrackService) Update(ctx context.Context, track *model.AudioTrack) error {
	return s.audioTrackRepo.Update(ctx, track)
}

// Delete deletes an audio track
func (s *AudioTrackService) Delete(ctx context.Context, id int64) error {
	return s.audioTrackRepo.Delete(ctx, id)
}

// SetDefault sets an audio track as default
func (s *AudioTrackService) SetDefault(ctx context.Context, mediaFileID, trackID int64) error {
	return s.audioTrackRepo.SetDefault(ctx, mediaFileID, trackID)
}

var subtitleCuePattern = regexp.MustCompile(`(?m)(\d{2}:\d{2}:\d{2}[,.]\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}[,.]\d{3})`)
var subtitleTokenSplitter = regexp.MustCompile(`[^a-z0-9]+`)

type subtitleTimingStats struct {
	cueCount   int
	firstStart float64
	lastEnd    float64
}

func rankSubtitlesForMediaFile(subtitles []model.Subtitle, mediaFile *model.MediaFile) []model.Subtitle {
	ranked := append([]model.Subtitle(nil), subtitles...)
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]

		leftLang := normalizeSubtitleLanguage(left.Language)
		rightLang := normalizeSubtitleLanguage(right.Language)

		if left.IsDefault != right.IsDefault {
			return left.IsDefault
		}
		if leftLang != rightLang {
			return leftLang < rightLang
		}

		leftScore := subtitleHeuristicScore(left, mediaFile)
		rightScore := subtitleHeuristicScore(right, mediaFile)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return left.ID < right.ID
	})
	return ranked
}

func subtitleHeuristicScore(sub model.Subtitle, mediaFile *model.MediaFile) int {
	score := 0

	if !sub.IsForced {
		score += 40
	}
	if !sub.IsSDH {
		score += 120
	}

	normalizedCodec := playback.NormalizeSubtitleCodec(sub.Codec)
	switch normalizedCodec {
	case playback.SubtitleSRT, playback.SubtitleVTT, playback.SubtitleASS:
		score += 600
	case playback.SubtitlePGS, playback.SubtitleVobSub:
		score += 100
	default:
		score += 300
	}

	if !sub.IsEmbedded {
		score += 120
	}

	score += releaseOverlapScore(mediaFile.FilePath, sub.FilePath, sub.Title)

	if !sub.IsEmbedded && sub.FilePath != "" {
		stats, ok := analyzeSubtitleTiming(sub.FilePath)
		if ok {
			score += subtitleTimingScore(stats, mediaFile.Duration)
		}
	}

	return score
}

func subtitleTimingScore(stats subtitleTimingStats, mediaDuration float64) int {
	score := 0

	if stats.cueCount >= 80 {
		score += 120
	} else if stats.cueCount >= 20 {
		score += 60
	}

	if stats.firstStart >= 0 {
		switch {
		case stats.firstStart <= 15:
			score += 200
		case stats.firstStart <= 60:
			score += 60
		default:
			score -= 80
		}
	}

	if mediaDuration > 0 && stats.lastEnd > 0 {
		diff := math.Abs(stats.lastEnd - mediaDuration)
		switch {
		case diff <= 3:
			score += 2600
		case diff <= 8:
			score += 1900
		case diff <= 15:
			score += 900
		case diff <= 30:
			score += 200
		default:
			score -= 900
		}
	}

	return score
}

func analyzeSubtitleTiming(path string) (subtitleTimingStats, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return subtitleTimingStats{}, false
	}

	matches := subtitleCuePattern.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		return subtitleTimingStats{}, false
	}

	firstStart, ok := parseSubtitleTimestamp(matches[0][1])
	if !ok {
		return subtitleTimingStats{}, false
	}
	lastEnd, ok := parseSubtitleTimestamp(matches[len(matches)-1][2])
	if !ok {
		return subtitleTimingStats{}, false
	}

	return subtitleTimingStats{
		cueCount:   len(matches),
		firstStart: firstStart,
		lastEnd:    lastEnd,
	}, true
}

func parseSubtitleTimestamp(raw string) (float64, bool) {
	normalized := strings.ReplaceAll(strings.TrimSpace(raw), ",", ".")
	parts := strings.Split(normalized, ":")
	if len(parts) != 3 {
		return 0, false
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, false
	}

	return float64(hours*3600+minutes*60) + seconds, true
}

func releaseOverlapScore(mediaPath, subtitlePath, subtitleTitle string) int {
	mediaTokens := releaseTokens(mediaPath)
	if len(mediaTokens) == 0 {
		return 0
	}

	subtitleTokens := releaseTokens(subtitlePath + " " + subtitleTitle)
	if len(subtitleTokens) == 0 {
		return 0
	}

	overlap := 0
	for token := range subtitleTokens {
		if _, ok := mediaTokens[token]; ok {
			overlap++
		}
	}

	if overlap == 0 {
		return 0
	}
	if overlap >= 6 {
		return 500
	}
	return overlap * 70
}

func releaseTokens(value string) map[string]struct{} {
	base := strings.ToLower(filepath.Base(value))
	ext := strings.TrimSuffix(base, filepath.Ext(base))
	parts := subtitleTokenSplitter.Split(ext, -1)
	tokens := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if len(part) < 3 {
			continue
		}
		switch part {
		case "subdl", "subtitle", "english", "vietnamese", "eng", "vie", "sdh":
			continue
		}
		tokens[part] = struct{}{}
	}
	return tokens
}

func normalizeSubtitleLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "en", "eng":
		return "eng"
	case "vi", "vie":
		return "vie"
	case "zh", "zho", "chi":
		return "zho"
	default:
		return strings.ToLower(strings.TrimSpace(language))
	}
}

func filterMalformedExternalSubtitles(subtitles []model.Subtitle) []model.Subtitle {
	filtered := make([]model.Subtitle, 0, len(subtitles))
	for _, subtitle := range subtitles {
		if isMalformedExternalTextSubtitle(subtitle) {
			continue
		}
		filtered = append(filtered, subtitle)
	}
	return filtered
}

func isMalformedExternalTextSubtitle(sub model.Subtitle) bool {
	if sub.IsEmbedded || sub.FilePath == "" {
		return false
	}

	switch playback.NormalizeSubtitleCodec(sub.Codec) {
	case playback.SubtitleSRT, playback.SubtitleVTT:
	default:
		return false
	}

	data, err := os.ReadFile(sub.FilePath)
	if err != nil {
		return true
	}
	if looksLikeHTMLDocument(data) {
		return true
	}

	_, ok := analyzeSubtitleTiming(sub.FilePath)
	return !ok
}

func looksLikeHTMLDocument(data []byte) bool {
	snippet := strings.ToLower(strings.TrimSpace(string(data)))
	if len(snippet) > 2048 {
		snippet = snippet[:2048]
	}
	return strings.HasPrefix(snippet, "<!doctype html") ||
		strings.HasPrefix(snippet, "<html") ||
		(strings.Contains(snippet, "<head") && strings.Contains(snippet, "<body"))
}
