package api

import (
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func tags(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getTags(c)
	case "PUT":
		return getTagsInBoards(c)
	case "POST":
		return createTag(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getTags(c *fiber.Ctx) error {
	var tags []models.Tag
	store.Database.Find(&tags)

	return c.Status(fiber.StatusOK).JSON(tags)
}

func getTagsInBoards(c *fiber.Ctx) error {
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

	var tags []models.Tag
	store.Database.Where("board_id IN ?", params.BoardIds).Find(&tags)

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
