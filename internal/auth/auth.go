package auth

import "v1-sg-deployment-tool/internal/models"

type Method string

const (
	MethodSSHKey           Method = "ssh_key"
	MethodSSHPassword      Method = "ssh_password"
	MethodWinRMHTTPSCert   Method = "winrm_https_cert"
	MethodWinRMHTTPSUserPW Method = "winrm_https_userpass"
)

type Attempt struct {
	Method    Method
	ErrorCode string
}

func OrderForOS(os models.TargetOS) []Method {
	switch os {
	case models.TargetOSWindows:
		return []Method{MethodWinRMHTTPSCert, MethodWinRMHTTPSUserPW}
	case models.TargetOSLinux, models.TargetOSMacOS:
		return []Method{MethodSSHKey, MethodSSHPassword}
	default:
		return nil
	}
}

func NextMethod(order []Method, attempts []Attempt) (Method, bool) {
	used := map[Method]bool{}
	for _, attempt := range attempts {
		used[attempt.Method] = true
	}

	for _, method := range order {
		if !used[method] {
			return method, true
		}
	}

	return "", false
}
