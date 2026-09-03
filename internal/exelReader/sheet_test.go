package exelreader

import (
	"testing"

	"github.com/mallvielfrass/templater/internal/utils/cell"
	"github.com/stretchr/testify/assert"
)

func TestColumnNumber(t *testing.T) {
	type test struct {
		column string
		num    int
	}
	tests := []test{
		{
			column: "A",
			num:    1,
		},
		{
			column: "AA",
			num:    27,
		},
		{
			column: "AB",
			num:    28,
		},
		{
			column: "XFD",
			num:    16384,
		},
	}
	for _, test := range tests {

		col := cell.ColumnNumber(test.column)
		assert.Equal(t, test.num, col)
	}
}

func TestSheetInfoWithRows(t *testing.T) {
	file, err := CreateFile()
	assert.NoError(t, err)
	sheetName := "Sheet1"
	_ = file.WriteCell(sheetName, "A1", "col1")
	_ = file.WriteCell(sheetName, "B1", "col2")
	_ = file.WriteCell(sheetName, "A2", "val1")
	_ = file.WriteCell(sheetName, "B2", "val2")
	_ = file.WriteCell(sheetName, "A3", "val3")
	_ = file.WriteCell(sheetName, "B3", "val4")

	info, err := file.SheetInfo(sheetName)
	assert.NoError(t, err)
	assert.Equal(t, 3, info.EndRow)
	assert.Equal(t, "B", info.EndColumn)
}
