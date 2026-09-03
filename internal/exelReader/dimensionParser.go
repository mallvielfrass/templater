package exelreader

import (
	"errors"
	"strings"

	"github.com/mallvielfrass/templater/internal/models"
	"github.com/mallvielfrass/templater/internal/utils/cell"
)

func parseDimension(dimension string) (models.Sheet, error) {
	if dimension == "" {
		return models.Sheet{}, errors.New("dimension is empty")
	}
	if !strings.Contains(dimension, ":") {
		return models.Sheet{}, errors.New("dimension is not valid")
	}
	parts := strings.Split(dimension, ":")
	if len(parts) != 2 {
		return models.Sheet{}, errors.New("dimension is not valid")
	}

	// Разбираем начальную ячейку
	startCol, startRow, err := cell.ParseCell(parts[0])
	if err != nil {
		return models.Sheet{}, errors.New("dimension is not valid")
	}

	// Разбираем конечную ячейку
	endCol, endRow, err := cell.ParseCell(parts[1])
	if err != nil {
		return models.Sheet{}, errors.New("dimension is not valid")
	}

	s := models.Sheet{
		StartColumn: startCol,
		EndColumn:   endCol,
		StartRow:    startRow,
		EndRow:      endRow,
	}

	// s.ColumnCount = s.getColumnCount()
	// s.RowCount = s.getRowCount()

	return s, nil
}
