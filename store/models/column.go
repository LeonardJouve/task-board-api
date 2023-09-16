package models

import "gorm.io/gorm"

type Column struct {
	gorm.Model
	BoardID uint   `json:"boardId" validate:"required"`
	NextID  *uint  `json:"nextId"`
	Name    string `json:"name" validate:"required"`
}
