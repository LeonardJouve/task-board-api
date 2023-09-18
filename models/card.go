package models

import (
	"gorm.io/gorm"
)

type Card struct {
	gorm.Model
	ColumnID uint
	Column   Column `gorm:"constraint:OnDelete:CASCADE"`
	NextID   *uint
	Next     *Card `gorm:"foreignKey:NextID"`
	Tags     []Tag `gorm:"many2many:card_tags;constraint:OnDelete:CASCADE"`
	Name     string
	Content  string
}
