package api

// TODO
// - Websocket

import (
	"strings"

	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/gofiber/fiber/v2"
)

var validate = schema.Init()

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
