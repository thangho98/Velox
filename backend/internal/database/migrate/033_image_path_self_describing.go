package migrate

import (
	"database/sql"
	"strconv"
)

// 033: Backfill local:// image paths to be self-describing.
//
// Previously storage wrote `local://{id}/{filename}`. The entity type ("media",
// "series", "episode") was inferred at read time by whichever helper loaded the
// row. Going forward storage writes `local://{entityType}/{id}/{filename}` so
// paths are self-describing and can be normalized without column context.
//
// This migration prepends the entity type to any existing legacy rows. The
// condition `local://_%/%` plus the NOT-LIKE guards ensures we only touch rows
// that still use the legacy 2-segment form (e.g. `local://42/poster.jpg`).
func up033(tx *sql.Tx) error {
	statements := []struct {
		table      string
		column     string
		entityType string
	}{
		{"media", "poster_path", "media"},
		{"media", "backdrop_path", "media"},
		{"media", "logo_path", "media"},
		{"media", "thumb_path", "media"},
		{"series", "poster_path", "series"},
		{"series", "backdrop_path", "series"},
		{"series", "logo_path", "series"},
		{"series", "thumb_path", "series"},
		{"seasons", "poster_path", "series"},
		{"episodes", "still_path", "episode"},
		{"users", "avatar_path", "user"},
	}

	for _, s := range statements {
		stmt := `UPDATE ` + s.table + ` SET ` + s.column + ` = 'local://` + s.entityType + `/' || substr(` + s.column + `, 9) WHERE ` + s.column + ` LIKE 'local://%' AND ` + s.column + ` NOT LIKE 'local://media/%' AND ` + s.column + ` NOT LIKE 'local://series/%' AND ` + s.column + ` NOT LIKE 'local://episode/%' AND ` + s.column + ` NOT LIKE 'local://user/%'`
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func down033(tx *sql.Tx) error {
	statements := []struct {
		table      string
		column     string
		entityType string
	}{
		{"media", "poster_path", "media"},
		{"media", "backdrop_path", "media"},
		{"media", "logo_path", "media"},
		{"media", "thumb_path", "media"},
		{"series", "poster_path", "series"},
		{"series", "backdrop_path", "series"},
		{"series", "logo_path", "series"},
		{"series", "thumb_path", "series"},
		{"seasons", "poster_path", "series"},
		{"episodes", "still_path", "episode"},
		{"users", "avatar_path", "user"},
	}

	for _, s := range statements {
		prefix := "local://" + s.entityType + "/"
		stmt := `UPDATE ` + s.table + ` SET ` + s.column + ` = 'local://' || substr(` + s.column + `, ` + strconv.Itoa(len(prefix)+1) + `) WHERE ` + s.column + ` LIKE '` + prefix + `%'`
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
