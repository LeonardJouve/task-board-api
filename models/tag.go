package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	BoardID uint
	Board   Board  `gorm:"constraint:OnDelete:CASCADE"`
	Cards   []Card `gorm:"many2many:card_tags;constraint:OnDelete:CASCADE"`
	Name    string
	Color   string
}

type SanitizedTag struct {
	ID      uint   `json:"id"`
	BoardID uint   `json:"boardId"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}

func SanitizeTag(tag *Tag) *SanitizedTag {
	return &SanitizedTag{
		ID:      tag.ID,
		BoardID: tag.BoardID,
		Name:    tag.Name,
		Color:   tag.Color,
	}
}

func SanitizeTags(tags *[]Tag) *[]SanitizedTag {
	sanitizedTags := []SanitizedTag{}
	for _, tag := range *tags {
		sanitizedTags = append(sanitizedTags, *(SanitizeTag(&tag)))
	}

	return &sanitizedTags
}
