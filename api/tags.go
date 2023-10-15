package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetTags(c *fiber.Ctx) error {
	tx := store.Database.Model(&models.Tag{})

	var tags []models.Tag

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
	if tx.Where("board_id IN ?", userBoardIds).Find(&tags).Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeTags(&tags))
}

func GetTag(c *fiber.Ctx) error {
	tagId, ok := getParamInt(c, "tag_id")
	if !ok {
		return nil
	}

	tag, ok := getUserTag(c, uint(tagId))
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeTag(&tag))
}

func CreateTag(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	tag, ok := schema.GetUpsertTagInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, tag.BoardID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Create(&tag).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(models.SanitizeTag(&tag))
}

func UpdateTag(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	tag, ok := schema.GetUpsertTagInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserTag(c, tag.ID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&models.Tag{}).Where("id = ?", tag.ID).Omit("BoardID").Updates(&tag).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(models.SanitizeTag(&tag))
}

func DeleteTag(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	tagId, ok := getParamInt(c, "tag_id")
	if !ok {
		return nil
	}

	tag, ok := getUserTag(c, uint(tagId))
	if !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Unscoped().Delete(&tag).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
