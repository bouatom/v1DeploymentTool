package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"v1-sg-deployment-tool/internal/store"
)

func parseListOptions(c *fiber.Ctx) store.ListOptions {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 500 {
		limit = 500
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return store.ListOptions{
		Limit:  limit,
		Offset: offset,
	}
}
