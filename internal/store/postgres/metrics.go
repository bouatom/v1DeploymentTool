package postgres

import (
	"context"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func getMetrics(ctx context.Context, pool queryExec) (store.MetricsSummary, error) {
	var metrics store.MetricsSummary

	err := pool.QueryRow(ctx, `
		SELECT
			COUNT(1) AS total_tasks,
			COUNT(1) FILTER (WHERE status = $1) AS running_tasks,
			COUNT(1) FILTER (WHERE status = $2) AS success_tasks,
			COUNT(1) FILTER (WHERE status = $3) AS failed_tasks,
			COUNT(1) FILTER (WHERE status = $4) AS pending_tasks
		FROM tasks
	`, models.TaskStatusRunning, models.TaskStatusSuccess, models.TaskStatusFailed, models.TaskStatusPending).Scan(
		&metrics.TotalTasks,
		&metrics.RunningTasks,
		&metrics.SuccessTasks,
		&metrics.FailedTasks,
		&metrics.PendingTasks,
	)
	if err != nil {
		return store.MetricsSummary{}, err
	}

	err = pool.QueryRow(ctx, `
		SELECT total_targets, targets_scanned
		FROM scan_summary
		WHERE id = 1
	`).Scan(&metrics.TargetsTotal, &metrics.TargetsScanned)
	if err != nil {
		metrics.TargetsTotal = 0
		metrics.TargetsScanned = 0
	}

	rows, err := pool.Query(ctx, `SELECT code, count FROM failure_reasons`)
	if err != nil {
		return store.MetricsSummary{}, err
	}
	defer rows.Close()

	metrics.FailureReasons = map[string]int{}
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return store.MetricsSummary{}, err
		}
		metrics.FailureReasons[code] = count
	}

	if err := rows.Err(); err != nil {
		return store.MetricsSummary{}, err
	}

	metrics.SuccessByOS = map[string]int{}
	metrics.FailureByOS = map[string]int{}
	metrics.AuthMethods = map[string]int{}

	osRows, err := pool.Query(ctx, `
		SELECT t.os, dr.status, COUNT(1)
		FROM deployment_results dr
		JOIN targets t ON t.id = dr.target_id
		GROUP BY t.os, dr.status
	`)
	if err != nil {
		return store.MetricsSummary{}, err
	}
	defer osRows.Close()

	for osRows.Next() {
		var os string
		var status string
		var count int
		if err := osRows.Scan(&os, &status, &count); err != nil {
			return store.MetricsSummary{}, err
		}
		if status == string(models.TaskStatusSuccess) {
			metrics.SuccessByOS[os] = metrics.SuccessByOS[os] + count
		} else if status == string(models.TaskStatusFailed) {
			metrics.FailureByOS[os] = metrics.FailureByOS[os] + count
		}
	}

	if err := osRows.Err(); err != nil {
		return store.MetricsSummary{}, err
	}

	authRows, err := pool.Query(ctx, `
		SELECT auth_method, COUNT(1)
		FROM deployment_results
		GROUP BY auth_method
	`)
	if err != nil {
		return store.MetricsSummary{}, err
	}
	defer authRows.Close()

	for authRows.Next() {
		var method string
		var count int
		if err := authRows.Scan(&method, &count); err != nil {
			return store.MetricsSummary{}, err
		}
		metrics.AuthMethods[method] = count
	}

	if err := authRows.Err(); err != nil {
		return store.MetricsSummary{}, err
	}

	return metrics, nil
}
