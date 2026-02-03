package models

import "time"

type DeploymentResult struct {
	ID          string
	TaskRunID   string
	TargetID    string
	Status      TaskStatus
	AuthMethod  string
	ErrorCode   string
	ErrorMessage string
	Remediation string
	FinishedAt  time.Time
}
