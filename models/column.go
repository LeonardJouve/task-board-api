package models

import "gorm.io/gorm"

type Column struct {
	gorm.Model
	BoardID uint
	Board   Board `gorm:"constraint:OnDelete:CASCADE"`
	NextID  *uint
	Name    string
}
