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
