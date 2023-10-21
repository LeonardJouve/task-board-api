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
	lastColumns := make(map[uint]Column)
	columnsMap := make(map[uint]map[uint]Column)
	var ok bool
	for _, c := range *columns {
		if c.NextID == nil {
			lastColumns[c.BoardID] = c
			continue
		}
		if len(columnsMap[c.BoardID]) == 0 {
			columnsMap[c.BoardID] = make(map[uint]Column)
		}
		columnsMap[c.BoardID][*c.NextID] = c
	}

	if len(lastColumns) == 0 {
		return &[]Column{}
	}

	for boardId, column := range lastColumns {
		for {
			sortedColumns = append([]Column{column}, sortedColumns...)
			column, ok = columnsMap[boardId][column.ID]
			if !ok {
				break
			}
		}
	}

	return &sortedColumns
}
