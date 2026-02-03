package errors

import "strings"

type Code string

const (
	CodeAuthDenied    Code = "auth_denied"
	CodeAuthTimeout   Code = "auth_timeout"
	CodePortClosed    Code = "port_closed"
	CodeUnsupportedOS Code = "unsupported_os"
	CodeInstallFailed Code = "install_failed"
	CodeNetworkIssue  Code = "network_issue"
)

type Detail struct {
	Code        Code
	Message     string
	Remediation string
}

func RemediationSteps(code Code) []string {
	switch code {
	case CodeAuthDenied:
		return []string{
			"Verify the account or key has access on the target.",
			"Check the target's auth policy (sudoers, local admin group, WinRM auth).",
			"Retry with key-based auth first, then password if needed.",
		}
	case CodeAuthTimeout:
		return []string{
			"Confirm the management port is reachable from the controller.",
			"Validate firewall rules and security groups.",
			"Retry after confirming DNS/IP routing.",
		}
	case CodePortClosed:
		return []string{
			"Enable SSH (22) or WinRM HTTPS (5986) on the target.",
			"Restrict access to the controller subnet.",
			"Re-run the assessment scan.",
		}
	case CodeUnsupportedOS:
		return []string{
			"Confirm the OS family is supported.",
			"Ensure the correct management service is installed.",
			"Re-scan to fingerprint the OS.",
		}
	case CodeInstallFailed:
		return []string{
			"Validate the binary URL is reachable from the target.",
			"Confirm write/execute permissions on the destination path.",
			"Re-run with verbose logging to capture command output.",
		}
	case CodeNetworkIssue:
		return []string{
			"Verify target connectivity (ping, DNS lookup).",
			"Confirm routing and VLAN rules.",
			"Re-run scan with a lower aggressiveness setting.",
		}
	default:
		return []string{
			"Review target configuration.",
			"Validate credentials and management ports.",
			"Retry with updated settings.",
		}
	}
}

func RemediationFor(code Code) string {
	steps := RemediationSteps(code)
	if len(steps) == 0 {
		return ""
	}

	return "1) " + strings.Join(steps, " 2) ")
}
