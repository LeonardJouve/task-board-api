package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type SanitizedTag struct {
	ID      uint   `json:"id"`
	BoardID uint   `json:"boardId"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}

func SanitizeTag(tag *models.Tag) *SanitizedTag {
	return &SanitizedTag{
		ID:      tag.ID,
		BoardID: tag.BoardID,
		Name:    tag.Name,
		Color:   tag.Color,
	}
}

func SanitizeTags(tags *[]models.Tag) *[]SanitizedTag {
	sanitizedTags := []SanitizedTag{}
	for _, tag := range *tags {
		sanitizedTags = append(sanitizedTags, *(SanitizeTag(&tag)))
	}

	return &sanitizedTags
}

type UpsertTagInput struct {
	BoardID uint   `json:"boardId" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Color   string `json:"color" validate:"required,color"`
}

func GetUpsertTagInput(c *fiber.Ctx) (models.Tag, bool) {
	var input UpsertTagInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Tag{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Tag{}, false
	}
	return models.Tag{
		BoardID: input.BoardID,
		Name:    input.Name,
		Color:   input.Color,
	}, true
}
