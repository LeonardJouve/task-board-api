package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	BoardID uint
	Board   Board `gorm:"constraint:OnDelete:CASCADE"`
	Name    string
	Color   string
}
