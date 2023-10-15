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

	return c.Status(fiber.StatusOK).JSON(models.SanitizeBoards(&boards))
}

func GetBoard(c *fiber.Ctx) error {
	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
	}

	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(models.SanitizeBoard(&board))
}

func CreateBoard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	board, ok := schema.GetCreateBoardInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}
	board.OwnerID = user.ID

	if ok := store.Execute(c, tx, tx.Create(&board).Error); !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&user).Association("Boards").Append([]models.Board{board})); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusCreated).JSON(models.SanitizeBoard(&board))
}

func UpdateBoard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
	}

	if _, ok := getUserBoard(c, uint(boardId)); !ok {
		return nil
	}

	board, ok := schema.GetUpdateBoardInput(c, uint(boardId))
	if !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&board).Updates(&board).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(models.SanitizeBoard(&board))
}

func DeleteBoard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

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

	if ok := store.Execute(c, tx, tx.Unscoped().Delete(&board).Error); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

func InviteBoard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

	boardId, ok := getParamInt(c, "board_id")
	if !ok {
		return nil
	}

	userId := c.QueryInt("userId")

	var user models.User
	if ok := store.Execute(c, tx, tx.First(&user, userId).Error); !ok {
		return nil
	}
	if user.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	}

	board, ok := getUserBoard(c, uint(boardId))
	if !ok {
		return nil
	}

	if ok := store.Execute(c, tx, tx.Model(&user).Association("Boards").Append([]models.Board{board})); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

func LeaveBoard(c *fiber.Ctx) error {
	tx, ok := store.BeginTransaction(c)
	if !ok {
		return nil
	}

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

	if ok := store.Execute(c, tx, tx.Model(&user).Association("Boards").Delete(&board)); !ok {
		return nil
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
