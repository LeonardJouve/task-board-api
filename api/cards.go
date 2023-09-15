package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func cards(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		return getCards(c)
	case "POST":
		return createCard(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getCards(c *fiber.Ctx) error {
	var cards []models.Card
	query := store.Database

	if len(c.Query("columnIds")) != 0 {
		var columnIds []uint
		for _, id := range strings.Split(c.Query("columnIds"), ",") {
			columnId, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			columnIds = append(columnIds, uint(columnId))
		}

		query = query.Where("column_id IN ?", columnIds)
	}

	query.Find(&cards)

	return c.Status(fiber.StatusOK).JSON(cards)
}

func createCard(c *fiber.Ctx) error {
	var card models.Card

	if err := c.BodyParser(&card); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(card); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := store.Database.Create(&card)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	var previous models.Card
	store.Database.Model(models.Card{NextID: nil}).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &card.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(card)
}
