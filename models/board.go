package models

import (
	"github.com/LeonardJouve/task-board-api/store"
	"gorm.io/gorm"
)

type Board struct {
	gorm.Model
	OwnerID     uint
	Owner       User   `gorm:"foreignKey:OwnerID"`
	Users       []User `gorm:"many2many:user_boards"`
	Name        string
	Description string
}

type SanitizedBoard struct {
	ID          uint   `json:"id"`
	OwnerID     uint   `json:"ownerId"`
	UserIds     []uint `json:"userIds"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func SanitizeBoard(board *Board) *SanitizedBoard {
	store.Database.Model(&board).Preload("Users").Find(&board)

	userIds := []uint{}
	for _, tag := range board.Users {
		userIds = append(userIds, tag.ID)
	}

	return &SanitizedBoard{
		ID:          board.ID,
		OwnerID:     board.OwnerID,
		UserIds:     userIds,
		Name:        board.Name,
		Description: board.Description,
	}
}

func SanitizeBoards(boards *[]Board) *[]SanitizedBoard {
	sanitizedBoards := []SanitizedBoard{}
	for _, board := range *boards {
		sanitizedBoards = append(sanitizedBoards, *(SanitizeBoard(&board)))
	}

	return &sanitizedBoards
}
