package api

import (
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func boards(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getBoards(c)
	case "POST":
		return createBoard(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getBoards(c *fiber.Ctx) error {
	var boards []models.Board
	store.Database.Find(&boards)

	return c.Status(fiber.StatusOK).JSON(boards)
}

func createBoard(c *fiber.Ctx) error {
	var board models.Board

	if err := c.BodyParser(&board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := store.Database.Create(&board)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	var previous models.Board
	store.Database.Model(models.Board{NextID: nil}).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &board.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(board)
}
