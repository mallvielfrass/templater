package exelreader

import (
	"errors"
	"strconv"
	"sync"

	"github.com/xuri/excelize/v2"
)

type syncSheetMap struct {
	mx sync.RWMutex
	m  map[string]*sheet
}

func NewSyncSheetMap() *syncSheetMap {
	return &syncSheetMap{
		m: make(map[string]*sheet),
	}
}

// Store save object to map
func (s *syncSheetMap) Store(sheetName string, sheet *sheet) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[sheetName] = sheet
}
func (s *syncSheetMap) Load(sheetName string) (*sheet, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.m[sheetName]
	return val, ok
}

func (s *syncSheetMap) ExtendDimension(sheetName, cell string) error {
	col, row, err := parseCell(cell)
	if err != nil {
		return errors.New("dimension is not valid")
	}
	colNum := columnNumber(col)
	s.mx.Lock()
	defer s.mx.Unlock()
	sheet, ok := s.m[sheetName]
	if !ok {
		return errors.New("ExtendDimension: table not exist")
	}
	startColumn := columnNumber(sheet.StartColumn)
	endColumn := columnNumber(sheet.EndColumn)
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
