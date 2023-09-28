package models

import (
	"gorm.io/gorm"
)

const (
	CREATED_TYPE = "created"
	UPDATED_TYPE = "updated"
	DELETED_TYPE = "deleted"
)

type HookMessage map[string]interface{}

var HookChannel = make(chan HookMessage)

func (board *Board) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":  CREATED_TYPE,
		"board": SanitizeBoard(board),
	}

	return nil
}

func (column *Column) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":   CREATED_TYPE,
		"column": SanitizeColumn(column),
	}

	return nil
}

func (card *Card) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": CREATED_TYPE,
		"card": SanitizeCard(card),
	}

	return nil
}

func (tag *Tag) AfterCreate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": CREATED_TYPE,
		"tag":  SanitizeTag(tag),
	}

	return nil
}

func (board *Board) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":  UPDATED_TYPE,
		"board": SanitizeBoard(board),
	}

	return nil
}

func (column *Column) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":   UPDATED_TYPE,
		"column": SanitizeColumn(column),
	}

	return nil
}

func (card *Card) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": UPDATED_TYPE,
		"card": SanitizeCard(card),
	}

	return nil
}

func (tag *Tag) AfterUpdate(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": UPDATED_TYPE,
		"tag":  SanitizeTag(tag),
	}

	return nil
}

func (board *Board) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":  DELETED_TYPE,
		"board": SanitizeBoard(board),
	}

	return nil
}

func (column *Column) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type":   DELETED_TYPE,
		"column": SanitizeColumn(column),
	}

	return nil
}

func (card *Card) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": DELETED_TYPE,
		"card": SanitizeCard(card),
	}

	return nil
}

func (tag *Tag) AfterDelete(tx *gorm.DB) (err error) {
	HookChannel <- HookMessage{
		"type": DELETED_TYPE,
		"tag":  SanitizeTag(tag),
	}

	return nil
}
