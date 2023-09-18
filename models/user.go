package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Boards   []Board `gorm:"many2many:user_boards;constraint:OnDelete:CASCADE"`
	Name     string
	Email    string `gorm:"unique"`
	Password string
}
