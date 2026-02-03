package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func createDeploymentResult(ctx context.Context, pool queryExec, input store.CreateDeploymentResultInput) (models.DeploymentResult, error) {
	if input.TargetID == "" {
		return models.DeploymentResult{}, errors.New("target id is required")
	}

	now := time.Now().UTC()
	resultID := generateID()

	_, err := pool.Exec(ctx, `
		INSERT INTO deployment_results (
			id, task_run_id, target_id, status, auth_method, error_code, error_message, remediation, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, resultID, input.TaskRunID, input.TargetID, input.Status, input.AuthMethod, input.ErrorCode, input.ErrorMessage, input.Remediation, now)
	if err != nil {
		return models.DeploymentResult{}, err
	}

	return models.DeploymentResult{
		ID:           resultID,
		TaskRunID:    input.TaskRunID,
		TargetID:     input.TargetID,
		Status:       input.Status,
		AuthMethod:   input.AuthMethod,
		ErrorCode:    input.ErrorCode,
		ErrorMessage: input.ErrorMessage,
		Remediation:  input.Remediation,
		FinishedAt:   now,
	}, nil
}

func listDeploymentResults(ctx context.Context, pool queryExec, targetID string, options store.ListOptions) ([]models.DeploymentResult, error) {
	if targetID == "" {
		return nil, errors.New("target id is required")
	}

	limit, offset := normalizeListOptions(options)
	rows, err := pool.Query(ctx, `
		SELECT id, task_run_id, target_id, status, auth_method, error_code, error_message, remediation, finished_at
		FROM deployment_results
		WHERE target_id = $1
		ORDER BY finished_at DESC
		LIMIT $2 OFFSET $3
	`, targetID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.DeploymentResult
	for rows.Next() {
		var result models.DeploymentResult
		if err := rows.Scan(
			&result.ID,
			&result.TaskRunID,
			&result.TargetID,
			&result.Status,
			&result.AuthMethod,
			&result.ErrorCode,
			&result.ErrorMessage,
			&result.Remediation,
			&result.FinishedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func listDeploymentResultsByTask(ctx context.Context, pool queryExec, taskID string, options store.ListOptions) ([]store.DeploymentResultDetail, error) {
	if taskID == "" {
		return nil, errors.New("task id is required")
	}

	limit, offset := normalizeListOptions(options)
	rows, err := pool.Query(ctx, `
		SELECT
			dr.id,
			dr.task_run_id,
			dr.target_id,
			COALESCE(t.hostname, t.ip_address) AS target_label,
			t.os,
			dr.status,
			dr.auth_method,
			dr.error_code,
			dr.error_message,
			dr.remediation,
			dr.finished_at
		FROM deployment_results dr
		JOIN task_runs tr ON tr.id = dr.task_run_id
		JOIN targets t ON t.id = dr.target_id
		WHERE tr.task_id = $1
		ORDER BY dr.finished_at DESC
		LIMIT $2 OFFSET $3
	`, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []store.DeploymentResultDetail
	for rows.Next() {
		var result store.DeploymentResultDetail
		var finishedAt time.Time
		if err := rows.Scan(
			&result.ID,
			&result.TaskRunID,
			&result.TargetID,
			&result.TargetLabel,
			&result.TargetOS,
			&result.Status,
			&result.AuthMethod,
			&result.ErrorCode,
			&result.ErrorMessage,
			&result.Remediation,
			&finishedAt,
		); err != nil {
			return nil, err
		}
		result.FinishedAt = finishedAt.UTC().Format(time.RFC3339)
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Store methods are defined in store.go to keep API surface centralized.
