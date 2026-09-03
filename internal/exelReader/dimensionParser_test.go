package exelreader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ParseDimensionTestSuite struct {
	suite.Suite
}

// Валидные тесты
func (s *ParseDimensionTestSuite) TestValidDimension() {
	sheet, err := parseDimension("A1:B2")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "A", sheet.StartColumn)
	assert.Equal(s.T(), "B", sheet.EndColumn)
	assert.Equal(s.T(), 1, sheet.StartRow)
	assert.Equal(s.T(), 2, sheet.EndRow)
}

func (s *ParseDimensionTestSuite) TestSingleCellRange() {
	sheet, err := parseDimension("C5:C5")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "C", sheet.StartColumn)
	assert.Equal(s.T(), "C", sheet.EndColumn)
	assert.Equal(s.T(), 5, sheet.StartRow)
	assert.Equal(s.T(), 5, sheet.EndRow)
}

func (s *ParseDimensionTestSuite) TestSingleCellRange2() {
	sheet, err := parseDimension("C5:D8")
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "C", sheet.StartColumn)
	assert.Equal(s.T(), "D", sheet.EndColumn)
	assert.Equal(s.T(), 5, sheet.StartRow)
	assert.Equal(s.T(), 8, sheet.EndRow)
}
func (s *ParseDimensionTestSuite) TestLongColumnNames() {
	sheet, err := parseDimension("XYZ100:ABC200")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "XYZ", sheet.StartColumn)
	assert.Equal(s.T(), "ABC", sheet.EndColumn)
	assert.Equal(s.T(), 100, sheet.StartRow)
	assert.Equal(s.T(), 200, sheet.EndRow)
}

// Невалидные тесты
func (s *ParseDimensionTestSuite) TestEmptyDimension() {
	sheet, err := parseDimension("")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "A", sheet.StartColumn)
	assert.Equal(s.T(), "A", sheet.EndColumn)
	assert.Equal(s.T(), 1, sheet.StartRow)
	assert.Equal(s.T(), 1, sheet.EndRow)
}

func (s *ParseDimensionTestSuite) TestSingleCellNoColon() {
	sheet, err := parseDimension("C5")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "C", sheet.StartColumn)
	assert.Equal(s.T(), "C", sheet.EndColumn)
	assert.Equal(s.T(), 5, sheet.StartRow)
	assert.Equal(s.T(), 5, sheet.EndRow)
}

func (s *ParseDimensionTestSuite) TestNoColon() {
	_, err := parseDimension("A1B2")
	require.Error(s.T(), err)
	assert.Equal(s.T(), "dimension is not valid", err.Error())
}

func (s *ParseDimensionTestSuite) TestTooManyParts() {
	_, err := parseDimension("A1:B2:C3")
	require.Error(s.T(), err)
	assert.Equal(s.T(), "dimension is not valid", err.Error())
}

func (s *ParseDimensionTestSuite) TestInvalidStartCell() {
	// Предположим, что "A0" — недопустимая ячейка (строка не может быть 0)
	_, err := parseDimension("A0:B2")
	require.Error(s.T(), err)
	assert.Equal(s.T(), "dimension is not valid", err.Error())
}

func (s *ParseDimensionTestSuite) TestInvalidEndCell() {
	_, err := parseDimension("A1:XYZ")
	require.Error(s.T(), err)
	assert.Equal(s.T(), "dimension is not valid", err.Error())
}

// Граничные случаи
func (s *ParseDimensionTestSuite) TestMaxValues() {
	sheet, err := parseDimension("XFD1048576:XFD1048576")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "XFD", sheet.StartColumn)
	assert.Equal(s.T(), "XFD", sheet.EndColumn)
	assert.Equal(s.T(), 1048576, sheet.StartRow)
	assert.Equal(s.T(), 1048576, sheet.EndRow)
}

func TestParseDimensionSuite(t *testing.T) {
	suite.Run(t, new(ParseDimensionTestSuite))
}
