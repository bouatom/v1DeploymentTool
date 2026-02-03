package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusRunning  Status = "running"
	StatusSuccess  Status = "success"
	StatusFailed   Status = "failed"
)

type Job struct {
	ID        string
	Kind      string
	Status    Status
	Error     string
	StartedAt time.Time
	EndedAt   time.Time
}

type Handler func(ctx context.Context) error

type Queue struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	handlersByKind map[string]Handler
	handlersByJob  map[string]Handler
	workChan chan string
}

func NewQueue(workerCount int) *Queue {
	queue := &Queue{
		jobs:     map[string]*Job{},
		handlersByKind: map[string]Handler{},
		handlersByJob:  map[string]Handler{},
		workChan: make(chan string, 100),
	}

	if workerCount <= 0 {
		workerCount = 2
	}

	for i := 0; i < workerCount; i++ {
		go queue.worker()
	}

	return queue
}

func (queue *Queue) Register(kind string, handler Handler) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	queue.handlersByKind[kind] = handler
}

func (queue *Queue) Enqueue(kind string) (Job, error) {
	queue.mu.Lock()
	handler := queue.handlersByKind[kind]
	queue.mu.Unlock()

	if handler == nil {
		return Job{}, errors.New("handler not registered")
	}

	return queue.EnqueueWithHandler(kind, handler)
}

func (queue *Queue) EnqueueWithHandler(kind string, handler Handler) (Job, error) {
	if handler == nil {
		return Job{}, errors.New("handler is required")
	}

	job := &Job{
		ID:     generateID(),
		Kind:   kind,
		Status: StatusPending,
	}

	queue.mu.Lock()
	queue.jobs[job.ID] = job
	queue.handlersByJob[job.ID] = handler
	queue.mu.Unlock()

	queue.workChan <- job.ID

	return *job, nil
}

func (queue *Queue) GetJob(jobID string) (Job, error) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	job, ok := queue.jobs[jobID]
	if !ok {
		return Job{}, errors.New("job not found")
	}

	return *job, nil
}

func (queue *Queue) worker() {
	for jobID := range queue.workChan {
		queue.mu.Lock()
		job := queue.jobs[jobID]
		handler := queue.handlersByJob[job.ID]
		job.Status = StatusRunning
		job.StartedAt = time.Now().UTC()
		queue.mu.Unlock()

		err := handler(context.Background())

		queue.mu.Lock()
		if err != nil {
			job.Status = StatusFailed
			job.Error = err.Error()
		} else {
			job.Status = StatusSuccess
		}
		job.EndedAt = time.Now().UTC()
		queue.mu.Unlock()
	}
}
