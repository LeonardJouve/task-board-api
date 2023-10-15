package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

type CreateBoardInput struct {
	Name string `json:"name" validate:"required"`
}

func GetCreateBoardInput(c *fiber.Ctx) (models.Board, bool) {
	var input CreateBoardInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}

	return models.Board{
		Name: input.Name,
	}, true
}

type UpdateBoardInput struct {
	Name string `json:"name"`
}

func GetUpdateBoardInput(c *fiber.Ctx, boardId uint) (models.Board, bool) {
	var input UpdateBoardInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}

	var board models.Board
	if err := store.Database.Model(&models.Board{}).Where("id = ?", boardId).First(&board).Error; err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}

	if board.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Board{}, false
	}

	if len(input.Name) != 0 {
		board.Name = input.Name
	}

	return board, true
}
