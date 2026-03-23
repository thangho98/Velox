package migrate

import "database/sql"

// 023: Notifications for real-time events and user inbox
func up023(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE notifications (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER DEFAULT NULL REFERENCES users(id) ON DELETE CASCADE,
			type       TEXT NOT NULL CHECK (type IN (
				'scan_complete',
				'media_added',
				'transcode_complete',
				'transcode_failed',
				'subtitle_downloaded',
				'identify_complete',
				'library_watcher'
			)),
			title      TEXT NOT NULL,
			message    TEXT NOT NULL DEFAULT '',
			data       TEXT NOT NULL DEFAULT '{}',
			read       INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			read_at    DATETIME DEFAULT NULL
		);

		-- Index for fetching user's notifications
		CREATE INDEX idx_notifications_user ON notifications(user_id);

		-- Index for unread count query (fast)
		CREATE INDEX idx_notifications_user_read ON notifications(user_id, read) WHERE read = 0;

		-- Index for listing by created_at DESC
		CREATE INDEX idx_notifications_created ON notifications(user_id, created_at DESC);
	`)
	return err
}

func down023(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS notifications;`)
	return err
}
