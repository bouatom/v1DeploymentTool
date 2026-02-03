package store

import "v1-sg-deployment-tool/internal/models"

type TargetStore interface {
	CreateTarget(input CreateTargetInput) (models.Target, error)
	ListTargets(options ListOptions) ([]models.Target, error)
	GetTarget(targetID string) (models.Target, error)
	RecordTargetScan(input TargetScanInput) (models.TargetScan, error)
}

type CreateTargetInput struct {
	Hostname  string
	IPAddress string
	OS        models.TargetOS
}

type TargetScanInput struct {
	TargetID  string
	Reachable bool
	OpenPorts []int
}
