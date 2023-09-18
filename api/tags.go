package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func tags(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getTags(c)
	case "POST":
		return createTag(c)
	case "PUT":
		return updateTag(c)
	case "DELETE":
		return deleteTag(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getTags(c *fiber.Ctx) error {
	var tags []models.Tag
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
	query.Where("board_id IN ?", userBoardIds).Find(&tags)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeTags(&tags))
}

func createTag(c *fiber.Ctx) error {
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

func updateTag(c *fiber.Ctx) error {
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

func deleteTag(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) < 4 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	tagId, err := strconv.ParseUint(paths[3], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
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
