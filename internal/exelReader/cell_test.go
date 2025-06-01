package exelreader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ParseCellTestSuite struct {
	suite.Suite
}

// Валидные тесты
func (s *ParseCellTestSuite) TestSimpleCell() {
	col, row, err := parseCell("A1")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "A", col)
	assert.Equal(s.T(), 1, row)
}

func (s *ParseCellTestSuite) TestTwoLetters() {
	col, row, err := parseCell("AB123")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "AB", col)
	assert.Equal(s.T(), 123, row)
}

func (s *ParseCellTestSuite) TestUpperCase() {
	col, row, err := parseCell("aB5")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "AB", col) // Предполагается, что столбец в верхнем регистре
	assert.Equal(s.T(), 5, row)
}

func (s *ParseCellTestSuite) TestLongColumnName() {
	col, row, err := parseCell("XYZ100")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "XYZ", col)
	assert.Equal(s.T(), 100, row)
}

func (s *ParseCellTestSuite) TestMaxRowNumber() {
	col, row, err := parseCell("XFD1048576")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "XFD", col)
	assert.Equal(s.T(), 1048576, row)
}

// Невалидные тесты
func (s *ParseCellTestSuite) TestInvalidFormatNoDigits() {
	_, _, err := parseCell("ABC")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный формат ячейки")
}

func (s *ParseCellTestSuite) TestInvalidFormatNoLetters() {
	_, _, err := parseCell("123")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный формат ячейки")
}

func (s *ParseCellTestSuite) TestInvalidRowNumber() {
	_, _, err := parseCell("A0")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный номер строки")
}

func (s *ParseCellTestSuite) TestInvalidCharacters() {
	_, _, err := parseCell("!@#")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный формат ячейки")
}

func (s *ParseCellTestSuite) TestInvalidOrder() {
	_, _, err := parseCell("1A2")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный формат ячейки")
}

// Граничные случаи
func (s *ParseCellTestSuite) TestMinRowNumber() {
	col, row, err := parseCell("A1")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "A", col)
	assert.Equal(s.T(), 1, row)
}

func (s *ParseCellTestSuite) TestZeroRow() {
	_, _, err := parseCell("A0")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный номер строки")
}

func (s *ParseCellTestSuite) TestSingleCharacter() {
	_, _, err := parseCell("A")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "некорректный формат ячейки")
}

func TestParseCellSuite(t *testing.T) {
	suite.Run(t, new(ParseCellTestSuite))
}
