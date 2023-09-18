package api

import (
	"github.com/LeonardJouve/task-board-api/models"
	"github.com/LeonardJouve/task-board-api/schema"
	"github.com/LeonardJouve/task-board-api/store"
	"github.com/gofiber/fiber/v2"
)

func GetBoards(c *fiber.Ctx) error {
	boards, ok := getUserBoards(c)
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(schema.SanitizeBoards(&boards))
}

func CreateBoard(c *fiber.Ctx) error {
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

func UpdateBoard(c *fiber.Ctx) error {
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

func DeleteBoard(c *fiber.Ctx) error {
	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
	}

	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
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

func InviteBoard(c *fiber.Ctx) error {
	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
	}

	userId := c.QueryInt("userId")

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

func LeaveBoard(c *fiber.Ctx) error {
	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
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
