package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func boards(c *fiber.Ctx) error {
	switch c.Method() {
	case "GET":
		paths := strings.Split(c.Path(), "/")
		if len(paths) == 5 && paths[4] == "invite" {
			return inviteUser(c)
		}
		if len(paths) == 5 && paths[4] == "leave" {
			return leaveBoard(c)
		}
		return getBoards(c)
	case "POST":
		return createBoard(c)
	case "PUT":
		return updateBoard(c)
	case "DELETE":
		return deleteBoard(c)
	default:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}
}

func getBoards(c *fiber.Ctx) error {
	boards, ok := getUserBoards(c)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeBoards(&boards))
}

func createBoard(c *fiber.Ctx) error {
	board, ok := schema.GetUpsertBoardInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}
	board.OwnerID = user.ID

	if err := store.Database.Create(&board).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	store.Database.Model(&user).Association("Boards").Append([]models.Board{board})

	return c.Status(fiber.StatusCreated).JSON(schema.SanitizeBoard(&board))
}

func updateBoard(c *fiber.Ctx) error {
	board, ok := schema.GetUpsertBoardInput(c)
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, board.ID); !ok {
		return nil
	}

	store.Database.Model(&board).Omit("OwnerID").Updates(&board)

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeBoard(&board))
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

	user, ok := getUser(c)
	if !ok {
		return nil
	}
	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}
	if board.OwnerID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

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
	boardId, err := strconv.ParseUint(paths[3], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
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

	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}

	store.Database.Model(&user).Association("Boards").Append([]models.Board{board})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func leaveBoard(c *fiber.Ctx) error {
	paths := strings.Split(c.Path(), "/")
	if len(paths) != 5 {
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
	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}
	user, ok := getUser(c)
	if !ok {
		return nil
	}

	if board.OwnerID == user.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "owner cannot leave board",
		})
	}

	store.Database.Model(&user).Association("Boards").Delete(&board)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
