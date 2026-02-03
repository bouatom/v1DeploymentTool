package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

type assessmentResponse struct {
	TargetID         string   `json:"targetId"`
	Label            string   `json:"label"`
	OS               models.TargetOS `json:"os"`
	Reachable        *bool    `json:"reachable"`
	OpenPorts        []int    `json:"openPorts"`
	PredictedSuccess int      `json:"predictedSuccess"`
	SecureMethod     string   `json:"secureMethod"`
	Guidelines       []string `json:"guidelines"`
	ScannedAt        string   `json:"scannedAt"`
}

func (api *API) handleListAssessments(c *fiber.Ctx) error {
	records, err := api.AssessmentStore.ListAssessments()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	responses := make([]assessmentResponse, 0, len(records))
	for _, record := range records {
		response := buildAssessment(record)
		responses = append(responses, response)
	}

	return c.JSON(responses)
}

func buildAssessment(record store.AssessmentRecord) assessmentResponse {
	label := record.Hostname
	if label == "" {
		label = record.IPAddress
	}

	score := predictSuccess(record)
	method, guidelines := secureGuidelines(record)

	scannedAt := ""
	if record.ScannedAt != nil {
		scannedAt = record.ScannedAt.Format("2006-01-02 15:04:05 MST")
	}

	return assessmentResponse{
		TargetID:         record.TargetID,
		Label:            label,
		OS:               record.OS,
		Reachable:        record.Reachable,
		OpenPorts:        record.OpenPorts,
		PredictedSuccess: score,
		SecureMethod:     method,
		Guidelines:       guidelines,
		ScannedAt:        scannedAt,
	}
}

func predictSuccess(record store.AssessmentRecord) int {
	if record.Reachable == nil {
		return 0
	}

	if !*record.Reachable {
		return 5
	}

	switch record.OS {
	case models.TargetOSWindows:
		if hasPort(record.OpenPorts, 5986) {
			return 90
		}
		if hasPort(record.OpenPorts, 5985) {
			return 70
		}
	case models.TargetOSLinux, models.TargetOSMacOS:
		if hasPort(record.OpenPorts, 22) {
			return 88
		}
	}

	return 40
}

func secureGuidelines(record store.AssessmentRecord) (string, []string) {
	if record.Reachable == nil || !*record.Reachable {
		return "No secure channel detected", []string{
			"Verify the host is powered on and reachable from the controller.",
			"Check firewall rules and routing from the controller subnet.",
			"Re-run the assessment scan after connectivity is confirmed.",
		}
	}

	switch record.OS {
	case models.TargetOSWindows:
		if hasPort(record.OpenPorts, 5986) {
			return "WinRM HTTPS (certificate)", []string{
				"Use WinRM HTTPS on port 5986 with certificate authentication.",
				"Confirm the certificate chain is trusted by the controller.",
				"Restrict WinRM to the controller subnet only.",
			}
		}
		if hasPort(record.OpenPorts, 5985) {
			return "WinRM HTTPS (user/password)", []string{
				"Enable WinRM HTTPS and use strong, rotated credentials.",
				"Transition to certificate authentication for stronger assurance.",
				"Restrict WinRM to the controller subnet only.",
			}
		}
		return "Windows management port missing", []string{
			"Enable WinRM HTTPS and allow it through the firewall.",
			"Restrict access to the controller subnet.",
			"Validate Windows OpenSSH if WinRM cannot be enabled.",
		}
	case models.TargetOSLinux, models.TargetOSMacOS:
		if hasPort(record.OpenPorts, 22) {
			return "SSH key authentication", []string{
				"Use SSH keys with agent forwarding disabled.",
				"Limit SSH access to the controller subnet.",
				"Disable password auth after validating keys.",
			}
		}
		return "SSH not detected", []string{
			"Enable SSH on port 22 and restrict access.",
			"Confirm host firewall allows the controller subnet.",
			"Re-run the assessment scan after enabling SSH.",
		}
	default:
		return "OS not identified", []string{
			"Run a fresh scan to fingerprint the operating system.",
			"Confirm management ports are open and reachable.",
			"Validate DNS and IP routing from the controller.",
		}
	}
}

func hasPort(ports []int, port int) bool {
	for _, value := range ports {
		if value == port {
			return true
		}
	}
	return false
}
