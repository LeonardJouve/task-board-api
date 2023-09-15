package models

import "gorm.io/gorm"

type Column struct {
	gorm.Model
	BoardID uint   `json:"boardId" validate:"required"`
	NextID  *uint  `json:"nextId"`
	Cards   []Card `json:"cards"`
	Name    string `json:"name" validate:"required"`
}
