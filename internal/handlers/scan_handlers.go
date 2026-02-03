package handlers

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/osdetect"
	"v1-sg-deployment-tool/internal/scanner"
	"v1-sg-deployment-tool/internal/store"
	"v1-sg-deployment-tool/internal/targets"
)

type executeScanRequest struct {
	Targets        []string `json:"targets"`
	Aggressiveness int      `json:"aggressiveness"`
}

type executeScanResponse struct {
	TargetCount    int `json:"targetCount"`
	TargetsScanned int `json:"targetsScanned"`
	Errors         []string `json:"errors"`
}

func (api *API) handleExecuteScan(c *fiber.Ctx) error {
	var request executeScanRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.Aggressiveness < 1 || request.Aggressiveness > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "aggressiveness must be between 1 and 5"})
	}

	results, parseErrors, err := api.executeScanWork(request)
	if err != nil {
		status := fiber.StatusInternalServerError
		if len(parseErrors) > 0 {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error(), "details": parseErrorMessages(parseErrors)})
	}

	return c.JSON(executeScanResponse{
		TargetCount:    len(results),
		TargetsScanned: len(results),
		Errors:         parseErrorMessages(parseErrors),
	})
}

func (api *API) handleExecuteScanAsync(c *fiber.Ctx) error {
	if api.Queue == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "queue not available"})
	}

	var request executeScanRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.Aggressiveness < 1 || request.Aggressiveness > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "aggressiveness must be between 1 and 5"})
	}

	job, err := api.Queue.EnqueueWithHandler("scan", func(ctx context.Context) error {
		_, _, err := api.executeScanWork(request)
		return err
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(job)
}

func (api *API) executeScanWork(request executeScanRequest) ([]scanner.ScanResult, []error, error) {
	specs, parseErrors := targets.ParseInputs(request.Targets)
	if len(parseErrors) > 0 {
		return nil, parseErrors, errors.New("invalid targets")
	}

	config := buildScannerConfig(request.Aggressiveness)
	probe := scanner.PortProbe{
		Ports:   []int{22, 5985, 5986},
		Timeout: config.Timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	results, err := scanner.ScanTargets(ctx, specs, config, probe)
	if err != nil {
		return nil, parseErrors, err
	}

	for _, result := range results {
		targetID, err := api.persistTarget(result)
		if err != nil {
			return nil, parseErrors, err
		}

		_, err = api.TargetStore.RecordTargetScan(store.TargetScanInput{
			TargetID:  targetID,
			Reachable: result.Reachable,
			OpenPorts: result.OpenPorts,
		})
		if err != nil {
			return nil, parseErrors, err
		}
	}

	if _, err := api.TaskStore.RecordScan(store.ScanInput{
		TargetCount:    len(results),
		TargetsScanned: len(results),
	}); err != nil {
		return nil, parseErrors, err
	}

	return results, parseErrors, nil
}

func (api *API) persistTarget(result scanner.ScanResult) (string, error) {
	os := osdetect.DetectFromPorts(result.OpenPorts)
	hostname := ""
	ipAddress := ""
	if parsed := net.ParseIP(result.Host); parsed != nil {
		ipAddress = result.Host
	} else {
		hostname = result.Host
	}

	target, err := api.TargetStore.CreateTarget(store.CreateTargetInput{
		Hostname:  hostname,
		IPAddress: ipAddress,
		OS:        os,
	})
	if err != nil {
		return "", err
	}

	return target.ID, nil
}

func buildScannerConfig(aggressiveness int) scanner.ScannerConfig {
	level := aggressiveness
	if level < 1 {
		level = 1
	}
	if level > 5 {
		level = 5
	}

	const baseConcurrency = 4
	const baseRate = 20
	const baseTimeout = 4 * time.Second

	return scanner.ScannerConfig{
		MaxConcurrency: baseConcurrency * level,
		RatePerSecond:  baseRate * level,
		Timeout:        baseTimeout,
	}
}

func parseErrorMessages(errs []error) []string {
	if len(errs) == 0 {
		return nil
	}

	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return messages
}
