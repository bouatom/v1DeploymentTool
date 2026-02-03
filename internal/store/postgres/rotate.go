package postgres

import (
	"context"
	"errors"
	"time"

	"v1-sg-deployment-tool/internal/crypto"
)

type credentialRow struct {
	ID           string
	PasswordEnc  string
	PrivateKeyEnc string
}

func RotateCredentials(ctx context.Context, pool queryExec, oldKey string, newKey string, newKeyID string) error {
	if oldKey == "" || newKey == "" {
		return errors.New("old and new keys are required")
	}
	if newKeyID == "" {
		return errors.New("new key id is required")
	}

	rows, err := pool.Query(ctx, `
		SELECT id, password_enc, private_key_enc
		FROM credentials
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var rowsToUpdate []credentialRow
	for rows.Next() {
		var row credentialRow
		if err := rows.Scan(&row.ID, &row.PasswordEnc, &row.PrivateKeyEnc); err != nil {
			return err
		}
		rowsToUpdate = append(rowsToUpdate, row)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, row := range rowsToUpdate {
		passwordEnc := ""
		privateKeyEnc := ""

		if row.PasswordEnc != "" {
			decrypted, err := crypto.Decrypt(oldKey, row.PasswordEnc)
			if err != nil {
				return err
			}
			encrypted, err := crypto.Encrypt(newKey, decrypted)
			if err != nil {
				return err
			}
			passwordEnc = encrypted
		}

		if row.PrivateKeyEnc != "" {
			decrypted, err := crypto.Decrypt(oldKey, row.PrivateKeyEnc)
			if err != nil {
				return err
			}
			encrypted, err := crypto.Encrypt(newKey, decrypted)
			if err != nil {
				return err
			}
			privateKeyEnc = encrypted
		}

		_, err := pool.Exec(ctx, `
			UPDATE credentials
			SET password_enc = $1, private_key_enc = $2, key_id = $3, updated_at = $4
			WHERE id = $5
		`, passwordEnc, privateKeyEnc, newKeyID, time.Now().UTC(), row.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
