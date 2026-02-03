package store

import "v1-sg-deployment-tool/internal/models"

type TaskStore interface {
	CreateTask(input CreateTaskInput) (models.Task, error)
	ListTasks(options ListOptions) ([]models.Task, error)
	CreateRun(input CreateRunInput) (models.TaskRun, error)
	UpdateRun(input UpdateRunInput) (models.TaskRun, error)
	ListRuns(taskID string) ([]models.TaskRun, error)
	RecordScan(input ScanInput) (ScanSummary, error)
	GetMetrics() (MetricsSummary, error)
}

type CreateTaskInput struct {
	Name        string
	TargetCount int
}

type CreateRunInput struct {
	TaskID string
}

type UpdateRunInput struct {
	RunID  string
	Status models.TaskStatus
}

type ScanInput struct {
	TargetCount  int
	TargetsScanned int
}

type ScanSummary struct {
	TotalTargets   int
	TargetsScanned int
}

type MetricsSummary struct {
	TotalTasks     int
	RunningTasks   int
	SuccessTasks   int
	FailedTasks    int
	PendingTasks   int
	TargetsTotal   int
	TargetsScanned int
	FailureReasons map[string]int
	SuccessByOS    map[string]int
	FailureByOS    map[string]int
	AuthMethods    map[string]int
}
