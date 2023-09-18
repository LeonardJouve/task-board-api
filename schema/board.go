package schema

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/gofiber/fiber/v2"
)

type SanitizedBoard struct {
	ID      uint   `json:"id"`
	OwnerID uint   `json:"ownerId"`
	Name    string `json:"name"`
}

func SanitizeBoard(board *models.Board) *SanitizedBoard {
	return &SanitizedBoard{
		ID:      board.ID,
		OwnerID: board.OwnerID,
		Name:    board.Name,
	}
}

func SanitizeBoards(boards *[]models.Board) *[]SanitizedBoard {
	sanitizedBoards := []SanitizedBoard{}
	for _, board := range *boards {
		sanitizedBoards = append(sanitizedBoards, *(SanitizeBoard(&board)))
	}

	return &sanitizedBoards
}

type UpsertBoardInput struct {
	Name string `json:"name" validate:"required"`
}

func GetUpsertBoardInput(c *fiber.Ctx) (models.Board, bool) {
	var input UpsertBoardInput
	if err := c.BodyParser(&input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}
	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
		return models.Board{}, false
	}

	return models.Board{
		Name: input.Name,
	}, true
}
