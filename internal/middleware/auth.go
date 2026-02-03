package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	LocalRoleKey  = "role"
	LocalActorKey = "actor"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

type AuthConfig struct {
	AdminKey  string
	ViewerKey string
}

func AuthMiddleware(config AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/uploads/") {
			return c.Next()
		}

		apiKey := strings.TrimSpace(c.Get("X-API-Key"))
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing api key"})
		}

		role := Role("")
		if apiKey == config.AdminKey {
			role = RoleAdmin
		} else if config.ViewerKey != "" && apiKey == config.ViewerKey {
			role = RoleViewer
		}

		if role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid api key"})
		}

		if role == RoleViewer && c.Method() != fiber.MethodGet {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
		}

		c.Locals(LocalRoleKey, string(role))
		c.Locals(LocalActorKey, apiKey)

		return c.Next()
	}
}
