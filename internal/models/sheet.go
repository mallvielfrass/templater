package models

import "github.com/mallvielfrass/templater/internal/utils/cell"

type Sheet struct {
	Name        string
	StartColumn string
	EndColumn   string
	StartRow    int
	EndRow      int

	// RowCount    int
	// ColumnCount int
}

func (s *Sheet) GetColumnCount() int {
	startColumn := cell.ColumnNumber(s.StartColumn)
	endColumn := cell.ColumnNumber(s.EndColumn)
	return startColumn - endColumn
}
func (s *Sheet) GetRowCount() int {
	return s.EndRow - s.StartRow
}
