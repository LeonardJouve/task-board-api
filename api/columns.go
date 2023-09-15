package api

import (
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func columns(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getColumns(c)
	case "PUT":
		return getColumnsInBoards(c)
	case "POST":
		return createColumn(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getColumns(c *fiber.Ctx) error {
	var columns []models.Column
	store.Database.Find(&columns)

	return c.Status(fiber.StatusOK).JSON(columns)
}

func getColumnsInBoards(c *fiber.Ctx) error {
	var params struct {
		BoardIds []uint `json:"boardIds" validate:"required"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var columns []models.Column
	store.Database.Where("board_id IN ?", params.BoardIds).Find(&columns)

	return c.Status(fiber.StatusOK).JSON(columns)
}

func createColumn(c *fiber.Ctx) error {
	var column models.Column

	if err := c.BodyParser(&column); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(column); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := store.Database.Create(&column)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	var previous models.Column
	store.Database.Model(models.Column{NextID: nil}).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &column.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(column)
}
