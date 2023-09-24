package schema

import (
	"errors"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SanitizedCard struct {
	ID       uint   `json:"id"`
	ColumnID uint   `json:"columnId"`
	TagIDs   []uint `json:"tagIds"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

func SanitizeCard(card *models.Card) *SanitizedCard {
	var c models.Card
	if err := store.Database.Where(&card).Preload("Tags").First(&c).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	tagIds := []uint{}
	for _, tag := range c.Tags {
		tagIds = append(tagIds, tag.ID)
	}

	return &SanitizedCard{
		ID:       card.ID,
		ColumnID: card.ColumnID,
		TagIDs:   tagIds,
		Name:     card.Name,
		Content:  card.Content,
	}
}

func SanitizeCards(cards *[]models.Card) *[]SanitizedCard {
	sanitizedCards := []SanitizedCard{}
	for _, card := range *cards {
		sanitizedCards = append(sanitizedCards, *(SanitizeCard(&card)))
	}

	return &sanitizedCards
}

type UpsertCardInput struct {
	ColumnID uint   `json:"columnId" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
}

func GetUpsertCardInput(c *fiber.Ctx) (models.Card, bool) {
	var input UpsertCardInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Card{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Card{}, false
	}

	return models.Card{
		ColumnID: input.ColumnID,
		Name:     input.Name,
		Content:  input.Content,
	}, true
}
