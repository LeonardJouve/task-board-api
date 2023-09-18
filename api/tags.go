package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetTags(c *fiber.Ctx) error {
	var tags []models.Tag
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
	query.Where("board_id IN ?", userBoardIds).Find(&tags)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeTags(&tags))
}

func CreateTag(c *fiber.Ctx) error {
	tag, ok := schema.GetUpsertTagInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, tag.BoardID); !ok {
		return nil
	}

	if err := store.Database.Create(&tag).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeTag(&tag))
}

func UpdateTag(c *fiber.Ctx) error {
	tag, ok := schema.GetUpsertTagInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserTag(c, tag.ID); !ok {
		return nil
	}

	store.Database.Model(&models.Tag{}).Where("id = ?", tag.ID).Updates(&tag)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeTag(&tag))
}

func DeleteTag(c *fiber.Ctx) error {
	tagId, ok := getParamInt(c, "tag_id")
	if !ok {
		return nil
	}

	tag, ok := getUserTag(c, uint(tagId))
	if !ok {
		return nil
	}

	store.Database.Unscoped().Delete(&tag)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
