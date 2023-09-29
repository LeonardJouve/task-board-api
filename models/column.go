package models

import "gorm.io/gorm"

type Column struct {
	gorm.Model
	BoardID uint
	Board   Board `gorm:"constraint:OnDelete:CASCADE"`
	NextID  *uint
	Name    string
}

type SanitizedColumn struct {
	ID      uint   `json:"id"`
	BoardID uint   `json:"boardId"`
	NextID  *uint  `json:"nextId"`
	Name    string `json:"name"`
}

func SanitizeColumn(column *Column) *SanitizedColumn {
	return &SanitizedColumn{
		ID:      column.ID,
		BoardID: column.BoardID,
		NextID:  column.NextID,
		Name:    column.Name,
	}
}

func SanitizeColumns(columns *[]Column) *[]SanitizedColumn {
	sanitizedColumns := []SanitizedColumn{}
	for _, column := range *columns {
		sanitizedColumns = append(sanitizedColumns, *(SanitizeColumn(&column)))
	}

	return &sanitizedColumns
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
