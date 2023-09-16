package models

import "gorm.io/gorm"

type Board struct {
	gorm.Model
	NextID *uint  `json:"nextId"`
	Name   string `json:"name" validate:"required"`
}
