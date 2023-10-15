package api

import (
	"errors"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetUsers(c *fiber.Ctx) error {
	tx := store.Database.Model(&models.User{})

	var users []models.User

	if boardIdsQuery := c.Query("boardIds"); len(boardIdsQuery) != 0 {
		boardIds, ok := getQueryUIntArray(c, "boardIds")
		if !ok {
			return nil
		}

		tx = tx.Preload("Boards").Where("boards.id IN ?", boardIds)
	}

	if tx.Find(&users).Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeUsers(&users))
}

func GetUser(c *fiber.Ctx) error {
	userId, ok := getParamInt(c, "user_id")
	if !ok {
		return nil
	}

	var user models.User
	if err := store.Database.Model(&models.User{}).Where("id = ?", userId).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeUser(&user))
}

func GetMe(c *fiber.Ctx) error {
	user, ok := getUser(c)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeUser(&user))
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
	if err := store.Database.Model(&user).Preload("Boards").First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return []models.Board{}, false
	}

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
	if err := store.Database.First(&column, columnId).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return models.Column{}, false
	}
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
	if err := store.Database.Where("board_id IN ?", boardIds).Find(&columns).Error; err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return []uint{}, false
	}
	var columnIds []uint
	for _, column := range columns {
		columnIds = append(columnIds, column.ID)
	}

	return columnIds, true
}

func getUserTag(c *fiber.Ctx, tagId uint) (models.Tag, bool) {
	var tag models.Tag
	if err := store.Database.First(&tag, tagId).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return models.Tag{}, false
	}
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
	if err := store.Database.First(&card, cardId).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
		return models.Card{}, false
	}
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
