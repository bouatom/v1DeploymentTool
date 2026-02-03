package models

import "time"

type TargetScan struct {
	ID        string
	TargetID  string
	Reachable bool
	OpenPorts []int
	ScannedAt time.Time
}
