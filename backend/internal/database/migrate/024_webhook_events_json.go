package migrate

import "database/sql"

// 024: Backfill webhook events column from CSV to JSON array format.
// Rows written before the JSON-array fix may contain comma-separated values
// like "scan_complete,media_added" instead of '["scan_complete","media_added"]'.
// json_each() treats those as malformed JSON, breaking dispatch and the UI.
// We convert any row where json_valid(events) is false by splitting on commas
// and rebuilding as a JSON array via json_group_array.
func up024(tx *sql.Tx) error {
	// Fetch all rows where events is not valid JSON.
	rows, err := tx.Query(`SELECT id, events FROM webhooks WHERE NOT json_valid(events)`)
	if err != nil {
		return err
	}
	type row struct {
		id     int64
		events string
	}
	var bad []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.events); err != nil {
			rows.Close()
			return err
		}
		bad = append(bad, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// For each bad row, split on comma and rebuild as JSON array.
	for _, r := range bad {
		// Use SQLite's own json_group_array over a WITH RECURSIVE split to
		// avoid reimplementing CSV splitting in Go.  The CTE approach works
		// reliably for simple comma-separated strings without quotes.
		_, err := tx.Exec(`
			UPDATE webhooks
			SET events = (
				WITH RECURSIVE split(word, rest) AS (
					SELECT
						TRIM(SUBSTR(?, 1, CASE WHEN INSTR(?, ',') > 0 THEN INSTR(?, ',') - 1 ELSE LENGTH(?) END)),
						CASE WHEN INSTR(?, ',') > 0 THEN SUBSTR(?, INSTR(?, ',') + 1) ELSE NULL END
					UNION ALL
					SELECT
						TRIM(SUBSTR(rest, 1, CASE WHEN INSTR(rest, ',') > 0 THEN INSTR(rest, ',') - 1 ELSE LENGTH(rest) END)),
						CASE WHEN INSTR(rest, ',') > 0 THEN SUBSTR(rest, INSTR(rest, ',') + 1) ELSE NULL END
					FROM split WHERE rest IS NOT NULL
				)
				SELECT json_group_array(word) FROM split WHERE word != ''
			)
			WHERE id = ?`,
			r.events, r.events, r.events, r.events,
			r.events, r.events, r.events,
			r.id,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func down024(tx *sql.Tx) error {
	// Intentionally no-op: converting JSON back to CSV would be lossy and
	// the old CSV format was never the intended schema.
	return nil
}
