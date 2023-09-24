package models

import "gorm.io/gorm"

type Column struct {
	gorm.Model
	BoardID uint
	Board   Board `gorm:"constraint:OnDelete:CASCADE"`
	NextID  *uint
	Name    string
}

func SortColumns(columns *[]Column) *[]Column {
	var sortedColumns []Column
	var column Column
	var ok bool
	columnMap := make(map[uint]Column, len(*columns))
	for _, c := range *columns {
		if c.NextID == nil {
			column = c
			continue
		}
		columnMap[*c.NextID] = c
	}

	if column.ID == 0 {
		return &[]Column{}
	}

	for {
		sortedColumns = append([]Column{column}, sortedColumns...)
		column, ok = columnMap[column.ID]
		if !ok {
			break
		}
	}

	return &sortedColumns
}
