package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/deploy"
	"v1-sg-deployment-tool/internal/store"
)

type uploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Checksum string `json:"checksum"`
	InstallerID string `json:"installerId"`
	PackageType string `json:"packageType"`
	OSFamily string `json:"osFamily"`
}

func (api *API) handleUploadInstaller(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	if file.Size == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is empty"})
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	safeName := sanitizeFilename(file.Filename)
	storedName := time.Now().UTC().Format("20060102-150405") + "-" + safeName
	targetPath := filepath.Join("uploads", storedName)

	if err := c.SaveFile(file, targetPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	checksum, err := fileChecksum(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	packageType, osFamily := detectInstallerMetadata(targetPath, safeName)
	installer, err := api.InstallerStore.CreateInstaller(store.CreateInstallerInput{
		Filename:    storedName,
		URL:         c.BaseURL() + "/uploads/" + storedName,
		PackageType: string(packageType),
		OSFamily:    osFamily,
		Checksum:    checksum,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	url := c.BaseURL() + "/uploads/" + storedName
	return c.JSON(uploadResponse{
		URL:      url,
		Filename: storedName,
		Checksum: checksum,
		InstallerID: installer.ID,
		PackageType: string(packageType),
		OSFamily: osFamily,
	})
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	clean := strings.ReplaceAll(base, " ", "-")
	clean = strings.ReplaceAll(clean, "..", "")
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, "\\", "")
	if clean == "" {
		return "installer.bin"
	}
	ext := filepath.Ext(clean)
	if ext == "" {
		return clean
	}

	for _, r := range ext {
		if r == '.' || r == '-' || r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return strings.TrimSuffix(clean, ext)
	}

	return clean
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func detectInstallerMetadata(path string, filename string) (deploy.PackageType, string) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".msi" {
		return deploy.PackageTypeMSI, "windows"
	}
	if ext == ".exe" {
		return deploy.PackageTypeEXE, "windows"
	}
	if ext == ".pkg" {
		return deploy.PackageTypePKG, "macos"
	}
	if ext == ".deb" {
		return deploy.PackageTypeDEB, "linux"
	}
	if ext == ".rpm" {
		return deploy.PackageTypeRPM, "linux"
	}

	file, err := os.Open(path)
	if err != nil {
		return deploy.PackageTypeBinary, "any"
	}
	defer file.Close()

	header := make([]byte, 8)
	_, _ = io.ReadFull(file, header)

	if len(header) >= 2 && header[0] == 0x4d && header[1] == 0x5a {
		return deploy.PackageTypeEXE, "windows"
	}
	if len(header) >= 8 && header[0] == 0xd0 && header[1] == 0xcf && header[2] == 0x11 && header[3] == 0xe0 && header[4] == 0xa1 && header[5] == 0xb1 && header[6] == 0x1a && header[7] == 0xe1 {
		return deploy.PackageTypeMSI, "windows"
	}
	if len(header) >= 4 && header[0] == 0xed && header[1] == 0xab && header[2] == 0xee && header[3] == 0xdb {
		return deploy.PackageTypeRPM, "linux"
	}
	if len(header) >= 8 && string(header) == "!<arch>\n" {
		return deploy.PackageTypeDEB, "linux"
	}
	if len(header) >= 4 && string(header[:4]) == "xar!" {
		return deploy.PackageTypePKG, "macos"
	}

	return deploy.PackageTypeBinary, "any"
}
