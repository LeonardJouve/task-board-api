package api

// TODO
// - Auth
// - Filter response with allowed ressources only
// - Websocket

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func Router(c *fiber.Ctx) error {
	switch strings.Split(c.Path(), "/")[2] {
	case "boards":
		return boards(c)
	case "columns":
		return columns(c)
	case "cards":
		return cards(c)
	case "tags":
		return tags(c)
	case "users":
		return users(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}
