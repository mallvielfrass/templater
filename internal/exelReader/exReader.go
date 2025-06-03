package exelreader

import (
	"bytes"

	"github.com/xuri/excelize/v2"
)

func newReader(file *excelize.File, name string) exelFile {
	return exelFile{
		file:      file,
		name:      name,
		sheets:    NewSyncSheetMap(),
		isScanned: false,
	}
}

func CreateFile() (exelFile, error) {
	file := excelize.NewFile()
	return newReader(file, ""), nil
}

func ReadFile(filePath string) (exelFile, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return exelFile{}, err
	}
	return newReader(file, filePath), nil
}
func (e *exelFile) WriteBuffer() (*bytes.Buffer, error) {
	return e.file.WriteToBuffer()
}
func (e *exelFile) Scan() error {
	if e.isScanned {
		return nil
	}
	e.isScanned = true
	sheetList := e.file.GetSheetList()
	for _, sheetName := range sheetList {
		sheet, err := e.SheetInfo(sheetName)
		if err != nil {
			return err
		}
		e.sheets.Store(sheetName, &sheet)
	}
	return nil
}
func (e *exelFile) FileInfo() (models.FileInfo, error) {
	var sheets []models.Sheet
	sheetList := e.file.GetSheetList()
	for _, sheetName := range sheetList {
		sheet, err := e.SheetInfo(sheetName)
		if err != nil {
			return models.FileInfo{}, err
		}
		sheets = append(sheets, sheet)
	}
	//size:= e.file.
	return models.FileInfo{
		FileName: e.name,
		Sheets:   sheets,
	}, nil
}
