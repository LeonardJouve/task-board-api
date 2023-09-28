package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type UpsertCardInput struct {
	ColumnID uint   `json:"columnId" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
}

func GetUpsertCardInput(c *fiber.Ctx) (models.Card, bool) {
	var input UpsertCardInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Card{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Card{}, false
	}

	return models.Card{
		ColumnID: input.ColumnID,
		Name:     input.Name,
		Content:  input.Content,
	}, true
}
