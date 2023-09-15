package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Router(c *fiber.Ctx) error {
	switch strings.Split(c.Path(), "/")[2] {
	case "register":
		return register(c)
	case "login":
		return login(c)
	case "logout":
		return logout(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func register(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": "not implemented",
	})
}

func login(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": "not implemented",
	})
}

func logout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": "not implemented",
	})
}
