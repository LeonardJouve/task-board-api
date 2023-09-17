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
	case "PUT":
		return updateCard(c)
	case "PATCH":
		return moveCard(c)
	case "DELETE":
		return deleteCard(c)
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
	store.Database.Where("next_id IS NULL AND column_id = ? AND id != ?", card.ColumnID, card.ID).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &card.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(card)
}

func updateCard(c *fiber.Ctx) error {
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

	var currentCard models.Card
	store.Database.First(&currentCard, card.ID)
	if currentCard.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	store.Database.Model(&models.Card{}).Where("id = ?", card.ID).Omit("NextID").Updates(&card)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func moveCard(c *fiber.Ctx) error {
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
	cardId, err := strconv.ParseUint(paths[4], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	var card models.Card
	store.Database.First(&card, cardId)
	if card.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	store.Database.Model(&models.Card{}).Where("next_id = ?", card.ID).Update("next_id", card.NextID)
	if nextId <= 0 {
		store.Database.Model(&models.Card{}).Where("next_id IS NULL").Update("next_id", &card.ID)
		store.Database.Model(&card).Update("next_id", nil)
	} else {
		var previous models.Card
		store.Database.Find(&previous, nextId)
		if previous.ID == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "not found",
			})
		}

		store.Database.Model(&models.Card{}).Where("next_id = ?", nextId).Update("next_id", &card.ID)
		store.Database.Model(&card).Update("next_id", nextId)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func deleteCard(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) < 4 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	cardId, err := strconv.ParseUint(paths[3], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var card models.Card
	store.Database.First(&card, cardId)
	if card.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	var previous models.Card
	if store.Database.Where("next_id = ?", cardId).First(&previous); previous.ID != 0 {
		previous.NextID = card.NextID
		store.Database.Save(&previous)
	}
	store.Database.Unscoped().Delete(&card)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
