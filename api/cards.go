package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func cards(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		if paths := strings.Split(c.Path(), "/"); len(paths) == 5 && paths[4] == "tag" {
			return addTag(c)
		}
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

	userColumnIds, ok := getUserColumnIds(c)
	if !ok {
		return nil
	}
	query.Where("column_id IN ?", userColumnIds).Find(&cards)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeCards(&cards))
}

func addTag(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) != 5 {
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
	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}
	column, ok := getUserColumn(c, card.ColumnID)
	if !ok {
		return nil
	}
	tagId := c.QueryInt("tagId")
	tag, ok := getUserTag(c, uint(tagId))
	if !ok {
		return nil
	}

	if column.BoardID != tag.BoardID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid tag",
		})
	}

	store.Database.Model(&card).Association("Tags").Append([]models.Tag{tag})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func createCard(c *fiber.Ctx) error {
	card, ok := schema.GetUpsertCardInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserColumn(c, card.ColumnID); !ok {
		return nil
	}

	if err := store.Database.Create(&card).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var previous models.Card
	store.Database.Where("next_id IS NULL AND column_id = ? AND id != ?", card.ColumnID, card.ID).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &card.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeCard(&card))
}

func updateCard(c *fiber.Ctx) error {
	card, ok := schema.GetUpsertCardInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserCard(c, card.ID); !ok {
		return nil
	}

	store.Database.Model(&models.Card{}).Where("id = ?", card.ID).Omit("NextID", "ColumnID").Updates(&card)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeCard(&card))
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
	cardId, err := strconv.ParseUint(paths[4], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	nextId := c.QueryInt("nextId")
	columnId := c.QueryInt("columnId")
	column, ok := getUserColumn(c, uint(columnId))
	if !ok {
		return nil
	}
	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}

	if nextId == 0 {
		store.Database.Model(&models.Card{}).Where("next_id = ?", card.ID).Update("next_id", card.NextID)
		store.Database.Model(&models.Card{}).Where("next_id IS NULL AND column_id = ?", column.ID).Update("next_id", &card.ID)
		store.Database.Model(&card).Updates(&models.Card{
			NextID:   nil,
			ColumnID: column.ID,
		})
	} else {
		next, ok := getUserCard(c, uint(nextId))
		if !ok {
			return nil
		}
		if next.ColumnID != column.ID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid columnId",
			})
		}

		store.Database.Model(&models.Card{}).Where("next_id = ?", card.ID).Update("next_id", card.NextID)
		store.Database.Model(&models.Card{}).Where("next_id = ?", nextId).Update("next_id", &card.ID)
		store.Database.Model(&card).Updates(&models.Card{
			NextID:   &next.ID,
			ColumnID: column.ID,
		})
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

	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}

	var previous models.Card
	if store.Database.Where("next_id = ?", cardId).First(&previous); previous.ID != 0 {
		store.Database.Model(&previous).Update("next_id", card.NextID)
	}
	store.Database.Unscoped().Delete(&card)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
