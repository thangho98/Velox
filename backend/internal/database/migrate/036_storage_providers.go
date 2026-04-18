package migrate

import "database/sql"

// 036: Storage providers — one row per cloud account (fshare, future: gdrive, onedrive).
// Credentials are AES-256-GCM encrypted at the application layer using VELOX_CLOUD_SECRET.
// Multiple libraries can share a single provider (e.g. /Movies and /TV Shows folders on
// the same fshare account).
func up036(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE storage_providers (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id               INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider_type         TEXT NOT NULL,
			display_name          TEXT NOT NULL,
			account_email         TEXT NOT NULL,
			credentials_encrypted BLOB NOT NULL,
			account_type          TEXT DEFAULT NULL,
			quota_bytes           INTEGER DEFAULT NULL,
			used_bytes            INTEGER DEFAULT NULL,
			account_expires_at    DATETIME DEFAULT NULL,
			token_expires_at      DATETIME DEFAULT NULL,
			last_refresh_at       DATETIME DEFAULT NULL,
			last_check_at         DATETIME DEFAULT NULL,
			last_error            TEXT DEFAULT NULL,
			created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, provider_type, account_email)
		);

		CREATE INDEX idx_storage_providers_user ON storage_providers(user_id);
		CREATE INDEX idx_storage_providers_type ON storage_providers(provider_type);
		CREATE INDEX idx_storage_providers_token_expires ON storage_providers(token_expires_at)
			WHERE token_expires_at IS NOT NULL;
	`)
	return err
}

func down036(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS storage_providers;`)
	return err
}
