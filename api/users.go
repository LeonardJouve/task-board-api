package api

import (
	"strings"

	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func users(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		if strings.Split(c.Path(), "/")[3] != "me" {
			break
		}
		return getMe(c)
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "not found",
	})
}

func getMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(user.Sanitize())
}
