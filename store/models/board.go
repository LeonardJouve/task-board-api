package models

import "gorm.io/gorm"

type Board struct {
	gorm.Model
	Columns []Column `json:"columns"`
	Tags    []Tag    `json:"tags"`
	NextID  *uint    `json:"nextId"`
	Name    string   `json:"name" validate:"required"`
}
