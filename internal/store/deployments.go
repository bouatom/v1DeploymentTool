package store

import "v1-sg-deployment-tool/internal/models"

type DeploymentStore interface {
	CreateDeploymentResult(input CreateDeploymentResultInput) (models.DeploymentResult, error)
	ListDeploymentResults(targetID string, options ListOptions) ([]models.DeploymentResult, error)
	ListDeploymentResultsByTask(taskID string, options ListOptions) ([]DeploymentResultDetail, error)
}

type CreateDeploymentResultInput struct {
	TaskRunID   string
	TargetID    string
	Status      models.TaskStatus
	AuthMethod  string
	ErrorCode   string
	ErrorMessage string
	Remediation string
}

type DeploymentResultDetail struct {
	ID           string
	TaskRunID    string
	TargetID     string
	TargetLabel  string
	TargetOS     models.TargetOS
	Status       models.TaskStatus
	AuthMethod   string
	ErrorCode    string
	ErrorMessage string
	Remediation  string
	FinishedAt   string
}
