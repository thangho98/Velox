package migrate

import "database/sql"

// 037: Extend libraries to support cloud sources.
//   - storage_provider_id (nullable): link to storage_providers row; NULL = local
//   - source_url (nullable): provider-specific folder URL (e.g. https://www.fshare.vn/folder/XYZ)
//
// Existing local libraries are unaffected (both columns default NULL).
func up037(tx *sql.Tx) error {
	checks := []struct {
		table  string
		column string
		ddl    string
	}{
		{"libraries", "storage_provider_id", `ALTER TABLE libraries ADD COLUMN storage_provider_id INTEGER REFERENCES storage_providers(id) ON DELETE SET NULL`},
		{"libraries", "source_url", `ALTER TABLE libraries ADD COLUMN source_url TEXT DEFAULT NULL`},
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

	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_libraries_provider ON libraries(storage_provider_id)`); err != nil {
		return err
	}
	return nil
}

func down037(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP INDEX IF EXISTS idx_libraries_provider;
		ALTER TABLE libraries DROP COLUMN storage_provider_id;
		ALTER TABLE libraries DROP COLUMN source_url;
	`)
	return err
}
