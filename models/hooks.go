package models

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var HookChannel = make(chan string)

func sendHookMessage(value interface{}) {
	formattedValue, err := json.Marshal(value)
	if err != nil {
		return
	}

	HookChannel <- string(formattedValue)
}

func (board *Board) AfterCreate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":  "created",
		"board": SanitizeBoard(board),
	})

	return nil
}

func (column *Column) AfterCreate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":   "created",
		"column": SanitizeColumn(column),
	})

	return nil
}

func (card *Card) AfterCreate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "created",
		"card": SanitizeCard(card),
	})

	return nil
}

func (tag *Tag) AfterCreate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "created",
		"tag":  SanitizeTag(tag),
	})

	return nil
}

func (board *Board) AfterUpdate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":  "updated",
		"board": SanitizeBoard(board),
	})

	return nil
}

func (column *Column) AfterUpdate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":   "updated",
		"column": SanitizeColumn(column),
	})

	return nil
}

func (card *Card) AfterUpdate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "updated",
		"card": SanitizeCard(card),
	})

	return nil
}

func (tag *Tag) AfterUpdate(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "updated",
		"tag":  SanitizeTag(tag),
	})

	return nil
}

func (board *Board) AfterDelete(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":  "deleted",
		"board": SanitizeBoard(board),
	})

	return nil
}

func (column *Column) AfterDelete(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type":   "deleted",
		"column": SanitizeColumn(column),
	})

	return nil
}

func (card *Card) AfterDelete(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "deleted",
		"card": SanitizeCard(card),
	})

	return nil
}

func (tag *Tag) AfterDelete(tx *gorm.DB) (err error) {
	sendHookMessage(fiber.Map{
		"type": "deleted",
		"tag":  SanitizeTag(tag),
	})

	return nil
}
