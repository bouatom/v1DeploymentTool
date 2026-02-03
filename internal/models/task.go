package models

import "time"

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusSuccess  TaskStatus = "success"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCanceled TaskStatus = "canceled"
)

type Task struct {
	ID          string
	Name        string
	Status      TaskStatus
	TargetCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskRun struct {
	ID        string
	TaskID    string
	Status    TaskStatus
	StartedAt time.Time
	EndedAt   time.Time
}
