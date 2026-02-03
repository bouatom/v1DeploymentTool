package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"

	"v1-sg-deployment-tool/internal/store"
)

type queryExec interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func generateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}

	return hex.EncodeToString(bytes)
}

func normalizeListOptions(options store.ListOptions) (int, int) {
	limit := options.Limit
	offset := options.Offset

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
