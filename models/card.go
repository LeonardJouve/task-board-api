package models

import (
	"github.com/LeonardJouve/task-board-api/store"
	"gorm.io/gorm"
)

type Card struct {
	gorm.Model
	ColumnID uint
	Column   Column `gorm:"constraint:OnDelete:CASCADE"`
	NextID   *uint
	Next     *Card  `gorm:"foreignKey:NextID"`
	Users    []User `gorm:"many2many:card_users;constraint:OnDelete:CASCADE"`
	Tags     []Tag  `gorm:"many2many:card_tags;constraint:OnDelete:CASCADE"`
	Name     string
	Content  string
}

type SanitizedCard struct {
	ID       uint   `json:"id"`
	ColumnID uint   `json:"columnId"`
	NextID   *uint  `json:"nextId"`
	UserIDs  []uint `json:"userIds"`
	TagIDs   []uint `json:"tagIds"`
	Name     string `json:"name"`
	Content  string `json:"content"`
}

func SanitizeCard(card *Card) *SanitizedCard {
	store.Database.Model(&card).Preload("Tags").Preload("Users").Find(&card)

	tagIds := []uint{}
	for _, tag := range card.Tags {
		tagIds = append(tagIds, tag.ID)
	}

	userIds := []uint{}
	for _, user := range card.Users {
		userIds = append(userIds, user.ID)
	}

	return &SanitizedCard{
		ID:       card.ID,
		ColumnID: card.ColumnID,
		NextID:   card.NextID,
		UserIDs:  userIds,
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
	lastCards := make(map[uint]Card)
	cardsMap := make(map[uint]map[uint]Card)
	var ok bool
	for _, c := range *cards {
		if c.NextID == nil {
			lastCards[c.ColumnID] = c
			continue
		}
		if len(cardsMap[c.ColumnID]) == 0 {
			cardsMap[c.ColumnID] = make(map[uint]Card)
		}
		cardsMap[c.ColumnID][*c.NextID] = c
	}

	if len(lastCards) == 0 {
		return &[]Card{}
	}

	for columnId, card := range lastCards {
		for {
			sortedCards = append([]Card{card}, sortedCards...)
			card, ok = cardsMap[columnId][card.ID]
			if !ok {
				break
			}
		}
	}

	return &sortedCards
}
