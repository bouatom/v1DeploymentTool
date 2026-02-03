package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"v1-sg-deployment-tool/internal/config"
	"v1-sg-deployment-tool/internal/db"
	"v1-sg-deployment-tool/internal/handlers"
	"v1-sg-deployment-tool/internal/maintenance"
	"v1-sg-deployment-tool/internal/middleware"
	"v1-sg-deployment-tool/internal/queue"
	"v1-sg-deployment-tool/internal/store/postgres"
)

func main() {
	appConfig, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.NewPool(context.Background(), appConfig.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.RunMigrations(context.Background(), pool); err != nil {
		log.Fatal(err)
	}

	apiStore, err := postgres.NewStore(pool, appConfig.CredentialsKey, appConfig.CredentialsKeyID)
	if err != nil {
		log.Fatal(err)
	}
	jobQueue := queue.NewQueue(4)

	app := buildApp(pool, apiStore, jobQueue, appConfig)
	maintenance.StartRetentionLoop(apiStore, appConfig.RetentionDays, log.Default())

	log.Fatal(app.Listen(appConfig.HTTPAddress))
}

func buildApp(pool *pgxpool.Pool, apiStore *postgres.Store, jobQueue *queue.Queue, appConfig config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(cors.New())
	app.Use(middleware.AuthMiddleware(middleware.AuthConfig{
		AdminKey:  appConfig.AdminAPIKey,
		ViewerKey: appConfig.ViewerAPIKey,
	}))
	app.Use(middleware.AuditMiddleware(apiStore))

	app.Get("/healthz", handleHealthz)
	app.Get("/readyz", handleReadyz(pool))
	app.Static("/uploads", "./uploads")

	handlers.RegisterRoutes(app, &handlers.API{
		TaskStore:   apiStore,
		TargetStore: apiStore,
		AssessmentStore: apiStore,
		DeploymentStore: apiStore,
		CredentialStore: apiStore,
		InstallerStore: apiStore,
		Queue: jobQueue,
	})

	return app
}

func handleHealthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func handleReadyz(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "db_unavailable",
			})
		}

		return c.JSON(fiber.Map{
			"status": "ready",
		})
	}
}
