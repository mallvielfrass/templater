package exelreader

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type dimensionSuite struct {
	suite.Suite
	Log func(args ...any)
}

const (
	sheetName = "toster"
)

// func createFile() (*excelize.File, error) {
// 	file := excelize.NewFile()
// 	_, err := file.NewSheet(sheetName)
// 	if err != nil {
// 		return &excelize.File{}, err
// 	}
// 	err = file.DeleteSheet("Sheet1")
// 	if err != nil {
// 		return &excelize.File{}, err
// 	}

//		return file, nil
//	}
func (suite *dimensionSuite) SetupTest() {
	suite.Log = suite.Suite.T().Log

}
func (suite *dimensionSuite) TestDimension() {
	type cell struct {
		Column string
		Row    int
		Value  string
	}
	type sheetTest struct {
		Cell []cell
		//Rows
	}

	tests := []sheetTest{
		{
			Cell: []cell{
				{
					Column: "A",
					Row:    1,
					Value:  "1",
				},
				{
					Column: "B",
					Row:    48,
					Value:  "2",
				},
				{
					Column: "CA",
					Row:    16,
					Value:  "2",
				},
			},
		},
	}
	for _, test := range tests {
		file, err := CreateFile()
		if err != nil {
			suite.Fail("create file:", err)
		}
		file.CreateSheet(sheetName)
		// //fill
		for _, column := range test.Cell {
			suite.Log("set cell value:", column.Column+strconv.Itoa(column.Row), column.Value)
			err = file.WriteCell(sheetName, column.Column+strconv.Itoa(column.Row), column.Value)
			if err != nil {
				suite.Fail("set cell value:", err)
			}
		}
		err = file.UpdateDimension(sheetName)
		if err != nil {
			suite.Fail("save file:", err)
		}
		// proprs, err := file.GetCalcProps()
		// if err != nil {
		// 	suite.Fail("save file:", err)
		// }
		// suite.Log("calc props:", proprs.CalcMode)
		dimension, err := file.GetSheetDimension(sheetName)
		if err != nil {
			suite.Fail("get sheet dimension:", err)
		}
		suite.Equal("A1:CA48", dimension)

	}
}
func (s *dimensionSuite) TearDownTest() {}
func TestDimensionSuite(t *testing.T) {
	suite.Run(t, new(dimensionSuite))
}
