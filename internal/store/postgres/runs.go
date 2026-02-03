package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func createRun(ctx context.Context, pool queryExec, input store.CreateRunInput) (models.TaskRun, error) {
	if input.TaskID == "" {
		return models.TaskRun{}, errors.New("task id is required")
	}

	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tasks WHERE id = $1)`, input.TaskID).Scan(&exists)
	if err != nil {
		return models.TaskRun{}, err
	}
	if !exists {
		return models.TaskRun{}, errors.New("task not found")
	}

	now := time.Now().UTC()
	runID := generateID()

	_, err = pool.Exec(ctx, `
		INSERT INTO task_runs (id, task_id, status, started_at)
		VALUES ($1, $2, $3, $4)
	`, runID, input.TaskID, models.TaskStatusRunning, now)
	if err != nil {
		return models.TaskRun{}, err
	}

	_, err = pool.Exec(ctx, `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, models.TaskStatusRunning, now, input.TaskID)
	if err != nil {
		return models.TaskRun{}, err
	}

	return models.TaskRun{
		ID:        runID,
		TaskID:    input.TaskID,
		Status:    models.TaskStatusRunning,
		StartedAt: now,
	}, nil
}

func updateRun(ctx context.Context, pool queryExec, input store.UpdateRunInput) (models.TaskRun, error) {
	if input.RunID == "" {
		return models.TaskRun{}, errors.New("run id is required")
	}

	var run models.TaskRun
	err := pool.QueryRow(ctx, `
		SELECT id, task_id, status, started_at, ended_at
		FROM task_runs
		WHERE id = $1
	`, input.RunID).Scan(&run.ID, &run.TaskID, &run.Status, &run.StartedAt, &run.EndedAt)
	if err != nil {
		return models.TaskRun{}, err
	}

	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		UPDATE task_runs
		SET status = $1, ended_at = $2
		WHERE id = $3
	`, input.Status, now, input.RunID)
	if err != nil {
		return models.TaskRun{}, err
	}

	_, err = pool.Exec(ctx, `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, input.Status, now, run.TaskID)
	if err != nil {
		return models.TaskRun{}, err
	}

	if input.Status == models.TaskStatusFailed {
		_, err = pool.Exec(ctx, `
			INSERT INTO failure_reasons (code, count)
			VALUES ($1, 1)
			ON CONFLICT (code) DO UPDATE SET count = failure_reasons.count + 1
		`, "unknown")
		if err != nil {
			return models.TaskRun{}, err
		}
	}

	run.Status = input.Status
	run.EndedAt = now

	return run, nil
}

func listRuns(ctx context.Context, pool queryExec, taskID string) ([]models.TaskRun, error) {
	if taskID == "" {
		return nil, errors.New("task id is required")
	}

	rows, err := pool.Query(ctx, `
		SELECT id, task_id, status, started_at, ended_at
		FROM task_runs
		WHERE task_id = $1
		ORDER BY started_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.TaskRun
	for rows.Next() {
		var run models.TaskRun
		if err := rows.Scan(&run.ID, &run.TaskID, &run.Status, &run.StartedAt, &run.EndedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}
