package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/store"
)

type createCredentialRequest struct {
	Name       string              `json:"name"`
	Kind       models.CredentialKind `json:"kind"`
	Username   string              `json:"username"`
	Password   string              `json:"password"`
	PrivateKey string              `json:"privateKey"`
}

func (api *API) handleCreateCredential(c *fiber.Ctx) error {
	var request createCredentialRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if request.Kind != models.CredentialKindSSH && request.Kind != models.CredentialKindWinRM {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid credential kind"})
	}

	credential, err := api.CredentialStore.CreateCredential(store.CreateCredentialInput{
		Name:       request.Name,
		Kind:       request.Kind,
		Username:   request.Username,
		Password:   request.Password,
		PrivateKey: request.PrivateKey,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(credential)
}

func (api *API) handleListCredentials(c *fiber.Ctx) error {
	credentials, err := api.CredentialStore.ListCredentials()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(credentials)
}
