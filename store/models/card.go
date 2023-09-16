package models

import "gorm.io/gorm"

type Card struct {
	gorm.Model
	ColumnID uint   `json:"columnId" validate:"required"`
	NextID   *uint  `json:"nextId"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
}
