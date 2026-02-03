package models

import "time"

type CredentialKind string

const (
	CredentialKindSSH   CredentialKind = "ssh"
	CredentialKindWinRM CredentialKind = "winrm"
)

type Credential struct {
	ID         string
	Name       string
	Kind       CredentialKind
	Username   string
	Password   string
	PrivateKey string
	KeyID      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
