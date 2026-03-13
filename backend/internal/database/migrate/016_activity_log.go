package migrate

import "database/sql"

// 016: Activity log for admin dashboard
func up016(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE activity_log (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER REFERENCES users(id) ON DELETE SET NULL,
			action       TEXT NOT NULL,
			media_id     INTEGER DEFAULT NULL,
			details_json TEXT DEFAULT '{}',
			ip_address   TEXT DEFAULT '',
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_activity_user ON activity_log(user_id);
		CREATE INDEX idx_activity_action ON activity_log(action);
		CREATE INDEX idx_activity_created ON activity_log(created_at DESC);
	`)
	return err
}

func down016(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS activity_log;`)
	return err
}
