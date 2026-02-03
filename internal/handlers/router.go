package handlers

import (
	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/queue"
	"v1-sg-deployment-tool/internal/store"
)

type API struct {
	TaskStore   store.TaskStore
	TargetStore store.TargetStore
	AssessmentStore store.AssessmentStore
	DeploymentStore store.DeploymentStore
	CredentialStore store.CredentialStore
	InstallerStore store.InstallerStore
	Queue *queue.Queue
}

func RegisterRoutes(app *fiber.App, api *API) {
	app.Post("/api/tasks", api.handleCreateTask)
	app.Get("/api/tasks", api.handleListTasks)
	app.Get("/api/tasks/:taskId/deployments", api.handleListTaskDeployments)
	app.Get("/api/tasks/:taskId/exports/csv", api.handleExportTaskCSV)
	app.Get("/api/tasks/:taskId/exports/pdf", api.handleExportTaskPDF)
	app.Post("/api/tasks/:taskId/runs", api.handleCreateRun)
	app.Get("/api/tasks/:taskId/runs", api.handleListRuns)
	app.Patch("/api/runs/:runId", api.handleUpdateRun)
	app.Post("/api/scans", api.handleRecordScan)
	app.Post("/api/scans/execute", api.handleExecuteScan)
	app.Post("/api/scans/execute-async", api.handleExecuteScanAsync)
	app.Post("/api/uploads/installer", api.handleUploadInstaller)
	app.Get("/api/metrics", api.handleMetrics)
	app.Get("/api/errors", api.handleErrorCatalog)
	app.Post("/api/targets", api.handleCreateTarget)
	app.Get("/api/targets", api.handleListTargets)
	app.Post("/api/targets/:targetId/scans", api.handleRecordTargetScan)
	app.Post("/api/deploy/plan", api.handleBuildDeployPlan)
	app.Post("/api/deploy/execute", api.handleExecuteDeploy)
	app.Post("/api/deploy/execute-async", api.handleExecuteDeployAsync)
	app.Post("/api/preflight", api.handlePreflight)
	app.Get("/api/targets/:targetId/deployments", api.handleListDeployments)
	app.Get("/api/assessments", api.handleListAssessments)
	app.Post("/api/credentials", api.handleCreateCredential)
	app.Get("/api/credentials", api.handleListCredentials)
	app.Get("/api/jobs/:jobId", api.handleGetJob)
}
