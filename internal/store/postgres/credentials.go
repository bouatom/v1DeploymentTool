package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/crypto"
	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

func (store *Store) CreateCredential(input store.CreateCredentialInput) (models.Credential, error) {
	if input.Name == "" {
		return models.Credential{}, errors.New("name is required")
	}
	if input.Username == "" {
		return models.Credential{}, errors.New("username is required")
	}

	passwordEnc := ""
	privateKeyEnc := ""
	var err error

	if input.Password != "" {
		passwordEnc, err = crypto.Encrypt(store.credentialsKey, input.Password)
		if err != nil {
			return models.Credential{}, err
		}
	}

	if input.PrivateKey != "" {
		privateKeyEnc, err = crypto.Encrypt(store.credentialsKey, input.PrivateKey)
		if err != nil {
			return models.Credential{}, err
		}
	}

	now := time.Now().UTC()
	credentialID := generateID()

	_, err = store.pool.Exec(context.Background(), `
		INSERT INTO credentials (
			id, name, kind, username, password_enc, private_key_enc, key_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, credentialID, input.Name, input.Kind, input.Username, passwordEnc, privateKeyEnc, store.credentialsKeyID, now, now)
	if err != nil {
		return models.Credential{}, err
	}

	return models.Credential{
		ID:         credentialID,
		Name:       input.Name,
		Kind:       input.Kind,
		Username:   input.Username,
		KeyID:      store.credentialsKeyID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (store *Store) ListCredentials() ([]models.Credential, error) {
	rows, err := store.pool.Query(context.Background(), `
		SELECT id, name, kind, username, key_id, created_at, updated_at
		FROM credentials
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []models.Credential
	for rows.Next() {
		var credential models.Credential
		if err := rows.Scan(
			&credential.ID,
			&credential.Name,
			&credential.Kind,
			&credential.Username,
			&credential.KeyID,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		); err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}

func (store *Store) GetCredential(id string) (models.Credential, error) {
	if id == "" {
		return models.Credential{}, errors.New("credential id is required")
	}

	var credential models.Credential
	var passwordEnc string
	var privateKeyEnc string

	err := store.pool.QueryRow(context.Background(), `
		SELECT id, name, kind, username, password_enc, private_key_enc, key_id, created_at, updated_at
		FROM credentials
		WHERE id = $1
	`, id).Scan(
		&credential.ID,
		&credential.Name,
		&credential.Kind,
		&credential.Username,
		&passwordEnc,
		&privateKeyEnc,
		&credential.KeyID,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)
	if err != nil {
		return models.Credential{}, err
	}

	if passwordEnc != "" {
		decrypted, err := crypto.Decrypt(store.credentialsKey, passwordEnc)
		if err != nil {
			return models.Credential{}, err
		}
		credential.Password = decrypted
	}

	if privateKeyEnc != "" {
		decrypted, err := crypto.Decrypt(store.credentialsKey, privateKeyEnc)
		if err != nil {
			return models.Credential{}, err
		}
		credential.PrivateKey = decrypted
	}

	return credential, nil
}
