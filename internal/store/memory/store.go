package memory

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"v1-sg-deployment-tool/internal/models"
	storepkg "v1-sg-deployment-tool/internal/store"
)

type Store struct {
	mu            sync.Mutex
	tasks         map[string]models.Task
	runs          map[string]models.TaskRun
	scanSummary   storepkg.ScanSummary
	failureCounts map[string]int
	installers    map[string]models.Installer
}

func NewStore() *Store {
	return &Store{
		tasks:         map[string]models.Task{},
		runs:          map[string]models.TaskRun{},
		scanSummary:   storepkg.ScanSummary{},
		failureCounts: map[string]int{},
		installers:    map[string]models.Installer{},
	}
}

func (store *Store) CreateTask(input storepkg.CreateTaskInput) (models.Task, error) {
	if input.Name == "" {
		return models.Task{}, errors.New("task name is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now().UTC()
	taskID := generateID()
	task := models.Task{
		ID:          taskID,
		Name:        input.Name,
		Status:      models.TaskStatusPending,
		TargetCount: input.TargetCount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	store.tasks[taskID] = task

	return task, nil
}

func (store *Store) ListTasks(options storepkg.ListOptions) ([]models.Task, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	tasks := make([]models.Task, 0, len(store.tasks))
	for _, task := range store.tasks {
		tasks = append(tasks, task)
	}

	return applyTaskListOptions(tasks, options), nil
}

func (store *Store) CreateRun(input storepkg.CreateRunInput) (models.TaskRun, error) {
	if input.TaskID == "" {
		return models.TaskRun{}, errors.New("task id is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	task, ok := store.tasks[input.TaskID]
	if !ok {
		return models.TaskRun{}, errors.New("task not found")
	}

	now := time.Now().UTC()
	runID := generateID()
	run := models.TaskRun{
		ID:        runID,
		TaskID:    input.TaskID,
		Status:    models.TaskStatusRunning,
		StartedAt: now,
	}

	task.Status = models.TaskStatusRunning
	task.UpdatedAt = now
	store.tasks[input.TaskID] = task
	store.runs[runID] = run

	return run, nil
}

func (store *Store) UpdateRun(input storepkg.UpdateRunInput) (models.TaskRun, error) {
	if input.RunID == "" {
		return models.TaskRun{}, errors.New("run id is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	run, ok := store.runs[input.RunID]
	if !ok {
		return models.TaskRun{}, errors.New("run not found")
	}

	run.Status = input.Status
	run.EndedAt = time.Now().UTC()
	store.runs[input.RunID] = run

	task, ok := store.tasks[run.TaskID]
	if ok {
		task.Status = input.Status
		task.UpdatedAt = run.EndedAt
		store.tasks[run.TaskID] = task
		if input.Status == models.TaskStatusFailed {
			store.failureCounts["unknown"] = store.failureCounts["unknown"] + 1
		}
	}

	return run, nil
}

func (store *Store) ListRuns(taskID string) ([]models.TaskRun, error) {
	if taskID == "" {
		return nil, errors.New("task id is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	var runs []models.TaskRun
	for _, run := range store.runs {
		if run.TaskID == taskID {
			runs = append(runs, run)
		}
	}

	return runs, nil
}

func applyTaskListOptions(tasks []models.Task, options storepkg.ListOptions) []models.Task {
	start := options.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(tasks) {
		return []models.Task{}
	}

	end := len(tasks)
	if options.Limit > 0 {
		end = start + options.Limit
		if end > len(tasks) {
			end = len(tasks)
		}
	}

	return tasks[start:end]
}

func (store *Store) RecordScan(input storepkg.ScanInput) (storepkg.ScanSummary, error) {
	if input.TargetCount < 0 || input.TargetsScanned < 0 {
		return storepkg.ScanSummary{}, errors.New("scan values must be non-negative")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.scanSummary = storepkg.ScanSummary{
		TotalTargets:   input.TargetCount,
		TargetsScanned: input.TargetsScanned,
	}

	return store.scanSummary, nil
}

func (store *Store) GetMetrics() (storepkg.MetricsSummary, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	metrics := storepkg.MetricsSummary{
		TotalTasks:     len(store.tasks),
		TargetsTotal:   store.scanSummary.TotalTargets,
		TargetsScanned: store.scanSummary.TargetsScanned,
		FailureReasons: map[string]int{},
		SuccessByOS:    map[string]int{},
		FailureByOS:    map[string]int{},
		AuthMethods:    map[string]int{},
	}

	for _, task := range store.tasks {
		switch task.Status {
		case models.TaskStatusRunning:
			metrics.RunningTasks++
		case models.TaskStatusSuccess:
			metrics.SuccessTasks++
		case models.TaskStatusFailed:
			metrics.FailedTasks++
		case models.TaskStatusPending:
			metrics.PendingTasks++
		}
	}

	for key, count := range store.failureCounts {
		metrics.FailureReasons[key] = count
	}

	return metrics, nil
}

func (store *Store) DeleteDeploymentsBefore(cutoff time.Time) (int64, error) {
	return 0, nil
}

func (store *Store) DeleteRunsBefore(cutoff time.Time) (int64, error) {
	return 0, nil
}

func (store *Store) DeleteAuditLogsBefore(cutoff time.Time) (int64, error) {
	return 0, nil
}

func (store *Store) CreateInstaller(input storepkg.CreateInstallerInput) (models.Installer, error) {
	if input.Filename == "" || input.URL == "" {
		return models.Installer{}, errors.New("filename and url are required")
	}
	if input.PackageType == "" || input.OSFamily == "" {
		return models.Installer{}, errors.New("package type and os family are required")
	}
	if input.Checksum == "" {
		return models.Installer{}, errors.New("checksum is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now().UTC()
	installerID := generateID()
	installer := models.Installer{
		ID:          installerID,
		Filename:    input.Filename,
		URL:         input.URL,
		PackageType: input.PackageType,
		OSFamily:    input.OSFamily,
		Checksum:    input.Checksum,
		CreatedAt:   now,
	}
	store.installers[installerID] = installer

	return installer, nil
}

func (store *Store) GetInstaller(installerID string) (models.Installer, error) {
	if installerID == "" {
		return models.Installer{}, errors.New("installer id is required")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	installer, ok := store.installers[installerID]
	if !ok {
		return models.Installer{}, errors.New("installer not found")
	}

	return installer, nil
}

func generateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}

	return hex.EncodeToString(bytes)
}
