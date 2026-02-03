package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

type createTargetRequest struct {
	Hostname  string          `json:"hostname"`
	IPAddress string          `json:"ipAddress"`
	OS        models.TargetOS `json:"os"`
}

type recordTargetScanRequest struct {
	Reachable bool  `json:"reachable"`
	OpenPorts []int `json:"openPorts"`
}

func (api *API) handleCreateTarget(c *fiber.Ctx) error {
	var request createTargetRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.Hostname == "" && request.IPAddress == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "hostname or ipAddress is required"})
	}

	target, err := api.TargetStore.CreateTarget(store.CreateTargetInput{
		Hostname:  request.Hostname,
		IPAddress: request.IPAddress,
		OS:        request.OS,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(target)
}

func (api *API) handleListTargets(c *fiber.Ctx) error {
	targets, err := api.TargetStore.ListTargets(parseListOptions(c))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(targets)
}

func (api *API) handleRecordTargetScan(c *fiber.Ctx) error {
	targetID := c.Params("targetId")
	var request recordTargetScanRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	scan, err := api.TargetStore.RecordTargetScan(store.TargetScanInput{
		TargetID:  targetID,
		Reachable: request.Reachable,
		OpenPorts: request.OpenPorts,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(scan)
}
