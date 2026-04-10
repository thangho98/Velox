package repository

import (
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

type AppVersionRepo struct {
	db *sql.DB
}

func NewAppVersionRepo(db *sql.DB) *AppVersionRepo {
	return &AppVersionRepo{db: db}
}

func (r *AppVersionRepo) GetLatest(platform string) (*model.AppVersion, error) {
	row := r.db.QueryRow(`
		SELECT id, platform, version_name, version_code, is_mandatory, release_notes, created_at
		FROM app_versions
		WHERE platform = ?
		ORDER BY version_code DESC, created_at DESC
		LIMIT 1
	`, platform)

	var v model.AppVersion
	var isMandatoryInt int
	err := row.Scan(
		&v.ID,
		&v.Platform,
		&v.VersionName,
		&v.VersionCode,
		&isMandatoryInt,
		&v.ReleaseNotes,
		&v.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no version found
		}
		return nil, err
	}

	v.IsMandatory = isMandatoryInt == 1
	return &v, nil
}

// Add a helper for admins to post new versions (if needed)
func (r *AppVersionRepo) Create(v *model.AppVersion) error {
	isMandatoryInt := 0
	if v.IsMandatory {
		isMandatoryInt = 1
	}

	result, err := r.db.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES (?, ?, ?, ?, ?)
	`, v.Platform, v.VersionName, v.VersionCode, isMandatoryInt, v.ReleaseNotes)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		v.ID = int(id)
	}

	return nil
}
