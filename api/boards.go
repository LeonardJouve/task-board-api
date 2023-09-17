package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/store"
	"github.com/LeonardJouve/task-board-api/store/models"
	"github.com/gofiber/fiber/v2"
)

func boards(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		if paths := strings.Split(c.Path(), "/"); len(paths) == 5 && paths[3] == "invite" {
			return inviteUser(c)
		}
		return getBoards(c)
	case "POST":
		return createBoard(c)
	case "PUT":
		return updateBoard(c)
	case "PATCH":
		return moveBoard(c)
	case "DELETE":
		return deleteBoard(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getBoards(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	store.Database.Model(&user).Preload("Boards").First(&user)

	return c.Status(fiber.StatusOK).JSON(user.Boards)
}

func createBoard(c *fiber.Ctx) error {
	var board models.Board

	if err := c.BodyParser(&board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := store.Database.Create(&board)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": result.Error.Error(),
		})
	}

	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	store.Database.Model(&user).Association("Boards").Append([]models.Board{board})

	var previous models.Board
	store.Database.Where("next_id IS NULL AND id != ?", board.ID).First(&previous)
	if previous.ID != 0 {
		previous.NextID = &board.ID
		store.Database.Save(&previous)
	}

	return c.Status(fiber.StatusCreated).JSON(board)
}

func updateBoard(c *fiber.Ctx) error {
	var board models.Board
	if err := c.BodyParser(&board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := validate.Struct(board); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var currentBoard models.Board
	store.Database.First(&currentBoard, board.ID)
	if currentBoard.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	store.Database.Model(&models.Board{}).Where("id = ?", board.ID).Omit("NextID").Updates(&board)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func moveBoard(c *fiber.Ctx) error {
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

	nextId := c.QueryInt("nextId")
	boardId, err := strconv.ParseUint(paths[4], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	var board models.Board
	store.Database.First(&board, boardId)
	if board.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	store.Database.Model(&models.Board{}).Where("next_id = ?", board.ID).Update("next_id", board.NextID)
	if nextId <= 0 {
		store.Database.Model(&models.Board{}).Where("next_id IS NULL").Update("next_id", &board.ID)
		store.Database.Model(&board).Update("next_id", nil)
	} else {
		var previous models.Board
		store.Database.Find(&previous, nextId)
		if previous.ID == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "not found",
			})
		}

		store.Database.Model(&models.Board{}).Where("next_id = ?", nextId).Update("next_id", &board.ID)
		store.Database.Model(&board).Update("next_id", nextId)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func deleteBoard(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) < 4 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	boardId, err := strconv.ParseUint(paths[3], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var board models.Board
	store.Database.First(&board, boardId)
	if board.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	var columns []models.Column
	store.Database.Where("board_id = ?", boardId).Find(&columns)
	for _, column := range columns {
		store.Database.Unscoped().Where("column_id = ?", column.ID).Delete(&[]models.Card{})
	}
	store.Database.Unscoped().Where("board_id = ?", boardId).Delete(&[]models.Column{})
	var previous models.Board
	if store.Database.Where("next_id = ?", boardId).First(&previous); previous.ID != 0 {
		previous.NextID = board.NextID
		store.Database.Save(&previous)
	}
	store.Database.Unscoped().Where("board_id = ?", boardId).Delete(&[]models.Tag{})
	store.Database.Unscoped().Delete(&board)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func inviteUser(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) != 5 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
	boardId64, err := strconv.ParseUint(paths[4], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	boardId := uint(boardId64)
	userId, err := strconv.ParseUint(c.Query("userId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	var user models.User
	store.Database.First(&user, userId)
	if user.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	currentUser, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "server error",
		})
	}
	store.Database.Model(&currentUser).Preload("Boards").First(&currentUser)

	var board models.Board
	for _, b := range currentUser.Boards {
		if b.ID == boardId {
			board = b
			break
		}
	}
	if board.ID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	store.Database.Model(&user).Association("Boards").Append([]models.Board{board})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
