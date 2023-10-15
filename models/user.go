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
	Username            string
	Password            string
	Picture             string `gorm:"default:'default_profile_picture.png'"`
	TokenAvailableSince time.Time
}

type SanitizedUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Picture  string `json:"picture"`
}

func SanitizeUser(user *User) *SanitizedUser {
	return &SanitizedUser{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Username: user.Username,
		Picture:  user.Picture,
	}
}

func SanitizeUsers(users *[]User) *[]SanitizedUser {
	sanitizedUsers := []SanitizedUser{}
	for _, user := range *users {
		sanitizedUsers = append(sanitizedUsers, *SanitizeUser(&user))
	}

	return &sanitizedUsers
}
