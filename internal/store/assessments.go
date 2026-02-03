package store

import (
	"time"

	"v1-sg-deployment-tool/internal/models"
)

type AssessmentStore interface {
	ListAssessments() ([]AssessmentRecord, error)
}

type AssessmentRecord struct {
	TargetID   string
	Hostname   string
	IPAddress  string
	OS         models.TargetOS
	Reachable  *bool
	OpenPorts  []int
	ScannedAt  *time.Time
	CreatedAt  time.Time
}
