package migrate

import (
	"database/sql"
)

// 025: No-op marker — superseded by the idempotent rewrite of up021.
//
// Originally intended as a safety net for DBs that manually added tagline /
// metadata_locked columns and would have failed on up021's bare ADD COLUMN.
// That edge case is now handled inside up021 itself (PRAGMA table_info check),
// so 025 carries no schema change. It must remain registered to avoid breaking
// DBs that have already recorded it in schema_migrations.
func up025(tx *sql.Tx) error { return nil }

func down025(tx *sql.Tx) error {
	// No-op: dropping these columns would break running code and the migration
	// may have been the only thing that added them — safest to leave them in place.
	return nil
}

// columnExists returns true if the given column is present in the given table.
// Uses PRAGMA table_info which is available on all SQLite versions we support.
func columnExists(tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
