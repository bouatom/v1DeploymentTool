package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/errors"
	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

type createTaskRequest struct {
	Name        string `json:"name"`
	TargetCount int    `json:"targetCount"`
}

type createRunRequest struct {
}

type updateRunRequest struct {
	Status models.TaskStatus `json:"status"`
}

type scanRequest struct {
	TargetCount    int `json:"targetCount"`
	TargetsScanned int `json:"targetsScanned"`
}

type errorCatalogResponse struct {
	Code        errors.Code `json:"code"`
	Message     string      `json:"message"`
	Remediation string      `json:"remediation"`
	Steps       []string    `json:"steps"`
}

func (api *API) handleCreateTask(c *fiber.Ctx) error {
	var request createTaskRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.TargetCount < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "targetCount must be non-negative"})
	}

	task, err := api.TaskStore.CreateTask(store.CreateTaskInput{
		Name:        request.Name,
		TargetCount: request.TargetCount,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(task)
}

func (api *API) handleListTasks(c *fiber.Ctx) error {
	tasks, err := api.TaskStore.ListTasks(parseListOptions(c))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tasks)
}

func (api *API) handleCreateRun(c *fiber.Ctx) error {
	taskID := c.Params("taskId")

	run, err := api.TaskStore.CreateRun(store.CreateRunInput{TaskID: taskID})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(run)
}

func (api *API) handleUpdateRun(c *fiber.Ctx) error {
	runID := c.Params("runId")
	var request updateRunRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	run, err := api.TaskStore.UpdateRun(store.UpdateRunInput{
		RunID:  runID,
		Status: request.Status,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(run)
}

func (api *API) handleListRuns(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	runs, err := api.TaskStore.ListRuns(taskID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(runs)
}

func (api *API) handleRecordScan(c *fiber.Ctx) error {
	var request scanRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	summary, err := api.TaskStore.RecordScan(store.ScanInput{
		TargetCount:    request.TargetCount,
		TargetsScanned: request.TargetsScanned,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(summary)
}

func (api *API) handleMetrics(c *fiber.Ctx) error {
	metrics, err := api.TaskStore.GetMetrics()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(metrics)
}

func (api *API) handleErrorCatalog(c *fiber.Ctx) error {
	catalog := []errorCatalogResponse{
		{Code: errors.CodeAuthDenied, Message: "Authentication denied", Remediation: errors.RemediationFor(errors.CodeAuthDenied), Steps: errors.RemediationSteps(errors.CodeAuthDenied)},
		{Code: errors.CodeAuthTimeout, Message: "Authentication timeout", Remediation: errors.RemediationFor(errors.CodeAuthTimeout), Steps: errors.RemediationSteps(errors.CodeAuthTimeout)},
		{Code: errors.CodePortClosed, Message: "Required port closed", Remediation: errors.RemediationFor(errors.CodePortClosed), Steps: errors.RemediationSteps(errors.CodePortClosed)},
		{Code: errors.CodeUnsupportedOS, Message: "Unsupported operating system", Remediation: errors.RemediationFor(errors.CodeUnsupportedOS), Steps: errors.RemediationSteps(errors.CodeUnsupportedOS)},
		{Code: errors.CodeInstallFailed, Message: "Install failed", Remediation: errors.RemediationFor(errors.CodeInstallFailed), Steps: errors.RemediationSteps(errors.CodeInstallFailed)},
		{Code: errors.CodeNetworkIssue, Message: "Network issue detected", Remediation: errors.RemediationFor(errors.CodeNetworkIssue), Steps: errors.RemediationSteps(errors.CodeNetworkIssue)},
	}

	return c.JSON(catalog)
}
