package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"v1-sg-deployment-tool/internal/store"
)

func listAssessments(ctx context.Context, pool queryExec) ([]store.AssessmentRecord, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			t.hostname,
			t.ip_address,
			t.os,
			t.created_at,
			ts.reachable,
			COALESCE(ts.open_ports, '{}') AS open_ports,
			ts.scanned_at
		FROM targets t
		LEFT JOIN LATERAL (
			SELECT reachable, open_ports, scanned_at
			FROM target_scans
			WHERE target_id = t.id
			ORDER BY scanned_at DESC
			LIMIT 1
		) ts ON true
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []store.AssessmentRecord
	for rows.Next() {
		var record store.AssessmentRecord
		var reachable pgtype.Bool
		var scannedAt pgtype.Timestamptz
		var openPorts []int

		if err := rows.Scan(
			&record.TargetID,
			&record.Hostname,
			&record.IPAddress,
			&record.OS,
			&record.CreatedAt,
			&reachable,
			&openPorts,
			&scannedAt,
		); err != nil {
			return nil, err
		}

		if reachable.Valid {
			value := reachable.Bool
			record.Reachable = &value
		}

		if scannedAt.Valid {
			value := scannedAt.Time.UTC()
			record.ScannedAt = &value
		}

		record.OpenPorts = openPorts

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (store *Store) ListAssessments() ([]store.AssessmentRecord, error) {
	return listAssessments(context.Background(), store.pool)
}
