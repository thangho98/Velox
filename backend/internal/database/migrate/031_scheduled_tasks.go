package migrate

import "database/sql"

func up031(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE scheduled_tasks (
			name TEXT PRIMARY KEY,
			last_run DATETIME,
			interval_seconds INTEGER NOT NULL
		);
	`)
	return err
}

func down031(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS scheduled_tasks;`)
	return err
}
