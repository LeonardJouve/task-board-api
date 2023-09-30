package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Boards              []Board `gorm:"many2many:user_boards;constraint:OnDelete:CASCADE"`
	Name                string
	Email               string `gorm:"unique"`
	Password            string
	TokenAvailableSince time.Time
}

type SanitizedUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func SanitizeUser(user *User) *SanitizedUser {
	return &SanitizedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
