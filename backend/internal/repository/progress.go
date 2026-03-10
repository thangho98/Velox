package repository

import (
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

type ProgressRepo struct {
	db *sql.DB
}

func NewProgressRepo(db *sql.DB) *ProgressRepo {
	return &ProgressRepo{db: db}
}

func (r *ProgressRepo) Get(mediaID int64) (*model.Progress, error) {
	var p model.Progress
	err := r.db.QueryRow(
		"SELECT media_id, position, completed, updated_at FROM progress WHERE media_id = ?",
		mediaID,
	).Scan(&p.MediaID, &p.Position, &p.Completed, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return &model.Progress{MediaID: mediaID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProgressRepo) Upsert(p *model.Progress) error {
	_, err := r.db.Exec(`INSERT INTO progress (media_id, position, completed, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(media_id) DO UPDATE SET
			position=excluded.position, completed=excluded.completed,
			updated_at=CURRENT_TIMESTAMP`,
		p.MediaID, p.Position, p.Completed)
	return err
}
