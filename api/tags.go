package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func tags(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getTags(c)
	case "POST":
		return createTag(c)
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

	query.Find(&tags)

	return c.Status(fiber.StatusOK).JSON(tags)
}

func createTag(c *fiber.Ctx) error {
	var tag models.Tag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := store.Database.Create(&tag)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
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

	var tag models.Tag
	store.Database.First(&tag, tagId)
	if tag.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	store.Database.Unscoped().Delete(&tag)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
