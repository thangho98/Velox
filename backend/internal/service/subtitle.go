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
	subtitlepkg "github.com/thawng/velox/pkg/subtitle"
	"github.com/thawng/velox/pkg/translate"
)

// SubtitleService handles subtitle business logic
type SubtitleService struct {
	db            *sql.DB
	subtitleRepo  *repository.SubtitleRepo
	mediaFileRepo *repository.MediaFileRepo
	settingsRepo  *repository.AppSettingsRepo
	subtitleCache string
}

func NewSubtitleService(db *sql.DB, subtitleRepo *repository.SubtitleRepo, mediaFileRepo *repository.MediaFileRepo) *SubtitleService {
	return &SubtitleService{
		db:            db,
		subtitleRepo:  subtitleRepo,
		mediaFileRepo: mediaFileRepo,
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

		cacheDir := filepath.Join(s.subtitleCache, fmt.Sprintf("%d", sub.MediaFileID))
		vttPath, err := subtitlepkg.ExtractSubtitle(mf.FilePath, sub.StreamIndex, cacheDir)
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
	if s.settingsRepo != nil {
		deeplAPIKey, _ = s.settingsRepo.Get(ctx, model.SettingDeepLAPIKey)
	}

	return s.TranslateSubtitle(ctx, subtitleID, targetLang, deeplAPIKey, s.subtitleCache)
}

// TranslateSubtitle translates a subtitle file to the target language.
// Uses DeepL (if API key configured) with Google Translate fallback.
// Returns the newly created subtitle record.
func (s *SubtitleService) TranslateSubtitle(ctx context.Context, subtitleID int64, targetLang, deeplAPIKey, subtitleDir string) (*model.Subtitle, error) {
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
		extractDir := filepath.Join(subtitleDir, strconv.FormatInt(source.MediaFileID, 10))
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return nil, fmt.Errorf("creating extract dir: %w", err)
		}
		extractPath := filepath.Join(extractDir, fmt.Sprintf("extracted_%d.srt", source.StreamIndex))
		// Extract via FFmpeg as SRT (not VTT, since our translator expects SRT)
		cmd := exec.Command("ffmpeg", "-y",
			"-i", mf.FilePath,
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

	// Choose translator: DeepL primary, Google fallback
	var translator translate.Translator
	if deeplAPIKey != "" {
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
		// If DeepL fails (quota exceeded), fallback to Google
		if deeplAPIKey != "" {
			slog.Warn("deepl translation failed, falling back to google", "error", err)
			translator = translate.NewGoogle()
			translated, err = translate.TranslateSRT(ctx, translator, content, targetLang)
			if err != nil {
				return nil, fmt.Errorf("translation failed: %w", err)
			}
		} else {
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
