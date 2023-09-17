package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Boards   []Board `json:"boards" gorm:"many2many:user_boards;"`
	Name     string  `json:"name"`
	Email    string  `json:"email" gorm:"unique"`
	Password string  `json:"password"`
}

type SanitizedUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (user *User) Sanitize() *SanitizedUser {
	return &SanitizedUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
