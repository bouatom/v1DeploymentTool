package handlers

import (
	"context"
	stdErrors "errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/auth"
	"v1-sg-deployment-tool/internal/deploy"
	"v1-sg-deployment-tool/internal/errors"
	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/runner"
	"v1-sg-deployment-tool/internal/store"
)

type executeDeployRequest struct {
	TaskRunID       string   `json:"taskRunId"`
	TargetID        string   `json:"targetId"`
	CredentialID    string   `json:"credentialId"`
	InstallerID     string   `json:"installerId"`
	BinaryURL       string   `json:"binaryUrl"`
	DestinationPath string   `json:"destinationPath"`
	PostInstallArgs []string `json:"postInstallArgs"`
	ExecuteOnInstall bool    `json:"executeOnInstall"`
	PackageType      deploy.PackageType `json:"packageType"`
	Checksum         string            `json:"checksum"`
	ChecksumAlg      string            `json:"checksumAlg"`
	ExpectedArch     string            `json:"expectedArch"`
	MinFreeMB        int               `json:"minFreeMB"`
	ProxyURL         string            `json:"proxyUrl"`
	RequiresReboot   bool              `json:"requiresReboot"`
	AllowReboot      bool              `json:"allowReboot"`
	SSHUsername     string   `json:"sshUsername"`
	SSHPassword     string   `json:"sshPassword"`
	SSHPrivateKey   string   `json:"sshPrivateKey"`
	WinRMUsername   string   `json:"winrmUsername"`
	WinRMPassword   string   `json:"winrmPassword"`
	WinRMPort       int      `json:"winrmPort"`
	WinRMInsecure   bool     `json:"winrmInsecure"`
}

type executeDeployResponse struct {
	TargetID         string          `json:"targetId"`
	Status           models.TaskStatus `json:"status"`
	AuthMethod       string          `json:"authMethod"`
	ErrorCode        string          `json:"errorCode,omitempty"`
	ErrorMessage     string          `json:"errorMessage,omitempty"`
	Remediation      string          `json:"remediation,omitempty"`
	DurationSeconds  int             `json:"durationSeconds"`
}

func (api *API) handleExecuteDeploy(c *fiber.Ctx) error {
	var request executeDeployRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	result, execErr := api.executeDeployWork(request)
	if execErr != nil && result.TargetID == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": execErr.Error()})
	}

	status := models.TaskStatusSuccess
	errorCode := ""
	errorMessage := ""
	remediation := ""

	if execErr != nil {
		status = models.TaskStatusFailed
		if result.ErrorDetail != nil {
			errorCode = string(result.ErrorDetail.Code)
			errorMessage = result.ErrorDetail.Message
			remediation = result.ErrorDetail.Remediation
		} else {
			errorCode = string(errors.CodeInstallFailed)
			errorMessage = execErr.Error()
			remediation = errors.RemediationFor(errors.CodeInstallFailed)
		}
	}

	return c.JSON(executeDeployResponse{
		TargetID:        result.TargetID,
		Status:          status,
		AuthMethod:      string(result.Method),
		ErrorCode:       errorCode,
		ErrorMessage:    errorMessage,
		Remediation:     remediation,
		DurationSeconds: result.Report.DurationSeconds,
	})
}

