package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetMe(c *fiber.Ctx) error {
	user, ok := getUser(c)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeUser(&user))
}

func getUser(c *fiber.Ctx) (models.User, bool) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return models.User{}, false
	}

	return user, true
}

func getUserBoards(c *fiber.Ctx) ([]models.Board, bool) {
	user, ok := getUser(c)
	if !ok {
		return []models.Board{}, false
	}
	store.Database.Model(&user).Preload("Boards").First(&user)

	return user.Boards, true
}

func getUserBoard(c *fiber.Ctx, boardId uint) (models.Board, bool) {
	boards, ok := getUserBoards(c)
	if !ok {
		return models.Board{}, false
	}

	var board models.Board
	for _, b := range boards {
		if b.ID == boardId {
			board = b
			break
		}
	}
	if board.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Board{}, false
	}

	return board, true
}

func getUserBoardIds(c *fiber.Ctx) ([]uint, bool) {
	boards, ok := getUserBoards(c)
	if !ok {
		return []uint{}, false
	}

	var boardIds []uint
	for _, board := range boards {
		boardIds = append(boardIds, board.ID)
	}

	return boardIds, true
}

func getUserColumn(c *fiber.Ctx, columnId uint) (models.Column, bool) {
	var column models.Column
	store.Database.First(&column, columnId)
	if column.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Column{}, false
	}

	if _, ok := getUserBoard(c, column.BoardID); !ok {
		return models.Column{}, false
	}

	return column, true
}

func getUserColumnIds(c *fiber.Ctx) ([]uint, bool) {
	boardIds, ok := getUserBoardIds(c)
	if !ok {
		return []uint{}, false
	}

	var columns []models.Column
	store.Database.Where("board_id IN ?", boardIds).Find(&columns)

	var columnIds []uint
	for _, column := range columns {
		columnIds = append(columnIds, column.ID)
	}

	return columnIds, true
}

func getUserTag(c *fiber.Ctx, tagId uint) (models.Tag, bool) {
	var tag models.Tag
	store.Database.First(&tag, tagId)
	if tag.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Tag{}, false
	}

	if _, ok := getUserBoard(c, tag.BoardID); !ok {
		return models.Tag{}, false
	}

	return tag, true
}

func getUserCard(c *fiber.Ctx, cardId uint) (models.Card, bool) {
	var card models.Card
	store.Database.First(&card, cardId)
	if card.ID == 0 {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
		return models.Card{}, false
	}

	if _, ok := getUserColumn(c, card.ColumnID); !ok {
		return models.Card{}, false
	}

	return card, true
}
