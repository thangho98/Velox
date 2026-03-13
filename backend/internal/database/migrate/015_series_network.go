package migrate

import "database/sql"

// 015: Add network column to series for TVmaze enrichment
func up015(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE series ADD COLUMN network TEXT DEFAULT '';
	`)
	return err
}

func down015(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE series DROP COLUMN network;`)
	return err
}
