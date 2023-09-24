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
