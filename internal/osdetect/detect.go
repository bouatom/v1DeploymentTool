package osdetect

import "v1-sg-deployment-tool/internal/models"

func DetectFromPorts(openPorts []int) models.TargetOS {
	if hasPort(openPorts, 5986) || hasPort(openPorts, 5985) {
		return models.TargetOSWindows
	}
	if hasPort(openPorts, 22) {
		return models.TargetOSLinux
	}

	return models.TargetOSUnknown
}

func hasPort(openPorts []int, port int) bool {
	for _, candidate := range openPorts {
		if candidate == port {
			return true
		}
	}
	return false
}
