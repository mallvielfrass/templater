package exelreader

import (
	"errors"
	"strings"
)

func (s *sheet) GetColumnCount() int {
	startColumn := columnNumber(s.StartColumn)
	endColumn := columnNumber(s.EndColumn)
	return startColumn - endColumn
}
func (s *sheet) GetRowCount() int {
	return s.EndRow - s.StartRow
}

func parseDimension(dimension string) (sheet, error) {
	if dimension == "" {
		return sheet{}, errors.New("dimension is empty")
	}
	if !strings.Contains(dimension, ":") {
		return sheet{}, errors.New("dimension is not valid")
	}
	parts := strings.Split(dimension, ":")
	if len(parts) != 2 {
		return sheet{}, errors.New("dimension is not valid")
	}

	// Разбираем начальную ячейку
	startCol, startRow, err := parseCell(parts[0])
	if err != nil {
		return sheet{}, errors.New("dimension is not valid")
	}

	// Разбираем конечную ячейку
	endCol, endRow, err := parseCell(parts[1])
	if err != nil {
		return sheet{}, errors.New("dimension is not valid")
	}

	s := sheet{
		StartColumn: startCol,
		EndColumn:   endCol,
		StartRow:    startRow,
		EndRow:      endRow,
	}

	// s.ColumnCount = s.getColumnCount()
	// s.RowCount = s.getRowCount()

	return s, nil
}
