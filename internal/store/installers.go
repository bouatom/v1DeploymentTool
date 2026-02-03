package store

import "v1-sg-deployment-tool/internal/models"

type InstallerStore interface {
	CreateInstaller(input CreateInstallerInput) (models.Installer, error)
	GetInstaller(installerID string) (models.Installer, error)
}

type CreateInstallerInput struct {
	Filename    string
	URL         string
	PackageType string
	OSFamily    string
	Checksum    string
}
