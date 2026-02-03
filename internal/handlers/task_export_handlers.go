package handlers

import (
	"bytes"
	"encoding/csv"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"

	"v1-sg-deployment-tool/internal/store"
)

func (api *API) handleListTaskDeployments(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	results, err := api.DeploymentStore.ListDeploymentResultsByTask(taskID, parseListOptions(c))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if results == nil {
		results = []store.DeploymentResultDetail{}
	}

	return c.JSON(results)
}

func (api *API) handleExportTaskCSV(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	results, err := api.DeploymentStore.ListDeploymentResultsByTask(taskID, parseListOptions(c))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{
		"target_label",
		"target_os",
		"status",
		"auth_method",
		"error_code",
		"error_message",
		"remediation",
		"finished_at",
	})

	for _, result := range results {
		_ = writer.Write([]string{
			result.TargetLabel,
			string(result.TargetOS),
			string(result.Status),
			result.AuthMethod,
			result.ErrorCode,
			result.ErrorMessage,
			result.Remediation,
			result.FinishedAt,
		})
	}

	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=task-"+taskID+"-deployments.csv")
	return c.Send(buffer.Bytes())
}

func (api *API) handleExportTaskPDF(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	results, err := api.DeploymentStore.ListDeploymentResultsByTask(taskID, store.ListOptions{Limit: 500})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(40, 10, "Deployment Report")
	pdf.Ln(12)
	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(40, 8, "Task ID: "+taskID)
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(50, 8, "Target", "1", 0, "", false, 0, "")
	pdf.CellFormat(18, 8, "OS", "1", 0, "", false, 0, "")
	pdf.CellFormat(20, 8, "Status", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Auth", "1", 0, "", false, 0, "")
	pdf.CellFormat(72, 8, "Error", "1", 0, "", false, 0, "")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 9)
	for _, result := range results {
		pdf.CellFormat(50, 7, result.TargetLabel, "1", 0, "", false, 0, "")
		pdf.CellFormat(18, 7, string(result.TargetOS), "1", 0, "", false, 0, "")
		pdf.CellFormat(20, 7, string(result.Status), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 7, result.AuthMethod, "1", 0, "", false, 0, "")
		pdf.CellFormat(72, 7, result.ErrorCode, "1", 0, "", false, 0, "")
		pdf.Ln(7)
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=task-"+taskID+"-deployments.pdf")
	return c.Send(buffer.Bytes())
}
