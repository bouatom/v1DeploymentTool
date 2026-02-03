package postgres

import (
	"context"
	"time"
)

func (store *Store) DeleteDeploymentsBefore(cutoff time.Time) (int64, error) {
	tag, err := store.pool.Exec(context.Background(), `
		DELETE FROM deployment_results
		WHERE finished_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func (store *Store) DeleteRunsBefore(cutoff time.Time) (int64, error) {
	tag, err := store.pool.Exec(context.Background(), `
		DELETE FROM task_runs
		WHERE ended_at IS NOT NULL AND ended_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func (store *Store) DeleteAuditLogsBefore(cutoff time.Time) (int64, error) {
	tag, err := store.pool.Exec(context.Background(), `
		DELETE FROM audit_logs
		WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
