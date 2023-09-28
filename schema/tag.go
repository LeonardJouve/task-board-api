package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type UpsertTagInput struct {
	BoardID uint   `json:"boardId" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Color   string `json:"color" validate:"required,color"`
}

func GetUpsertTagInput(c *fiber.Ctx) (models.Tag, bool) {
	var input UpsertTagInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Tag{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Tag{}, false
	}
	return models.Tag{
		BoardID: input.BoardID,
		Name:    input.Name,
		Color:   input.Color,
	}, true
}
