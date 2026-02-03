package main

import (
	"context"
	"log"
	"os"

	"v1-sg-deployment-tool/internal/db"
	"v1-sg-deployment-tool/internal/store/postgres"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	oldKey := os.Getenv("OLD_CREDENTIALS_KEY")
	newKey := os.Getenv("NEW_CREDENTIALS_KEY")
	newKeyID := os.Getenv("NEW_CREDENTIALS_KEY_ID")

	if databaseURL == "" || oldKey == "" || newKey == "" || newKeyID == "" {
		log.Fatal("DATABASE_URL, OLD_CREDENTIALS_KEY, NEW_CREDENTIALS_KEY, NEW_CREDENTIALS_KEY_ID are required")
	}

	pool, err := db.NewPool(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := postgres.RotateCredentials(context.Background(), pool, oldKey, newKey, newKeyID); err != nil {
		log.Fatal(err)
	}
}
