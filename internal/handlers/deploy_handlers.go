package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/deploy"
	"v1-sg-deployment-tool/internal/models"
)

type buildPlanRequest struct {
	OS               models.TargetOS `json:"os"`
	BinaryURL        string          `json:"binaryUrl"`
	InstallerID      string          `json:"installerId"`
	DestinationPath  string          `json:"destinationPath"`
	PostInstallArgs  []string        `json:"postInstallArgs"`
	ExecuteOnInstall bool            `json:"executeOnInstall"`
	PackageType      deploy.PackageType `json:"packageType"`
	Checksum         string            `json:"checksum"`
	ChecksumAlg      string            `json:"checksumAlg"`
	ExpectedArch     string            `json:"expectedArch"`
	MinFreeMB        int               `json:"minFreeMB"`
	ProxyURL         string            `json:"proxyUrl"`
	RequiresReboot   bool              `json:"requiresReboot"`
	AllowReboot      bool              `json:"allowReboot"`
}

func (api *API) handleBuildDeployPlan(c *fiber.Ctx) error {
	var request buildPlanRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	installRequest := deploy.InstallRequest{
		OS:               request.OS,
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
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if installer.OSFamily != "any" && !osFamilyMatches(request.OS, installer.OSFamily) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "installer os does not match target os"})
		}
		installRequest.BinaryURL = installer.URL
		installRequest.PackageType = deploy.PackageType(installer.PackageType)
		installRequest.Checksum = installer.Checksum
	}

	if installRequest.BinaryURL == "" && request.InstallerID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "binaryUrl or installerId is required"})
	}

	plan, err := deploy.BuildPlan(installRequest)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(plan)
}

