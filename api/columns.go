package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetColumns(c *fiber.Ctx) error {
	tx := store.Database.Model(&models.Column{})

	var columns []models.Column

	if boardIdsQuery := c.Query("boardIds"); len(boardIdsQuery) != 0 {
		boardIds, ok := getQueryUIntArray(c, "boardIds")
		if !ok {
			return nil
		}

		tx = tx.Where("board_id IN ?", boardIds)
	}

	userBoardIds, ok := getUserBoardIds(c)
	if !ok {
		return nil
	}
	if tx.Where("board_id IN ?", userBoardIds).Find(&columns).Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeColumns(models.SortColumns(&columns)))
}

func CreateColumn(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	column, ok := schema.GetUpsertColumnInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, column.BoardID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Create(&column).Error); !ok {
		return nil
	}

	var previous models.Column
	if ok := store.Execute(c, tx, tx.Where("next_id IS NULL AND board_id = ? AND id != ?", column.BoardID, column.ID).First(&previous).Error); !ok {
		return nil
	}
	if previous.ID != 0 {
		if ok := store.Execute(c, tx, tx.Model(&previous).Update("next_id", &column.ID).Error); !ok {
			return nil
		}
	}

	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeColumn(&column))
}

func UpdateColumn(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	column, ok := schema.GetUpsertColumnInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserColumn(c, column.ID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&column).Omit("NextID", "BoardID").Updates(&column).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeColumn(&column))
}

func MoveColumn(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	columnId, ok := getParamInt(c, "column_id")
	if !ok {
		return nil
	}

	nextId := c.QueryInt("nextId")

	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&models.Column{}).Where("next_id = ?", column.ID).Update("next_id", column.NextID).Error); !ok {
		return nil
	}
	if nextId == 0 {
		if ok := store.Execute(c, tx, tx.Model(&models.Column{}).Where("next_id IS NULL AND board_id = ?", column.BoardID).Update("next_id", &column.ID).Error); !ok {
			return nil
		}
		if ok := store.Execute(c, tx, tx.Model(&column).Update("next_id", nil).Error); !ok {
			return nil
		}
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

		if ok := store.Execute(c, tx, tx.Model(&models.Column{}).Where("next_id = ?", nextId).Update("next_id", &column.ID).Error); !ok {
			return nil
		}
		if ok := store.Execute(c, tx, tx.Model(&column).Update("next_id", nextId).Error); !ok {
			return nil
		}
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func DeleteColumn(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	columnId, ok := getParamInt(c, "column_id")
	if !ok {
		return nil
	}

	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}

	var previous models.Column
	if ok := store.Execute(c, tx, tx.Where("next_id = ?", columnId).First(&previous).Error); !ok {
		return nil
	}
	if previous.ID != 0 {
		if ok := store.Execute(c, tx, tx.Where(&previous).Update("next_id", column.NextID).Error); !ok {
			return nil
		}
	}

	if ok := store.Execute(c, tx, tx.Unscoped().Delete(&column).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
