package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/store"
)

func (store *Store) RecordAudit(input store.AuditInput) error {
	if input.Actor == "" {
		return errors.New("actor is required")
	}
	if input.Action == "" {
		return errors.New("action is required")
	}
	if input.Path == "" {
		return errors.New("path is required")
	}

	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err := store.pool.Exec(context.Background(), `
		INSERT INTO audit_logs (id, actor, role, action, path, status_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, generateID(), input.Actor, input.Role, input.Action, input.Path, input.StatusCode, createdAt)

	return err
}
