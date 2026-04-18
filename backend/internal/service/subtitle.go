package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/pkg/ffmpegbin"
	subtitlepkg "github.com/thawng/velox/pkg/subtitle"
	"github.com/thawng/velox/pkg/translate"
)

// SubtitleService handles subtitle business logic
type SubtitleService struct {
	db            *sql.DB
	subtitleRepo  *repository.SubtitleRepo
	mediaFileRepo *repository.MediaFileRepo
	mediaRepo     *repository.MediaRepo
	genreRepo     *repository.GenreRepo
	settingsRepo  *repository.AppSettingsRepo
	subtitleCache string

	cloudResolver func(context.Context, int64, *model.MediaFile) (string, error)
}

func NewSubtitleService(
	db *sql.DB,
	subtitleRepo *repository.SubtitleRepo,
	mediaFileRepo *repository.MediaFileRepo,
	mediaRepo *repository.MediaRepo,
	genreRepo *repository.GenreRepo,
) *SubtitleService {
	return &SubtitleService{
		db:            db,
		subtitleRepo:  subtitleRepo,
		mediaFileRepo: mediaFileRepo,
		mediaRepo:     mediaRepo,
		genreRepo:     genreRepo,
	}
}

// SetSettingsRepo configures app settings access used by subtitle translation.
func (s *SubtitleService) SetSettingsRepo(settingsRepo *repository.AppSettingsRepo) {
	s.settingsRepo = settingsRepo
}

// SetCacheDir configures the subtitle extraction/translation cache directory.
func (s *SubtitleService) SetCacheDir(subtitleCache string) {
	s.subtitleCache = subtitleCache
}

// SetCloudResolver injects a function that translates cloud native paths into HTTP URLs.
func (s *SubtitleService) SetCloudResolver(resolver func(context.Context, int64, *model.MediaFile) (string, error)) {
	s.cloudResolver = resolver
}

// resolveFilePath returns the effective file path for FFmpeg, resolving cloud paths to HTTP URLs.
func (s *SubtitleService) resolveFilePath(ctx context.Context, mf *model.MediaFile) (string, error) {
	if s.cloudResolver == nil || !scanner.IsCloudPath(mf.FilePath) {
		return mf.FilePath, nil
	}
	return s.cloudResolver(ctx, mf.MediaID, mf)
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
		if errors.Is(err, repository.ErrNotFound) {
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
	if err := s.subtitleRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("deleting subtitle %d: %w", id, err)
	}
	return nil
}

