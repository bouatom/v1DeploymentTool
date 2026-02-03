package handlers

import "v1-sg-deployment-tool/internal/models"

func osFamilyMatches(os models.TargetOS, family string) bool {
	switch os {
	case models.TargetOSWindows:
		return family == "windows"
	case models.TargetOSMacOS:
		return family == "macos"
	case models.TargetOSLinux:
		return family == "linux"
	default:
		return false
	}
}
