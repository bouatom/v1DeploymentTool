package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (api *API) handleGetJob(c *fiber.Ctx) error {
	if api.Queue == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "queue not available"})
	}

	jobID := c.Params("jobId")
	job, err := api.Queue.GetJob(jobID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(job)
}