// SetDefault sets a subtitle as default (atomic: clear all defaults + set new default in one transaction)
func (s *SubtitleService) SetDefault(ctx context.Context, mediaFileID, subtitleID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Clear all defaults for this media file
	if _, err := tx.ExecContext(ctx, "UPDATE subtitles SET is_default = 0 WHERE media_file_id = ?", mediaFileID); err != nil {
		return fmt.Errorf("clearing default subtitles for media file %d: %w", mediaFileID, err)
	}

	// Set the new default
	res, err := tx.ExecContext(ctx, "UPDATE subtitles SET is_default = 1 WHERE id = ? AND media_file_id = ?", subtitleID, mediaFileID)
	if err != nil {
		return fmt.Errorf("setting default subtitle %d for media file %d: %w", subtitleID, mediaFileID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// ServeContent returns subtitle content as WebVTT-compatible bytes.
func (s *SubtitleService) ServeContent(ctx context.Context, subtitleID int64) ([]byte, error) {
	sub, err := s.Get(ctx, subtitleID)
	if err != nil {
		return nil, err
	}

	if sub.IsEmbedded {
		if s.mediaFileRepo == nil {
			return nil, fmt.Errorf("media file repository not configured")
		}
		if s.subtitleCache == "" {
			return nil, fmt.Errorf("subtitle cache not configured")
		}

		mf, err := s.mediaFileRepo.GetByID(ctx, sub.MediaFileID)
		if err != nil {
			return nil, fmt.Errorf("looking up media file: %w", err)
		}

		inputPath, err := s.resolveFilePath(ctx, mf)
		if err != nil {
			return nil, fmt.Errorf("resolving cloud path for subtitle extraction: %w", err)
		}

		cacheDir := filepath.Join(s.subtitleCache, fmt.Sprintf("%d", sub.MediaFileID))
		vttPath, err := subtitlepkg.ExtractSubtitle(inputPath, sub.StreamIndex, cacheDir)
		if err != nil {
			return nil, fmt.Errorf("extracting subtitle: %w", err)
		}

		data, err := os.ReadFile(vttPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("reading extracted subtitle: %w", err)
		}

		return data, nil
	}

	if sub.FilePath == "" {
		return nil, ErrNotFound
	}

	data, err := os.ReadFile(sub.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("reading subtitle file: %w", err)
	}

	codec := strings.ToLower(sub.Codec)
	if codec == "subrip" || codec == "srt" {
		return subtitlepkg.SRTToVTT(data), nil
	}

	return data, nil
}

// Translate translates a subtitle using configured provider settings.
func (s *SubtitleService) Translate(ctx context.Context, subtitleID int64, targetLang string) (*model.Subtitle, error) {
	deeplAPIKey := ""
	aiConfig := translate.AIConfig{}
	if s.settingsRepo != nil {
		vals, _ := s.settingsRepo.GetMulti(
			ctx,
			model.SettingDeepLAPIKey,
			model.SettingAITranslationProvider,
			model.SettingAITranslationAPIKey,
			model.SettingAITranslationBaseURL,
			model.SettingAITranslationModel,
		)
		deeplAPIKey = vals[model.SettingDeepLAPIKey]
		aiConfig = translate.AIConfig{
			Provider: translate.AIProvider(vals[model.SettingAITranslationProvider]),
			APIKey:   vals[model.SettingAITranslationAPIKey],
			BaseURL:  vals[model.SettingAITranslationBaseURL],
			Model:    vals[model.SettingAITranslationModel],
		}
	}

	return s.TranslateSubtitle(ctx, subtitleID, targetLang, aiConfig, deeplAPIKey, s.subtitleCache)
}

// TranslateSubtitle translates a subtitle file to the target language.
// Uses the configured AI provider first, then DeepL, then Google Translate fallback.
// Returns the newly created subtitle record.
func (s *SubtitleService) TranslateSubtitle(
	ctx context.Context,
	subtitleID int64,
	targetLang string,
	aiConfig translate.AIConfig,
	deeplAPIKey,
	subtitleDir string,
) (*model.Subtitle, error) {
	// Get source subtitle
	source, err := s.Get(ctx, subtitleID)
	if err != nil {
		return nil, fmt.Errorf("getting source subtitle: %w", err)
	}

	if source.Language == targetLang {
		return nil, fmt.Errorf("source and target language are the same: %s", targetLang)
	}

	// Read source file content — extract embedded subs via FFmpeg if needed
	var content string
	if source.FilePath != "" && !source.IsEmbedded {
		data, err := os.ReadFile(source.FilePath)
		if err != nil {
			return nil, fmt.Errorf("reading subtitle file: %w", err)
		}
		content = string(data)
	} else if source.IsEmbedded {
		// Extract embedded subtitle to a temp SRT file
		mf, err := s.mediaFileRepo.GetByID(ctx, source.MediaFileID)
		if err != nil {
			return nil, fmt.Errorf("getting media file for extraction: %w", err)
		}
		inputPath, err := s.resolveFilePath(ctx, mf)
		if err != nil {
			return nil, fmt.Errorf("resolving cloud path for subtitle translation: %w", err)
		}
		extractDir := filepath.Join(subtitleDir, strconv.FormatInt(source.MediaFileID, 10))
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return nil, fmt.Errorf("creating extract dir: %w", err)
		}
		extractPath := filepath.Join(extractDir, fmt.Sprintf("extracted_%d.srt", source.StreamIndex))
		cmd := exec.Command(ffmpegbin.FFmpeg(), "-y",
			"-i", inputPath,
			"-map", fmt.Sprintf("0:%d", source.StreamIndex),
			"-c:s", "srt",
			extractPath,
		)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("extracting embedded subtitle: %w", err)
		}
		data, err := os.ReadFile(extractPath)
		if err != nil {
			return nil, fmt.Errorf("reading extracted subtitle: %w", err)
		}
		content = string(data)
	} else {
		return nil, fmt.Errorf("subtitle has no file path")
	}

	aiConfig.Context = s.buildAITranslationContext(ctx, source.MediaFileID)

	// Choose translator: AI provider first, then DeepL, then Google.
	var translator translate.Translator
	if aiConfig.Provider != "" {
		translator, err = translate.NewAI(aiConfig)
		if err != nil {
			return nil, fmt.Errorf("configuring ai translator: %w", err)
		}
	} else if deeplAPIKey != "" {
		translator = translate.NewDeepL(deeplAPIKey)
	} else {
		translator = translate.NewGoogle()
	}

	slog.Info("translating subtitle",
		"subtitle_id", subtitleID,
		"from", source.Language,
		"to", targetLang,
		"translator", translator.Name(),
		"cues", len(translate.ParseSRT(content)),
	)

	// Translate
	translated, err := translate.TranslateSRT(ctx, translator, content, targetLang)
	if err != nil {
		switch {
		case aiConfig.Provider != "":
			slog.Warn("ai subtitle translation failed", "provider", aiConfig.Provider, "error", err)
			if deeplAPIKey != "" {
				translator = translate.NewDeepL(deeplAPIKey)
				slog.Warn("falling back to deepl after ai failure", "error", err)
				translated, err = translate.TranslateSRT(ctx, translator, content, targetLang)
				if err == nil {
					break
				}
				slog.Warn("deepl fallback failed after ai failure", "error", err)
			}
			translator = translate.NewGoogle()
			slog.Warn("falling back to google after ai failure", "error", err)
			translated, err = translate.TranslateSRT(ctx, translator, content, targetLang)
			if err != nil {
				return nil, fmt.Errorf("translation failed after ai fallback: %w", err)
			}
		case deeplAPIKey != "":
			slog.Warn("deepl translation failed, falling back to google", "error", err)
			translator = translate.NewGoogle()
			translated, err = translate.TranslateSRT(ctx, translator, content, targetLang)
			if err != nil {
				return nil, fmt.Errorf("translation failed: %w", err)
			}
		default:
			return nil, fmt.Errorf("translation failed: %w", err)
		}
	}

	// Save translated file
	dir := filepath.Join(subtitleDir, strconv.FormatInt(source.MediaFileID, 10))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating subtitle dir: %w", err)
	}

	langNames := map[string]string{
		"en": "English", "vi": "Vietnamese", "fr": "French", "de": "German",
		"es": "Spanish", "pt": "Portuguese", "it": "Italian", "nl": "Dutch",
		"ja": "Japanese", "ko": "Korean", "zh": "Chinese", "ar": "Arabic",
		"ru": "Russian", "th": "Thai", "pl": "Polish", "tr": "Turkish",
	}
	langName := langNames[targetLang]
	if langName == "" {
		langName = strings.ToUpper(targetLang)
	}

	savePath := filepath.Join(dir, fmt.Sprintf("translated_%s_%d.srt", targetLang, subtitleID))

	// Delete existing translated subtitle with same file path if it exists
	if existing, err := s.subtitleRepo.FindByMediaFileIDAndFilePath(ctx, source.MediaFileID, savePath); err == nil {
		slog.Info("deleting existing translated subtitle before re-translating",
			"existing_subtitle_id", existing.ID,
			"media_file_id", source.MediaFileID,
			"file_path", savePath,
		)
		if err := s.subtitleRepo.Delete(ctx, existing.ID); err != nil {
			slog.Warn("failed to delete existing subtitle, will overwrite", "error", err)
		}
		// Also remove the old file
		if _, err := os.Stat(savePath); err == nil {
			os.Remove(savePath)
		}
	}

	if err := os.WriteFile(savePath, []byte(translated), 0644); err != nil {
		return nil, fmt.Errorf("saving translated subtitle: %w", err)
	}

	// Create DB record
	title := fmt.Sprintf("%s (%s auto)", langName, translator.Name())
	sub := &model.Subtitle{
		MediaFileID: source.MediaFileID,
		Language:    targetLang,
		Codec:       "srt",
		Title:       title,
		IsEmbedded:  false,
		StreamIndex: -1,
		FilePath:    savePath,
	}
	if err := s.subtitleRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("saving subtitle record: %w", err)
	}

	slog.Info("subtitle translated",
		"subtitle_id", sub.ID,
		"from", source.Language,
		"to", targetLang,
		"translator", translator.Name(),
		"file", savePath,
	)

	return sub, nil
}

func (s *SubtitleService) buildAITranslationContext(ctx context.Context, mediaFileID int64) *translate.AIContext {
	if s.mediaFileRepo == nil || s.mediaRepo == nil {
		return nil
	}

	mediaFile, err := s.mediaFileRepo.GetByID(ctx, mediaFileID)
	if err != nil {
		return nil
	}

	media, err := s.mediaRepo.GetByID(ctx, mediaFile.MediaID)
	if err != nil {
		return nil
	}

	context := &translate.AIContext{
		Title:     media.Title,
		MediaType: media.MediaType,
		Overview:  media.Overview,
		Tagline:   media.Tagline,
	}

	if s.genreRepo != nil {
		genres, err := s.genreRepo.ListByMediaID(ctx, media.ID)
		if err == nil {
			context.Genres = make([]string, 0, len(genres))
			for _, genre := range genres {
				if genre.Name != "" {
					context.Genres = append(context.Genres, genre.Name)
				}
			}
		}
	}

	if context.Title == "" && context.MediaType == "" && context.Overview == "" && context.Tagline == "" && len(context.Genres) == 0 {
		return nil
	}

	return context
}
