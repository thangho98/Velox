package repository

import (
	"context"

	"github.com/thawng/velox/internal/model"
)

type LibraryRepo struct {
	db DBTX
}

func NewLibraryRepo(db DBTX) *LibraryRepo {
	return &LibraryRepo{db: db}
}

func (r *LibraryRepo) List(ctx context.Context) ([]model.Library, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, path, type, created_at FROM libraries ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libs []model.Library
	for rows.Next() {
		var l model.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Path, &l.Type, &l.CreatedAt); err != nil {
			return nil, err
		}
		libs = append(libs, l)
	}
	return libs, rows.Err()
}

func (r *LibraryRepo) GetByID(ctx context.Context, id int64) (*model.Library, error) {
	var l model.Library
	err := r.db.QueryRowContext(ctx, "SELECT id, name, path, type, created_at FROM libraries WHERE id = ?", id).
		Scan(&l.ID, &l.Name, &l.Path, &l.Type, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LibraryRepo) Create(ctx context.Context, name, path, libType string) (*model.Library, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO libraries (name, path, type) VALUES (?, ?, ?)", name, path, libType)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(ctx, id)
}

func (r *LibraryRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM libraries WHERE id = ?", id)
	return err
}
