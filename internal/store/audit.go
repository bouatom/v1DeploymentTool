package store

import "time"

type AuditStore interface {
	RecordAudit(input AuditInput) error
}

type AuditInput struct {
	Actor      string
	Role       string
	Action     string
	Path       string
	StatusCode int
	CreatedAt  time.Time
}
