package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// SubtitleRepo handles subtitles database operations
type SubtitleRepo struct {
	db DBTX
}

func NewSubtitleRepo(db DBTX) *SubtitleRepo {
	return &SubtitleRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *SubtitleRepo) WithTx(tx *sql.Tx) *SubtitleRepo {
	return &SubtitleRepo{db: tx}
}

// Create inserts a new subtitle
func (r *SubtitleRepo) Create(ctx context.Context, s *model.Subtitle) error {
	query := `INSERT INTO subtitles
		(media_file_id, language, codec, title, is_embedded, stream_index,
		 file_path, is_forced, is_default, is_sdh)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id`

	isEmbedded := 0
	if s.IsEmbedded {
		isEmbedded = 1
	}
	isForced := 0
	if s.IsForced {
		isForced = 1
	}
	isDefault := 0
	if s.IsDefault {
		isDefault = 1
	}
	isSDH := 0
	if s.IsSDH {
		isSDH = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		s.MediaFileID, s.Language, s.Codec, s.Title, isEmbedded, s.StreamIndex,
		s.FilePath, isForced, isDefault, isSDH)
	return row.Scan(&s.ID)
}

// GetByID retrieves a subtitle by ID
func (r *SubtitleRepo) GetByID(ctx context.Context, id int64) (*model.Subtitle, error) {
	var s model.Subtitle
	var isEmbedded, isForced, isDefault, isSDH int

	err := r.db.QueryRowContext(ctx, `SELECT id, media_file_id, language, codec, title,
		is_embedded, stream_index, file_path, is_forced, is_default, is_sdh
		FROM subtitles WHERE id = ?`, id).
		Scan(&s.ID, &s.MediaFileID, &s.Language, &s.Codec, &s.Title,
			&isEmbedded, &s.StreamIndex, &s.FilePath, &isForced, &isDefault, &isSDH)
	if err != nil {
		return nil, err
	}
	s.IsEmbedded = isEmbedded == 1
	s.IsForced = isForced == 1
	s.IsDefault = isDefault == 1
	s.IsSDH = isSDH == 1
	return &s, nil
}

// ListByMediaFileID retrieves all subtitles for a media file
func (r *SubtitleRepo) ListByMediaFileID(ctx context.Context, mediaFileID int64) ([]model.Subtitle, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_file_id, language, codec, title,
		is_embedded, stream_index, file_path, is_forced, is_default, is_sdh
		FROM subtitles WHERE media_file_id = ?
		ORDER BY is_default DESC, language`, mediaFileID)
	if err != nil {
		return nil, fmt.Errorf("listing subtitles: %w", err)
	}
	defer rows.Close()

	var items []model.Subtitle
	for rows.Next() {
		var s model.Subtitle
		var isEmbedded, isForced, isDefault, isSDH int

		if err := rows.Scan(&s.ID, &s.MediaFileID, &s.Language, &s.Codec, &s.Title,
			&isEmbedded, &s.StreamIndex, &s.FilePath, &isForced, &isDefault, &isSDH); err != nil {
			return nil, fmt.Errorf("scanning subtitle: %w", err)
		}
		s.IsEmbedded = isEmbedded == 1
		s.IsForced = isForced == 1
		s.IsDefault = isDefault == 1
		s.IsSDH = isSDH == 1
		items = append(items, s)
	}
	return items, rows.Err()
}

// DeleteByMediaFileID removes all subtitles for a media file (for rescan)
func (r *SubtitleRepo) DeleteByMediaFileID(ctx context.Context, mediaFileID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM subtitles WHERE media_file_id = ?", mediaFileID)
	return err
}

// Delete removes a subtitle
func (r *SubtitleRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM subtitles WHERE id = ?", id)
	return err
}

// Update updates a subtitle
func (r *SubtitleRepo) Update(ctx context.Context, s *model.Subtitle) error {
	isEmbedded := 0
	if s.IsEmbedded {
		isEmbedded = 1
	}
	isForced := 0
	if s.IsForced {
		isForced = 1
	}
	isDefault := 0
	if s.IsDefault {
		isDefault = 1
	}
	isSDH := 0
	if s.IsSDH {
		isSDH = 1
	}

	_, err := r.db.ExecContext(ctx, `UPDATE subtitles SET
		language = ?, codec = ?, title = ?, is_embedded = ?, stream_index = ?,
		file_path = ?, is_forced = ?, is_default = ?, is_sdh = ?
		WHERE id = ?`,
		s.Language, s.Codec, s.Title, isEmbedded, s.StreamIndex,
		s.FilePath, isForced, isDefault, isSDH, s.ID)
	return err
}

// SetDefault sets a subtitle as the default for its media file
func (r *SubtitleRepo) SetDefault(ctx context.Context, mediaFileID, subtitleID int64) error {
	// First clear default for all subtitles of this media file
	_, err := r.db.ExecContext(ctx, "UPDATE subtitles SET is_default = 0 WHERE media_file_id = ?", mediaFileID)
	if err != nil {
		return err
	}
	// Then set the new default
	_, err = r.db.ExecContext(ctx, "UPDATE subtitles SET is_default = 1 WHERE id = ? AND media_file_id = ?", subtitleID, mediaFileID)
	return err
}
