package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func (store *Store) CreateInstaller(input store.CreateInstallerInput) (models.Installer, error) {
	if input.Filename == "" || input.URL == "" {
		return models.Installer{}, errors.New("filename and url are required")
	}
	if input.PackageType == "" || input.OSFamily == "" {
		return models.Installer{}, errors.New("package type and os family are required")
	}
	if input.Checksum == "" {
		return models.Installer{}, errors.New("checksum is required")
	}

	now := time.Now().UTC()
	installerID := generateID()

	_, err := store.pool.Exec(context.Background(), `
		INSERT INTO installers (id, filename, url, package_type, os_family, checksum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, installerID, input.Filename, input.URL, input.PackageType, input.OSFamily, input.Checksum, now)
	if err != nil {
		return models.Installer{}, err
	}

	return models.Installer{
		ID:          installerID,
		Filename:    input.Filename,
		URL:         input.URL,
		PackageType: input.PackageType,
		OSFamily:    input.OSFamily,
		Checksum:    input.Checksum,
		CreatedAt:   now,
	}, nil
}

func (store *Store) GetInstaller(installerID string) (models.Installer, error) {
	if installerID == "" {
		return models.Installer{}, errors.New("installer id is required")
	}

	var installer models.Installer
	err := store.pool.QueryRow(context.Background(), `
		SELECT id, filename, url, package_type, os_family, checksum, created_at
		FROM installers
		WHERE id = $1
	`, installerID).Scan(
		&installer.ID,
		&installer.Filename,
		&installer.URL,
		&installer.PackageType,
		&installer.OSFamily,
		&installer.Checksum,
		&installer.CreatedAt,
	)
	if err != nil {
		return models.Installer{}, err
	}

	return installer, nil
}
