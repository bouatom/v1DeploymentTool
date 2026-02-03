package handlers

import (
	"context"
	stdErrors "errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/auth"
	"v1-sg-deployment-tool/internal/errors"
	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/runner"
)

type preflightRequest struct {
	TargetID      string `json:"targetId"`
	CredentialID  string `json:"credentialId"`
	SSHUsername   string `json:"sshUsername"`
	SSHPassword   string `json:"sshPassword"`
	SSHPrivateKey string `json:"sshPrivateKey"`
	WinRMUsername string `json:"winrmUsername"`
	WinRMPassword string `json:"winrmPassword"`
	WinRMPort     int    `json:"winrmPort"`
	WinRMInsecure bool   `json:"winrmInsecure"`
}

type preflightResponse struct {
	TargetID        string `json:"targetId"`
	Success         bool   `json:"success"`
	AuthMethod      string `json:"authMethod,omitempty"`
	ErrorCode       string `json:"errorCode,omitempty"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
	Remediation     string `json:"remediation,omitempty"`
	DurationSeconds int    `json:"durationSeconds"`
}

func (api *API) handlePreflight(c *fiber.Ctx) error {
	var request preflightRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.TargetID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "targetId is required"})
	}

	target, err := api.TargetStore.GetTarget(request.TargetID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	credentials, err := api.resolvePreflightCredentials(request)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	runCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	order := auth.OrderForOS(target.OS)
	if len(order) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "unsupported os"})
	}

	for _, method := range order {
		report, runErr := runPreflightCommand(runCtx, method, targetAddress(target), credentials)
		if runErr == nil && hasNonZeroExit(report.Results) == false {
			return c.JSON(preflightResponse{
				TargetID:        request.TargetID,
				Success:         true,
				AuthMethod:      string(method),
				DurationSeconds: report.DurationSeconds,
			})
		}
		_ = report
	}

	return c.JSON(preflightResponse{
		TargetID:     request.TargetID,
		Success:      false,
		ErrorCode:    string(errors.CodeAuthDenied),
		ErrorMessage: "authentication preflight failed",
		Remediation:  errors.RemediationFor(errors.CodeAuthDenied),
	})
}

func hasNonZeroExit(results []runner.CommandResult) bool {
	for _, result := range results {
		if result.ExitCode != 0 {
			return true
		}
	}
	return false
}

func (api *API) resolvePreflightCredentials(request preflightRequest) (deployCredentials, error) {
	if request.CredentialID != "" {
		credential, err := api.CredentialStore.GetCredential(request.CredentialID)
		if err != nil {
			return deployCredentials{}, err
		}

		switch credential.Kind {
		case models.CredentialKindSSH:
			return deployCredentials{
				SSH: runner.SSHCredentials{
					Username:   credential.Username,
					Password:   credential.Password,
					PrivateKey: credential.PrivateKey,
				},
			}, nil
		case models.CredentialKindWinRM:
			return deployCredentials{
				WinRM: runner.WinRMCredentials{
					Username: credential.Username,
					Password: credential.Password,
					UseHTTPS: true,
					Port:     request.WinRMPort,
					Insecure: request.WinRMInsecure,
				},
			}, nil
		default:
			return deployCredentials{}, stdErrors.New("unsupported credential type")
		}
	}

	return deployCredentials{
		SSH: runner.SSHCredentials{
			Username:   request.SSHUsername,
			Password:   request.SSHPassword,
			PrivateKey: request.SSHPrivateKey,
		},
		WinRM: runner.WinRMCredentials{
			Username: request.WinRMUsername,
			Password: request.WinRMPassword,
			UseHTTPS: true,
			Port:     request.WinRMPort,
			Insecure: request.WinRMInsecure,
		},
	}, nil
}

type deployCredentials struct {
	SSH   runner.SSHCredentials
	WinRM runner.WinRMCredentials
}

func runPreflightCommand(ctx context.Context, method auth.Method, host string, credentials deployCredentials) (runner.RunReport, error) {
	switch method {
	case auth.MethodSSHKey, auth.MethodSSHPassword:
		return runner.SSHRunner{}.RunSSH(ctx, host, []string{"whoami"}, credentials.SSH)
	case auth.MethodWinRMHTTPSCert, auth.MethodWinRMHTTPSUserPW:
		credentials.WinRM.UseHTTPS = true
		return runner.WinRMRunner{}.RunWinRM(ctx, host, []string{"whoami"}, credentials.WinRM)
	default:
		return runner.RunReport{}, stdErrors.New("unsupported auth method")
	}
}
