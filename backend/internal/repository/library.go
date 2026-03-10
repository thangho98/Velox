package repository

import (
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

type LibraryRepo struct {
	db *sql.DB
}

func NewLibraryRepo(db *sql.DB) *LibraryRepo {
	return &LibraryRepo{db: db}
}

func (r *LibraryRepo) List() ([]model.Library, error) {
	rows, err := r.db.Query("SELECT id, name, path, created_at FROM libraries ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libs []model.Library
	for rows.Next() {
		var l model.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Path, &l.CreatedAt); err != nil {
			return nil, err
		}
		libs = append(libs, l)
	}
	return libs, rows.Err()
}

func (r *LibraryRepo) GetByID(id int64) (*model.Library, error) {
	var l model.Library
	err := r.db.QueryRow("SELECT id, name, path, created_at FROM libraries WHERE id = ?", id).
		Scan(&l.ID, &l.Name, &l.Path, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LibraryRepo) Create(name, path string) (*model.Library, error) {
	res, err := r.db.Exec("INSERT INTO libraries (name, path) VALUES (?, ?)", name, path)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *LibraryRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM libraries WHERE id = ?", id)
	return err
}
