package store

import "v1-sg-deployment-tool/internal/models"

type CredentialStore interface {
	CreateCredential(input CreateCredentialInput) (models.Credential, error)
	ListCredentials() ([]models.Credential, error)
	GetCredential(id string) (models.Credential, error)
}

type CreateCredentialInput struct {
	Name       string
	Kind       models.CredentialKind
	Username   string
	Password   string
	PrivateKey string
}
