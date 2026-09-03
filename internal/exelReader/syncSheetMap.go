package exelreader

import (
	"errors"
	"strconv"
	"sync"

	"github.com/mallvielfrass/templater/internal/models"
	"github.com/mallvielfrass/templater/internal/utils/cell"
	"github.com/xuri/excelize/v2"
)

type syncSheetMap struct {
	mx sync.RWMutex
	m  map[string]*models.Sheet
}

func NewSyncSheetMap() *syncSheetMap {
	return &syncSheetMap{
		m: make(map[string]*models.Sheet),
	}
}

// Store save object to map
func (s *syncSheetMap) Store(sheetName string, sheet *models.Sheet) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[sheetName] = sheet
}
func (s *syncSheetMap) Load(sheetName string) (*models.Sheet, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.m[sheetName]
	return val, ok
}

func (s *syncSheetMap) ExtendDimension(sheetName, targetCell string) error {
	col, row, err := cell.ParseCell(targetCell)
	if err != nil {
		return errors.New("dimension is not valid")
	}
	colNum := cell.ColumnNumber(col)
	s.mx.Lock()
	defer s.mx.Unlock()
	sheet, ok := s.m[sheetName]
	if !ok {
		return errors.New("ExtendDimension: table not exist")
	}
	startColumn := cell.ColumnNumber(sheet.StartColumn)
	endColumn := cell.ColumnNumber(sheet.EndColumn)
	startRow := sheet.StartRow
	endRow := sheet.EndRow
	isEdited := false
	if colNum < startColumn {
		sheet.StartColumn = col
		isEdited = true
	}
	if endColumn < colNum {
		sheet.EndColumn = col
		isEdited = true
	}
	if row < startRow {
		sheet.StartRow = row
		isEdited = true
	}
	if endRow < row {
		sheet.EndRow = row
		isEdited = true
	}
	if isEdited {
		s.m[sheetName] = sheet
	}
	return nil
}
func (s *syncSheetMap) UpdateDimension(file *excelize.File, sheetName string) error {
	s.mx.RLock()
	defer s.mx.RUnlock()
	sheet, ok := s.m[sheetName]
	if !ok {
		return errors.New("ExtendDimension: table not exist")
	}
	err := file.SetSheetDimension(sheetName, sheet.StartColumn+strconv.Itoa(sheet.StartRow)+
		":"+sheet.EndColumn+strconv.Itoa(sheet.EndRow))
	return err
}
