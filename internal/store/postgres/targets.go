package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func createTarget(ctx context.Context, pool queryExec, input store.CreateTargetInput) (models.Target, error) {
	if input.Hostname == "" && input.IPAddress == "" {
		return models.Target{}, errors.New("hostname or ip address is required")
	}

	now := time.Now().UTC()
	targetID := generateID()

	_, err := pool.Exec(ctx, `
		INSERT INTO targets (id, hostname, ip_address, os, last_seen_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, targetID, input.Hostname, input.IPAddress, input.OS, now, now, now)
	if err != nil {
		return models.Target{}, err
	}

	return models.Target{
		ID:         targetID,
		Hostname:   input.Hostname,
		IPAddress:  input.IPAddress,
		OS:         input.OS,
		LastSeenAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func listTargets(ctx context.Context, pool queryExec, options store.ListOptions) ([]models.Target, error) {
	limit, offset := normalizeListOptions(options)
	rows, err := pool.Query(ctx, `
		SELECT id, hostname, ip_address, os, last_seen_at, created_at, updated_at
		FROM targets
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []models.Target
	for rows.Next() {
		var target models.Target
		if err := rows.Scan(
			&target.ID,
			&target.Hostname,
			&target.IPAddress,
			&target.OS,
			&target.LastSeenAt,
			&target.CreatedAt,
			&target.UpdatedAt,
		); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

func getTarget(ctx context.Context, pool queryExec, targetID string) (models.Target, error) {
	if targetID == "" {
		return models.Target{}, errors.New("target id is required")
	}

	var target models.Target
	err := pool.QueryRow(ctx, `
		SELECT id, hostname, ip_address, os, last_seen_at, created_at, updated_at
		FROM targets
		WHERE id = $1
	`, targetID).Scan(
		&target.ID,
		&target.Hostname,
		&target.IPAddress,
		&target.OS,
		&target.LastSeenAt,
		&target.CreatedAt,
		&target.UpdatedAt,
	)
	if err != nil {
		return models.Target{}, err
	}

	return target, nil
}

func recordTargetScan(ctx context.Context, pool queryExec, input store.TargetScanInput) (models.TargetScan, error) {
	if input.TargetID == "" {
		return models.TargetScan{}, errors.New("target id is required")
	}

	now := time.Now().UTC()
	scanID := generateID()
	openPorts := input.OpenPorts
	if openPorts == nil {
		openPorts = []int{}
	}
	portValues := make([]int32, 0, len(openPorts))
	for _, port := range openPorts {
		portValues = append(portValues, int32(port))
	}
	portArray := pgtype.Array[int32]{
		Elements: portValues,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(portValues)), LowerBound: 1}},
		Valid:    true,
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO target_scans (id, target_id, reachable, open_ports, scanned_at)
		VALUES ($1, $2, $3, $4, $5)
	`, scanID, input.TargetID, input.Reachable, portArray, now)
	if err != nil {
		return models.TargetScan{}, err
	}

	_, err = pool.Exec(ctx, `
		UPDATE targets
		SET last_seen_at = $1, updated_at = $2
		WHERE id = $3
	`, now, now, input.TargetID)
	if err != nil {
		return models.TargetScan{}, err
	}

	return models.TargetScan{
		ID:        scanID,
		TargetID:  input.TargetID,
		Reachable: input.Reachable,
		OpenPorts: openPorts,
		ScannedAt: now,
	}, nil
}
