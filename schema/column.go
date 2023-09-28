package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type UpsertColumnInput struct {
	Name    string `json:"name" validate:"required"`
	BoardID uint   `json:"boardId" validate:"required"`
}

func GetUpsertColumnInput(c *fiber.Ctx) (models.Column, bool) {
	var input UpsertColumnInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Column{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Column{}, false
	}

	return models.Column{
		Name:    input.Name,
		BoardID: input.BoardID,
	}, true
}
