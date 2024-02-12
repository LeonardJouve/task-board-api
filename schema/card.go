package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

type CreateCardInput struct {
	ColumnID uint   `json:"columnId" validate:"required"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

func GetCreateCardInput(c *fiber.Ctx) (models.Card, bool) {
	var input CreateCardInput
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

type UpdateCardInput struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func GetUpdateCardInput(c *fiber.Ctx, cardId uint) (models.Card, bool) {
	var input UpdateCardInput
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

	var card models.Card
	if err := store.Database.Model(&models.Card{}).Where("id = ?", cardId).First(&card).Error; err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Card{}, false
	}

	if card.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Card{}, false
	}

	if len(input.Name) != 0 {
		card.Name = input.Name
	}

	if len(input.Content) != 0 {
		card.Content = input.Content
	}

	return card, true
}
