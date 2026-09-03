package exelreader

import (
	"strconv"
	"strings"

	"github.com/mallvielfrass/templater/internal/models"
	"github.com/mallvielfrass/templater/internal/utils/cell"
)

func (e *ExelFile) SheetInfo(sheetName string) (models.Sheet, error) {
	dimension, err := e.file.GetSheetDimension(sheetName)
	if err != nil {
		return models.Sheet{}, err
	}
	parsedDimension, err := parseDimension(dimension)
	if err != nil {
		return models.Sheet{}, err
	}
	parsedDimension.Name = sheetName

	rows, err := e.file.GetRows(sheetName)
	if err == nil && len(rows) > 0 {
		if len(rows) > parsedDimension.EndRow {
			parsedDimension.EndRow = len(rows)
		}
		maxCols := 0
		for _, r := range rows {
			if len(r) > maxCols {
				maxCols = len(r)
			}
		}
		if maxCols > 0 {
			endColNum := cell.ColumnNumber(parsedDimension.EndColumn)
			if maxCols > endColNum {
				parsedDimension.EndColumn = cell.ColumnChar(maxCols)
			}
		}
	}

	return parsedDimension, nil
}

func (e *ExelFile) CreateSheet(sheetName string) error {
	_, err := e.file.NewSheet(sheetName)
	if err != nil {
		return err
	}
	e.file.SetSheetDimension(sheetName, "A1:A1")
	e.sheets.Store(sheetName, &models.Sheet{
		Name:        sheetName,
		StartColumn: "A",
		EndColumn:   "A",
		StartRow:    1,
		EndRow:      1,
	})
	return nil
}

//	func (e *ExelFile) GetRow(sheetName string, row int) (map[string]string, error) {
//		rowData, err := e.file.GetRows(sheetName)
//		if err != nil {
//			return nil, err
//		}
//		return rowData, nil
//	}
func (e *ExelFile) ReadFirstRow(sheetName string) ([]string, error) {
	info, err := e.SheetInfo(sheetName)
	if err != nil {
		return nil, err
	}
	columnCount := info.GetColumnCount()
	rowData := []string{}
	for i := 1; i <= columnCount; i++ {
		columnString := cell.ColumnChar(i)
		cell := columnString + strconv.Itoa(1)
		cellValue, err := e.file.GetCellValue(sheetName, cell)
		if err != nil {
			return nil, err
		}
		rowData = append(rowData, strings.TrimSpace(cellValue))
	}
	return rowData, nil
}
func (e *ExelFile) ReadSheetRows(sheetName string, minRow int, maxRow int) ([][]string, error) {
	info, err := e.SheetInfo(sheetName)
	if err != nil {
		return nil, err
	}
	columnCount := info.GetColumnCount()

	rows := [][]string{}
	for row := minRow; row <= maxRow; row++ {
		rowData := []string{}
		for i := 1; i <= columnCount; i++ {
			columnString := cell.ColumnChar(i)
			cell := columnString + strconv.Itoa(row)
			cellValue, err := e.file.GetCellValue(sheetName, cell)
			if err != nil {
				return nil, err
			}
			rowData = append(rowData, cellValue)
		}
		rows = append(rows, rowData)
	}
	return rows, nil
}
