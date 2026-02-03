package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/store"
)

func recordScan(ctx context.Context, pool queryExec, input store.ScanInput) (store.ScanSummary, error) {
	if input.TargetCount < 0 || input.TargetsScanned < 0 {
		return store.ScanSummary{}, errors.New("scan values must be non-negative")
	}

	now := time.Now().UTC()
	_, err := pool.Exec(ctx, `
		INSERT INTO scan_summary (id, total_targets, targets_scanned, updated_at)
		VALUES (1, $1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET total_targets = $1, targets_scanned = $2, updated_at = $3
	`, input.TargetCount, input.TargetsScanned, now)
	if err != nil {
		return store.ScanSummary{}, err
	}

	return store.ScanSummary{
		TotalTargets:   input.TargetCount,
		TargetsScanned: input.TargetsScanned,
	}, nil
}
