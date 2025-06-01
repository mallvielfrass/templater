package exelreader

import (
	"github.com/xuri/excelize/v2"
)

type sheet struct {
	Name        string
	StartColumn string
	EndColumn   string
	StartRow    int
	EndRow      int

	// RowCount    int
	// ColumnCount int
}
type fileInfo struct {
	Sheets   []sheet
	FileName string
	//Size     int
}

type exelFile struct {
	file      *excelize.File
	name      string
	isScanned bool
	sheets    *syncSheetMap // map[string]*sheet
}
