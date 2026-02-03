package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func createTask(ctx context.Context, pool queryExec, input store.CreateTaskInput) (models.Task, error) {
	if input.Name == "" {
		return models.Task{}, errors.New("task name is required")
	}

	now := time.Now().UTC()
	taskID := generateID()

	_, err := pool.Exec(ctx, `
		INSERT INTO tasks (id, name, status, target_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, taskID, input.Name, models.TaskStatusPending, input.TargetCount, now, now)
	if err != nil {
		return models.Task{}, err
	}

	return models.Task{
		ID:          taskID,
		Name:        input.Name,
		Status:      models.TaskStatusPending,
		TargetCount: input.TargetCount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func listTasks(ctx context.Context, pool queryExec, options store.ListOptions) ([]models.Task, error) {
	limit, offset := normalizeListOptions(options)
	rows, err := pool.Query(ctx, `
		SELECT id, name, status, target_count, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.TargetCount, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
