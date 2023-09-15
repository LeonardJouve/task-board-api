package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	BoardID uint   `json:"boardId" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Color   string `json:"color" validate:"required"`
}