func (api *API) handleExecuteDeployAsync(c *fiber.Ctx) error {
	if api.Queue == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "queue not available"})
	}

	var request executeDeployRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	job, err := api.Queue.EnqueueWithHandler("deploy", func(ctx context.Context) error {
		_, err := api.executeDeployWork(request)
		return err
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(job)
}

func (api *API) handleListDeployments(c *fiber.Ctx) error {
	targetID := c.Params("targetId")
	results, err := api.DeploymentStore.ListDeploymentResults(targetID, parseListOptions(c))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if results == nil {
		results = []models.DeploymentResult{}
	}

	return c.JSON(results)
}

func (api *API) resolveCredentials(request executeDeployRequest) (deploy.CredentialSet, error) {
	if request.CredentialID != "" {
		credential, err := api.CredentialStore.GetCredential(request.CredentialID)
		if err != nil {
			return deploy.CredentialSet{}, err
		}

		switch credential.Kind {
		case models.CredentialKindSSH:
			return deploy.CredentialSet{
				SSH: runner.SSHCredentials{
					Username:   credential.Username,
					Password:   credential.Password,
					PrivateKey: credential.PrivateKey,
				},
			}, nil
		case models.CredentialKindWinRM:
			return deploy.CredentialSet{
				WinRM: runner.WinRMCredentials{
					Username: credential.Username,
					Password: credential.Password,
					UseHTTPS: true,
					Port:     request.WinRMPort,
					Insecure: request.WinRMInsecure,
				},
			}, nil
		default:
			return deploy.CredentialSet{}, stdErrors.New("unsupported credential type")
		}
	}

	return deploy.CredentialSet{
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

type deployWorkResult struct {
	TargetID string
	Method   auth.Method
	Report   runner.RunReport
	ErrorDetail   *errors.Detail
}

func (api *API) executeDeployWork(request executeDeployRequest) (deployWorkResult, error) {
	if request.TargetID == "" {
		return deployWorkResult{}, stdErrors.New("targetId is required")
	}
	if request.BinaryURL == "" && request.InstallerID == "" {
		return deployWorkResult{}, stdErrors.New("binaryUrl or installerId is required")
	}

	target, err := api.TargetStore.GetTarget(request.TargetID)
	if err != nil {
		return deployWorkResult{}, err
	}

	credentials, err := api.resolveCredentials(request)
	if err != nil {
		return deployWorkResult{}, err
	}

	engine := deploy.Engine{
		Runner: runner.MultiRunner{
			SSH:   runner.SSHRunner{},
			WinRM: runner.WinRMRunner{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	installRequest := deploy.InstallRequest{
		OS:               target.OS,
		BinaryURL:        request.BinaryURL,
		DestinationPath:  request.DestinationPath,
		PostInstallArgs:  request.PostInstallArgs,
		ExecuteOnInstall: request.ExecuteOnInstall,
		PackageType:      request.PackageType,
		Checksum:         request.Checksum,
		ChecksumAlg:      request.ChecksumAlg,
		ExpectedArch:     request.ExpectedArch,
		MinFreeMB:        request.MinFreeMB,
		ProxyURL:         request.ProxyURL,
		RequiresReboot:   request.RequiresReboot,
		AllowReboot:      request.AllowReboot,
	}

	if request.InstallerID != "" {
		installer, err := api.InstallerStore.GetInstaller(request.InstallerID)
		if err != nil {
			return deployWorkResult{}, err
		}
		if installer.OSFamily != "any" && !osFamilyMatches(target.OS, installer.OSFamily) {
			return deployWorkResult{}, stdErrors.New("installer os does not match target os")
		}
		installRequest.BinaryURL = installer.URL
		installRequest.PackageType = deploy.PackageType(installer.PackageType)
		installRequest.Checksum = installer.Checksum
	}

	result, execErr := engine.Execute(ctx, targetAddress(target), target.OS, installRequest, credentials)

	status := models.TaskStatusSuccess
	errorCode := ""
	errorMessage := ""
	remediation := ""

	if execErr != nil {
		status = models.TaskStatusFailed
		if result.ErrorDetail != nil {
			errorCode = string(result.ErrorDetail.Code)
			errorMessage = result.ErrorDetail.Message
			remediation = result.ErrorDetail.Remediation
		} else {
			errorCode = string(errors.CodeInstallFailed)
			errorMessage = execErr.Error()
			remediation = errors.RemediationFor(errors.CodeInstallFailed)
		}
	}

	_, recordErr := api.DeploymentStore.CreateDeploymentResult(store.CreateDeploymentResultInput{
		TaskRunID:    request.TaskRunID,
		TargetID:     target.ID,
		Status:       status,
		AuthMethod:   string(result.Method),
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Remediation:  remediation,
	})
	if recordErr != nil {
		return deployWorkResult{}, recordErr
	}

	return deployWorkResult{
		TargetID: target.ID,
		Method:   result.Method,
		Report:   result.Report,
		ErrorDetail:   result.ErrorDetail,
	}, execErr
}

func targetAddress(target models.Target) string {
	if target.IPAddress != "" {
		return target.IPAddress
	}
	return target.Hostname
}

