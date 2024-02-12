package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

type CreateTagInput struct {
	BoardID uint   `json:"boardId" validate:"required"`
	Name    string `json:"name"`
	Color   string `json:"color" validate:"omitempty,color"`
}

func GetCreateTagInput(c *fiber.Ctx) (models.Tag, bool) {
	var input CreateTagInput
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

type UpdateTagInput struct {
	Name  string `json:"name"`
	Color string `json:"color" validate:"omitempty,color"`
}

func GetUpdateTagInput(c *fiber.Ctx, tagId uint) (models.Tag, bool) {
	var input UpdateTagInput
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

	var tag models.Tag
	if err := store.Database.Model(&models.Tag{}).Where("id = ?", tagId).First(&tag).Error; err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Tag{}, false
	}

	if tag.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Tag{}, false
	}

	if len(input.Name) != 0 {
		tag.Name = input.Name
	}

	if len(input.Color) != 0 {
		tag.Color = input.Color
	}

	return tag, true
}
