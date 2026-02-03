package models

import "time"

type Installer struct {
	ID          string
	Filename    string
	URL         string
	PackageType string
	OSFamily    string
	Checksum    string
	CreatedAt   time.Time
}
