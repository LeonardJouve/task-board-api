package models

import (
	"gorm.io/gorm"
)

type Board struct {
	gorm.Model
	Name    string
	OwnerID uint
	Owner   User `gorm:"foreignKey:OwnerID"`
}

type SanitizedBoard struct {
	ID      uint   `json:"id"`
	OwnerID uint   `json:"ownerId"`
	Name    string `json:"name"`
}

func SanitizeBoard(board *Board) *SanitizedBoard {
	return &SanitizedBoard{
		ID:      board.ID,
		OwnerID: board.OwnerID,
		Name:    board.Name,
	}
}

func SanitizeBoards(boards *[]Board) *[]SanitizedBoard {
	sanitizedBoards := []SanitizedBoard{}
	for _, board := range *boards {
		sanitizedBoards = append(sanitizedBoards, *(SanitizeBoard(&board)))
	}

	return &sanitizedBoards
}
