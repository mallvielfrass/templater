package exelreader

func (e *exelFile) SheetInfo(sheetName string) (sheet, error) {
	dimension, err := e.file.GetSheetDimension(sheetName)
	if err != nil {
		return sheet{}, err
	}
	parsedDimension, err := parseDimension(dimension)
	if err != nil {
		return sheet{}, err
	}
	return parsedDimension, nil

}
func (e *exelFile) CreateSheet(sheetName string) error {
	_, err := e.file.NewSheet(sheetName)
	if err != nil {
		return err
	}
	e.file.SetSheetDimension(sheetName, "A1:A1")
	e.sheets.Store(sheetName, &sheet{
		Name:        sheetName,
		StartColumn: "A",
		EndColumn:   "A",
		StartRow:    1,
		EndRow:      1,
	})
	return nil
}
