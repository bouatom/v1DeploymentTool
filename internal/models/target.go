package models

import "time"

type TargetOS string

const (
	TargetOSUnknown TargetOS = "unknown"
	TargetOSLinux   TargetOS = "linux"
	TargetOSMacOS   TargetOS = "macos"
	TargetOSWindows TargetOS = "windows"
)

type Target struct {
	ID         string
	Hostname   string
	IPAddress  string
	OS         TargetOS
	LastSeenAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
