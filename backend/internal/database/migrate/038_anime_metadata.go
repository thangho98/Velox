package migrate

import "database/sql"

// 038: Add AniList-specific metadata fields to media and series.
// libraries.type is already plain TEXT, so no schema change is needed to store
// the new "anime" library type.
func up038(tx *sql.Tx) error {
	checks := []struct {
		table  string
		column string
		ddl    string
	}{
		{"media", "anilist_id", `ALTER TABLE media ADD COLUMN anilist_id INTEGER DEFAULT NULL`},
		{"media", "romaji_title", `ALTER TABLE media ADD COLUMN romaji_title TEXT NOT NULL DEFAULT ''`},
		{"media", "studio", `ALTER TABLE media ADD COLUMN studio TEXT NOT NULL DEFAULT ''`},
		{"series", "anilist_id", `ALTER TABLE series ADD COLUMN anilist_id INTEGER DEFAULT NULL`},
		{"series", "romaji_title", `ALTER TABLE series ADD COLUMN romaji_title TEXT NOT NULL DEFAULT ''`},
		{"series", "studio", `ALTER TABLE series ADD COLUMN studio TEXT NOT NULL DEFAULT ''`},
	}
	for _, c := range checks {
		exists, err := columnExists(tx, c.table, c.column)
		if err != nil {
			return err
		}
		if !exists {
			if _, err := tx.Exec(c.ddl); err != nil {
				return err
			}
		}
	}

	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_media_anilist ON media(anilist_id) WHERE anilist_id IS NOT NULL`); err != nil {
		return err
	}
	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_series_anilist ON series(anilist_id) WHERE anilist_id IS NOT NULL`); err != nil {
		return err
	}
	return nil
}

func down038(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP INDEX IF EXISTS idx_media_anilist;
		DROP INDEX IF EXISTS idx_series_anilist;
		ALTER TABLE media DROP COLUMN anilist_id;
		ALTER TABLE media DROP COLUMN romaji_title;
		ALTER TABLE media DROP COLUMN studio;
		ALTER TABLE series DROP COLUMN anilist_id;
		ALTER TABLE series DROP COLUMN romaji_title;
		ALTER TABLE series DROP COLUMN studio;
	`)
	return err
}
