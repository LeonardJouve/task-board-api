package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Router(c *fiber.Ctx) error {
	switch strings.Split(c.Path(), "/")[1] {
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}
