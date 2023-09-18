package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func columns(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getColumns(c)
	case "POST":
		return createColumn(c)
	case "PUT":
		return updateColumn(c)
	case "PATCH":
		return moveColumn(c)
	case "DELETE":
		return deleteColumn(c)
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
			boardId, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			boardIds = append(boardIds, uint(boardId))
		}

		query = query.Where("board_id IN ?", boardIds)
	}

	userBoardIds, ok := getUserBoardIds(c)
	if !ok {
		return nil
	}
	query.Where("board_id IN ?", userBoardIds).Find(&columns)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeColumns(&columns))
}

func createColumn(c *fiber.Ctx) error {
	column, ok := schema.GetUpsertColumnInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, column.BoardID); !ok {
		return nil
	}

	result := store.Database.Create(&column)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	var previous models.Column
	store.Database.Where("next_id IS NULL AND board_id = ? AND id != ?", column.BoardID, column.ID).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &column.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeColumn(&column))
}

func updateColumn(c *fiber.Ctx) error {
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

	_, ok := getUserColumn(c, column.ID)
	if !ok {
		return nil
	}

	store.Database.Model(&column).Omit("NextID", "BoardID").Updates(&column)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeColumn(&column))
}

func moveColumn(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) < 5 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	if paths[3] != "move" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	nextId := c.QueryInt("nextId")
	columnId, err := strconv.ParseUint(paths[4], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}

	store.Database.Model(&models.Column{}).Where("next_id = ?", column.ID).Update("next_id", column.NextID)
	if nextId <= 0 {
		store.Database.Model(&models.Column{}).Where("next_id IS NULL AND board_id = ?", column.BoardID).Update("next_id", &column.ID)
		store.Database.Model(&column).Update("next_id", nil)
	} else {
		var previous models.Column
		store.Database.Find(&previous, nextId)
		if previous.ID == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "not found",
			})
		}

		store.Database.Model(&models.Column{}).Where("next_id = ?", nextId).Update("next_id", &column.ID)
		store.Database.Model(&column).Update("next_id", nextId)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func deleteColumn(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) < 4 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	columnId, err := strconv.ParseUint(paths[3], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}

	var previous models.Column
	if store.Database.Where("next_id = ?", columnId).First(&previous); previous.ID != 0 {
		store.Database.Where(&previous).Update("next_id", column.NextID)
	}
	store.Database.Unscoped().Delete(&column)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
