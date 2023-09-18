package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type SanitizedColumn struct {
	ID      uint   `json:"id"`
	BoardID uint   `json:"boardId"`
	Name    string `json:"name"`
}

func SanitizeColumn(column *models.Column) *SanitizedColumn {
	return &SanitizedColumn{
		ID:      column.ID,
		BoardID: column.BoardID,
		Name:    column.Name,
	}
}

func SanitizeColumns(columns *[]models.Column) *[]SanitizedColumn {
	sanitizedColumns := []SanitizedColumn{}
	for _, column := range *columns {
		sanitizedColumns = append(sanitizedColumns, *(SanitizeColumn(&column)))
	}

	return &sanitizedColumns
}

type UpsertColumnInput struct {
	Name    string `json:"name" validate:"required"`
	BoardID uint   `json:"boardId" validate:"required"`
}

func GetUpsertColumnInput(c *fiber.Ctx) (models.Column, bool) {
	var input UpsertColumnInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Column{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Column{}, false
	}

	return models.Column{
		Name:    input.Name,
		BoardID: input.BoardID,
	}, true
}
