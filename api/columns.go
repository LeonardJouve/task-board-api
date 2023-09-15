package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func columns(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getColumns(c)
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
	query := store.Database

	if len(c.Query("boardIds")) != 0 {
		var boardIds []uint
		for _, id := range strings.Split(c.Query("boardIds"), ",") {
			columnId, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			boardIds = append(boardIds, uint(columnId))
		}

		query = query.Where("board_id IN ?", boardIds)
	}

	query.Find(&columns)

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
