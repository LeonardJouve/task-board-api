package api

import (
	"strings"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func users(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		if strings.Split(c.Path(), "/")[3] != "me" {
			break
		}
		return getMe(c)
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "not found",
	})
}

func getMe(c *fiber.Ctx) error {
	user, err := getUser(c)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(user.Sanitize())
}

func getUser(c *fiber.Ctx) (models.User, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return models.User{}, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return user, nil
}

func getUserBoard(c *fiber.Ctx, boardId uint) (models.Board, error) {
	user, err := getUser(c)
	if err != nil {
		return models.Board{}, err
	}
	store.Database.Model(&user).Preload("Boards").First(&user)

	var board models.Board
	for _, b := range user.Boards {
		if b.ID == boardId {
			board = b
			break
		}
	}

	if board.ID == 0 {
		return models.Board{}, c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	return board, nil
}

func getUserBoardIds(c *fiber.Ctx) ([]uint, error) {
	user, err := getUser(c)
	if err != nil {
		return []uint{}, err
	}

	var boardIds []uint
	for _, board := range user.Boards {
		boardIds = append(boardIds, board.ID)
	}

	return boardIds, nil
}

func getUserColumn(c *fiber.Ctx, columnId uint) (models.Column, error) {
	var column models.Column
	store.Database.First(&column, columnId)
	if column.ID == 0 {
		return models.Column{}, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	if _, err := getUserBoard(c, column.BoardID); err != nil {
		return models.Column{}, err
	}

	return column, nil
}

func getUserColumnIds(c *fiber.Ctx) ([]uint, error) {
	boardIds, err := getUserBoardIds(c)
	if err != nil {
		return []uint{}, err
	}

	var columns []models.Column
	store.Database.Where("board_id IN ?", boardIds).Find(&columns)

	var columnIds []uint
	for _, column := range columns {
		columnIds = append(columnIds, column.ID)
	}

	return columnIds, nil
}

func getUserTag(c *fiber.Ctx, tagId uint) (models.Tag, error) {
	var tag models.Tag
	store.Database.First(&tag, tagId)
	if tag.ID == 0 {
		return models.Tag{}, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	if _, err := getUserBoard(c, tag.BoardID); err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func getUserCard(c *fiber.Ctx, cardId uint) (models.Card, error) {
	var card models.Card
	store.Database.First(&card, card.ID)
	if card.ID == 0 {
		return models.Card{}, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	if _, err := getUserColumn(c, card.ColumnID); err != nil {
		return models.Card{}, err
	}

	return card, nil
}
