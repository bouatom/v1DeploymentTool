package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/store"
)

func AuditMiddleware(auditStore store.AuditStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if auditStore == nil {
			return err
		}

		role, _ := c.Locals(LocalRoleKey).(string)
		actor, _ := c.Locals(LocalActorKey).(string)
		statusCode := c.Response().StatusCode()

		_ = auditStore.RecordAudit(store.AuditInput{
			Actor:      actor,
			Role:       role,
			Action:     c.Method(),
			Path:       c.Path(),
			StatusCode: statusCode,
			CreatedAt:  time.Now().UTC(),
		})

		return err
	}
}
