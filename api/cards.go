package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetCards(c *fiber.Ctx) error {
	tx := store.Database.Model(&models.Card{})

	var cards []models.Card

	if len(c.Query("columnIds")) != 0 {
		columnIds, ok := getQueryUIntArray(c, "columnIds")
		if !ok {
			return nil
		}

		tx = tx.Where("column_id IN ?", columnIds)
	}

	userColumnIds, ok := getUserColumnIds(c)
	if !ok {
		return nil
	}
	if tx.Where("column_id IN ?", userColumnIds).Preload("Tags").Find(&cards).Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeCards(models.SortCards(&cards)))
}

func GetCard(c *fiber.Ctx) error {
	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
	}

	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeCard(&card))
}

func AddTag(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

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

	if ok := store.Execute(c, tx, tx.Model(&card).Association("Tags").Append([]models.Tag{tag})); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(models.SanitizeCard(&card))
}

func CreateCard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

	card, ok := schema.GetCreateCardInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserColumn(c, card.ColumnID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Create(&card).Error); !ok {
		return nil
	}

	var previous models.Card
	if ok := store.Execute(c, tx, tx.Where("next_id IS NULL AND column_id = ?", card.ColumnID).First(&previous).Error); !ok {
		return nil
	}
	if previous.ID != 0 {
		if ok := store.Execute(c, tx, tx.Model(&previous).Update("next_id", &card.ID).Error); !ok {
			return nil
		}
	}

	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(models.SanitizeCard(&card))
}

func UpdateCard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
	}

	card, ok := schema.GetUpdateCardInput(c, uint(cardId))
	if !ok {
		return nil
	}

	if _, ok := getUserCard(c, card.ID); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&models.Card{}).Where("id = ?", card.ID).Omit("NextID", "ColumnID").Preload("Tags").Updates(&card).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(models.SanitizeCard(&card))
}

func MoveCard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

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

	if ok := store.Execute(c, tx, tx.Model(&card).Preload("Column").Find(&card).Error); !ok {
		return nil
	}

	if card.Column.BoardID != column.BoardID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid next id",
		})
	}

	if ok := store.Execute(c, tx, tx.Model(&models.Card{}).Where("next_id = ?", card.ID).Update("next_id", card.NextID).Error); !ok {
		return nil
	}
	if nextId == 0 {
		if ok := store.Execute(c, tx, tx.Model(&models.Card{}).Where("next_id IS NULL AND column_id = ?", column.ID).Update("next_id", &card.ID).Error); !ok {
			return nil
		}
		if ok := store.Execute(c, tx, tx.Model(&card).Updates(map[string]interface{}{
			"next_id":   nil,
			"column_id": column.ID,
		}).Error); !ok {
			return nil
		}
	} else {
		next, ok := getUserCard(c, uint(nextId))
		if !ok {
			return nil
		}

		if next.ID == card.ID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "card id must be different from next id",
			})
		}

		if next.ColumnID != column.ID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid columnId",
			})
		}

		if ok := store.Execute(c, tx, tx.Model(&models.Card{}).Where("next_id = ?", next.ID).Update("next_id", &card.ID).Error); !ok {
			return nil
		}
		if ok := store.Execute(c, tx, tx.Model(&card).Updates(&models.Card{
			NextID:   &next.ID,
			ColumnID: column.ID,
		}).Error); !ok {
			return nil
		}
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(models.SanitizeCard(&card))
}

func DeleteCard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer store.RollbackTransactionIfNeeded(c, tx)

	cardId, ok := getParamInt(c, "card_id")
	if !ok {
		return nil
	}

	card, ok := getUserCard(c, uint(cardId))
	if !ok {
		return nil
	}

	var previous models.Card
	if ok := store.Execute(c, tx, tx.Where("next_id = ?", cardId).First(&previous).Error); !ok {
		return nil
	}
	if previous.ID != 0 {
		if ok := store.Execute(c, tx, tx.Model(&previous).Update("next_id", card.NextID).Error); !ok {
			return nil
		}
	}

	if ok := store.Execute(c, tx, tx.Unscoped().Delete(&card).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
