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
