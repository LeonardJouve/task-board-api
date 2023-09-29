package models

import (
	"gorm.io/gorm"
)

const (
	CREATED_TYPE = "created"
	UPDATED_TYPE = "updated"
	DELETED_TYPE = "deleted"
)

type HookMessage = struct {
	BoardId uint
	Type    string
	Message map[string]interface{}
}

var HookChannel = make(chan HookMessage)

func (board *Board) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: board.ID,
		Type:    CREATED_TYPE,
		Message: map[string]interface{}{
			"board": SanitizeBoard(board),
		},
	}

	return nil
}

func (column *Column) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: column.BoardID,
		Type:    CREATED_TYPE,
		Message: map[string]interface{}{
			"column": SanitizeColumn(column),
		},
	}

	return nil
}

func (card *Card) AfterCreate(tx *gorm.DB) (err error) {
	tx.Model(card).Preload("Column").First(card)

	HookChannel <- HookMessage{
		BoardId: card.Column.BoardID,
		Type:    CREATED_TYPE,
		Message: map[string]interface{}{
			"card": SanitizeCard(card),
		},
	}

	return nil
}

func (tag *Tag) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: tag.BoardID,
		Type:    CREATED_TYPE,
		Message: map[string]interface{}{
			"tag": SanitizeTag(tag),
		},
	}

	return nil
}

func (board *Board) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: board.ID,
		Type:    UPDATED_TYPE,
		Message: map[string]interface{}{
			"board": SanitizeBoard(board),
		},
	}

	return nil
}

func (column *Column) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: column.BoardID,
		Type:    UPDATED_TYPE,
		Message: map[string]interface{}{
			"column": SanitizeColumn(column),
		},
	}

	return nil
}

func (card *Card) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Model(card).Preload("Column").First(card)

	HookChannel <- HookMessage{
		BoardId: card.Column.BoardID,
		Type:    UPDATED_TYPE,
		Message: map[string]interface{}{
			"card": SanitizeCard(card),
		},
	}

	return nil
}

func (tag *Tag) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: tag.BoardID,
		Type:    UPDATED_TYPE,
		Message: map[string]interface{}{
			"tag": SanitizeTag(tag),
		},
	}

	return nil
}

func (board *Board) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: board.ID,
		Type:    DELETED_TYPE,
		Message: map[string]interface{}{
			"board": SanitizeBoard(board),
		},
	}

	return nil
}

func (column *Column) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: column.BoardID,
		Type:    DELETED_TYPE,
		Message: map[string]interface{}{
			"column": SanitizeColumn(column),
		},
	}

	return nil
}

func (card *Card) AfterDelete(tx *gorm.DB) (err error) {
	tx.Model(card).Preload("Column").First(card)

	HookChannel <- HookMessage{
		BoardId: card.Column.BoardID,
		Type:    DELETED_TYPE,
		Message: map[string]interface{}{
			"card": SanitizeCard(card),
		},
	}

	return nil
}

func (tag *Tag) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		BoardId: tag.BoardID,
		Type:    DELETED_TYPE,
		Message: map[string]interface{}{
			"tag": SanitizeTag(tag),
		},
	}

	return nil
}
