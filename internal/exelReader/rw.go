package exelreader

import (
	"strconv"
)

func convertCellAddress(column string, row int) string {

	return column + strconv.Itoa(row)
}
func (e *exelFile) ReadCell(tableName string, cell string) (string, error) {
	return e.file.GetCellValue(tableName, cell)
}
func (e *exelFile) WriteCell(tableName string, cell string, value interface{}) error {
	err := e.file.SetCellValue(tableName, cell, value)
	if err != nil {
		return err
	}

	return e.sheets.ExtendDimension(tableName, cell)
}
func (e *exelFile) UpdateDimension(tableName string) error {
	return e.sheets.UpdateDimension(e.file, tableName)
}
func (e *exelFile) GetSheetDimension(tableName string) (string, error) {
	return e.file.GetSheetDimension(tableName)
}
