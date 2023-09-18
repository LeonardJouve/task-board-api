package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetColumns(c *fiber.Ctx) error {
	var columns []models.Column
	query := store.Database

	if boardIdsQuery := c.Query("boardIds"); len(boardIdsQuery) != 0 {
		boardIds, ok := getQueryUIntArray(c, "boardIds")
		if !ok {
			return nil
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

func CreateColumn(c *fiber.Ctx) error {
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

func UpdateColumn(c *fiber.Ctx) error {
	column, ok := schema.GetUpsertColumnInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserColumn(c, column.ID); !ok {
		return nil
	}

	store.Database.Model(&column).Omit("NextID", "BoardID").Updates(&column)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeColumn(&column))
}

func MoveColumn(c *fiber.Ctx) error {
	columnId, ok := getParamInt(c, "column_id")
	if !ok {
		return nil
	}

	nextId := c.QueryInt("nextId")

	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}

	store.Database.Model(&models.Column{}).Where("next_id = ?", column.ID).Update("next_id", column.NextID)
	if nextId == 0 {
		store.Database.Model(&models.Column{}).Where("next_id IS NULL AND board_id = ?", column.BoardID).Update("next_id", &column.ID)
		store.Database.Model(&column).Update("next_id", nil)
	} else {
		next, ok := getUserColumn(c, uint(nextId))
		if !ok {
			return nil
		}
		if next.BoardID != column.BoardID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid boardId",
			})
		}

		store.Database.Model(&models.Column{}).Where("next_id = ?", nextId).Update("next_id", &column.ID)
		store.Database.Model(&column).Update("next_id", nextId)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func DeleteColumn(c *fiber.Ctx) error {
	columnId, ok := getParamInt(c, "column_id")
	if !ok {
		return nil
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
