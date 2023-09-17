package models

import "gorm.io/gorm"

type Board struct {
	gorm.Model
	Name string `json:"name" validate:"required"`
}
