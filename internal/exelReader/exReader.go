package exelreader

import (
	"bytes"

	"github.com/mallvielfrass/templater/internal/models"
	"github.com/xuri/excelize/v2"
)

func newReader(file *excelize.File, name string) ExelFile {
	return ExelFile{
		file:      file,
		name:      name,
		sheets:    NewSyncSheetMap(),
		isScanned: false,
	}
}

func CreateFile() (ExelFile, error) {
	file := excelize.NewFile()
	return newReader(file, ""), nil
}
func ReadBuffer(path string, data []byte) (ExelFile, error) {
	// Преобразуем []byte в io.Reader
	reader := bytes.NewReader(data)

	file, err := excelize.OpenReader(reader)
	if err != nil {
		return ExelFile{}, err
	}
	return newReader(file, path), nil

}
func ReadFile(filePath string) (ExelFile, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return ExelFile{}, err
	}
	return newReader(file, filePath), nil
}
func (e *ExelFile) WriteBuffer() (*bytes.Buffer, error) {
	return e.file.WriteToBuffer()
}
func (e *ExelFile) Scan() error {
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
func (e *ExelFile) FileInfo() (models.FileInfo, error) {
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
