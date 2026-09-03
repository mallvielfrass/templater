package models

import "github.com/mallvielfrass/templater/internal/utils/cell"

type Sheet struct {
	Name        string `json:"name"`
	StartColumn string `json:"start_column"`
	EndColumn   string `json:"end_column"`
	StartRow    int    `json:"start_row"`
	EndRow      int    `json:"end_row"`

	// RowCount    int
	// ColumnCount int
}

func (s *Sheet) GetColumnCount() int {
	startColumn := cell.ColumnNumber(s.StartColumn)
	endColumn := cell.ColumnNumber(s.EndColumn)
	return endColumn - startColumn + 1
}
func (s *Sheet) GetRowCount() int {
	return s.EndRow - s.StartRow + 1
}
