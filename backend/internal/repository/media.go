package repository

import (
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

type MediaRepo struct {
	db *sql.DB
}

func NewMediaRepo(db *sql.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

func (r *MediaRepo) List(libraryID int64, search string, limit, offset int) ([]model.Media, error) {
	query := `SELECT id, library_id, title, file_path, duration, size,
		width, height, video_codec, audio_codec, container, bitrate,
		has_subtitle, poster_path, created_at, updated_at
		FROM media WHERE 1=1`
	args := []any{}

	if libraryID > 0 {
		query += " AND library_id = ?"
		args = append(args, libraryID)
	}
	if search != "" {
		query += " AND title LIKE ?"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY title"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(
			&m.ID, &m.LibraryID, &m.Title, &m.FilePath, &m.Duration, &m.Size,
			&m.Width, &m.Height, &m.VideoCodec, &m.AudioCodec, &m.Container, &m.Bitrate,
			&m.HasSubtitle, &m.PosterPath, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *MediaRepo) GetByID(id int64) (*model.Media, error) {
	var m model.Media
	err := r.db.QueryRow(`SELECT id, library_id, title, file_path, duration, size,
		width, height, video_codec, audio_codec, container, bitrate,
		has_subtitle, poster_path, created_at, updated_at
		FROM media WHERE id = ?`, id).
		Scan(&m.ID, &m.LibraryID, &m.Title, &m.FilePath, &m.Duration, &m.Size,
			&m.Width, &m.Height, &m.VideoCodec, &m.AudioCodec, &m.Container, &m.Bitrate,
			&m.HasSubtitle, &m.PosterPath, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MediaRepo) Upsert(m *model.Media) error {
	_, err := r.db.Exec(`INSERT INTO media
		(library_id, title, file_path, duration, size, width, height,
		 video_codec, audio_codec, container, bitrate, has_subtitle, poster_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			duration=excluded.duration, size=excluded.size,
			width=excluded.width, height=excluded.height,
			video_codec=excluded.video_codec, audio_codec=excluded.audio_codec,
			container=excluded.container, bitrate=excluded.bitrate,
			has_subtitle=excluded.has_subtitle, updated_at=CURRENT_TIMESTAMP`,
		m.LibraryID, m.Title, m.FilePath, m.Duration, m.Size,
		m.Width, m.Height, m.VideoCodec, m.AudioCodec, m.Container,
		m.Bitrate, m.HasSubtitle, m.PosterPath)
	return err
}

func (r *MediaRepo) ExistsByPath(path string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM media WHERE file_path = ?", path).Scan(&count)
	return count > 0, err
}

func (r *MediaRepo) DeleteByLibraryID(libraryID int64) error {
	_, err := r.db.Exec("DELETE FROM media WHERE library_id = ?", libraryID)
	return err
}
