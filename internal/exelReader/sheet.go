package exelreader

import "github.com/mallvielfrass/templater/internal/models"

func (e *exelFile) SheetInfo(sheetName string) (models.Sheet, error) {
	dimension, err := e.file.GetSheetDimension(sheetName)
	if err != nil {
		return models.Sheet{}, err
	}
	parsedDimension, err := parseDimension(dimension)
	if err != nil {
		return models.Sheet{}, err
	}
	return parsedDimension, nil

}
func (e *exelFile) CreateSheet(sheetName string) error {
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
