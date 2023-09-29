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

type SanitizedCard struct {
	ID       uint   `json:"id"`
	ColumnID uint   `json:"columnId"`
	NextID   *uint  `json:"nextId"`
	TagIDs   []uint `json:"tagIds"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

func SanitizeCard(card *Card) *SanitizedCard {
	tagIds := []uint{}
	for _, tag := range card.Tags {
		tagIds = append(tagIds, tag.ID)
	}

	return &SanitizedCard{
		ID:       card.ID,
		ColumnID: card.ColumnID,
		NextID:   card.NextID,
		TagIDs:   tagIds,
		Name:     card.Name,
		Content:  card.Content,
	}
}

func SanitizeCards(cards *[]Card) *[]SanitizedCard {
	sanitizedCards := []SanitizedCard{}
	for _, card := range *cards {
		sanitizedCards = append(sanitizedCards, *(SanitizeCard(&card)))
	}

	return &sanitizedCards
}

func SortCards(cards *[]Card) *[]Card {
	var sortedCards []Card
	var card Card
	var ok bool
	cardMap := make(map[uint]Card, len(*cards))
	for _, c := range *cards {
		if c.NextID == nil {
			card = c
			continue
		}
		cardMap[*c.NextID] = c
	}

	if card.ID == 0 {
		return &[]Card{}
	}

	for {
		sortedCards = append([]Card{card}, sortedCards...)
		card, ok = cardMap[card.ID]
		if !ok {
			break
		}
	}

	return &sortedCards
}
