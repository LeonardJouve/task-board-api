package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

type CreateColumnInput struct {
	BoardID uint   `json:"boardId" validate:"required"`
	Name    string `json:"name"`
}

func GetCreateColumnInput(c *fiber.Ctx) (models.Column, bool) {
	var input CreateColumnInput
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

type UpdateColumnInput struct {
	Name string `json:"name"`
}

func GetUpdateColumnInput(c *fiber.Ctx, columnId uint) (models.Column, bool) {
	var input UpdateColumnInput
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

	var column models.Column
	if err := store.Database.Model(&models.Column{}).Where("id = ?", columnId).First(&column).Error; err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Column{}, false
	}

	if column.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Column{}, false
	}

	if len(input.Name) != 0 {
		column.Name = input.Name
	}

	return column, true
}
