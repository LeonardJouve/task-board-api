package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetCards(c *fiber.Ctx) error {
	var cards []models.Card
	query := store.Database

	if len(c.Query("columnIds")) != 0 {
		columnIds, ok := getQueryUIntArray(c, "columnIds")
		if !ok {
			return nil
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

func AddTag(c *fiber.Ctx) error {
	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
	}

	tagId := c.QueryInt("tagId")

	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}

	column, ok := getUserColumn(c, card.ColumnID)
	if !ok {
		return nil
	}

	tag, ok := getUserTag(c, uint(tagId))
	if !ok {
		return nil
	}

	if column.BoardID != tag.BoardID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid tagId",
		})
	}

	store.Database.Model(&card).Association("Tags").Append([]models.Tag{tag})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func CreateCard(c *fiber.Ctx) error {
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
		store.Database.Model(&previous).Update("next_id", &card.ID)
	}

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeCard(&card))
}

func UpdateCard(c *fiber.Ctx) error {
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

func MoveCard(c *fiber.Ctx) error {
	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
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

func DeleteCard(c *fiber.Ctx) error {
	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
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
