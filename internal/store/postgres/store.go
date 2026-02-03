package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

type Store struct {
	pool             *pgxpool.Pool
	credentialsKey   string
	credentialsKeyID string
}

func NewStore(pool *pgxpool.Pool, credentialsKey string, credentialsKeyID string) (*Store, error) {
	if pool == nil {
		return nil, errors.New("pool is required")
	}
	if credentialsKey == "" {
		return nil, errors.New("credentials key is required")
	}
	if credentialsKeyID == "" {
		credentialsKeyID = "default"
	}

	return &Store{
		pool:             pool,
		credentialsKey:   credentialsKey,
		credentialsKeyID: credentialsKeyID,
	}, nil
}

func (store *Store) CreateTask(input store.CreateTaskInput) (models.Task, error) {
	return createTask(context.Background(), store.pool, input)
}

func (store *Store) ListTasks(options store.ListOptions) ([]models.Task, error) {
	return listTasks(context.Background(), store.pool, options)
}

func (store *Store) CreateRun(input store.CreateRunInput) (models.TaskRun, error) {
	return createRun(context.Background(), store.pool, input)
}

func (store *Store) UpdateRun(input store.UpdateRunInput) (models.TaskRun, error) {
	return updateRun(context.Background(), store.pool, input)
}

func (store *Store) ListRuns(taskID string) ([]models.TaskRun, error) {
	return listRuns(context.Background(), store.pool, taskID)
}

func (store *Store) RecordScan(input store.ScanInput) (store.ScanSummary, error) {
	return recordScan(context.Background(), store.pool, input)
}

func (store *Store) GetMetrics() (store.MetricsSummary, error) {
	return getMetrics(context.Background(), store.pool)
}

func (store *Store) CreateTarget(input store.CreateTargetInput) (models.Target, error) {
	return createTarget(context.Background(), store.pool, input)
}

func (store *Store) ListTargets(options store.ListOptions) ([]models.Target, error) {
	return listTargets(context.Background(), store.pool, options)
}

func (store *Store) GetTarget(targetID string) (models.Target, error) {
	return getTarget(context.Background(), store.pool, targetID)
}

func (store *Store) RecordTargetScan(input store.TargetScanInput) (models.TargetScan, error) {
	return recordTargetScan(context.Background(), store.pool, input)
}

func (store *Store) CreateDeploymentResult(input store.CreateDeploymentResultInput) (models.DeploymentResult, error) {
	return createDeploymentResult(context.Background(), store.pool, input)
}

func (store *Store) ListDeploymentResults(targetID string, options store.ListOptions) ([]models.DeploymentResult, error) {
	return listDeploymentResults(context.Background(), store.pool, targetID, options)
}

func (store *Store) ListDeploymentResultsByTask(taskID string, options store.ListOptions) ([]store.DeploymentResultDetail, error) {
	return listDeploymentResultsByTask(context.Background(), store.pool, taskID, options)
}
