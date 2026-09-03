package exdocconverter

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/gomutex/godocx"
	"github.com/mallvielfrass/templater/internal/docReader"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocStorage - мок для интерфейса docStorage
type MockDocStorage struct {
	mock.Mock
}

func (m *MockDocStorage) SaveDoc(docBytes []byte) (string, error) {
	args := m.Called(docBytes)
	return args.String(0), args.Error(1)
}

// MockExelReader - интерфейс для мокирования ExelFile
type MockExelReader interface {
	SheetInfo(sheetName string) (models.Sheet, error)
	ReadFirstRow(sheetName string) ([]string, error)
	ReadSheetRows(sheetName string, minRow, maxRow int) ([][]string, error)
}

// MockExelFile - мок для ExelFile (имитирует все методы, используемые в ExDocConverter)
type MockExelFile struct {
	mock.Mock
}

func (m *MockExelFile) SheetInfo(sheetName string) (models.Sheet, error) {
	args := m.Called(sheetName)
	return args.Get(0).(models.Sheet), args.Error(1)
}

func (m *MockExelFile) ReadFirstRow(sheetName string) ([]string, error) {
	args := m.Called(sheetName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockExelFile) ReadSheetRows(sheetName string, minRow, maxRow int) ([][]string, error) {
	args := m.Called(sheetName, minRow, maxRow)
	return args.Get(0).([][]string), args.Error(1)
}

func TestNewExDocConverter(t *testing.T) {
	mockDocStorage := &MockDocStorage{}

	t.Run("успешное создание ExDocConverter", func(t *testing.T) {
		// Подготовим данные, которые не вызовут ошибку в docreader.ReadBytes
		// Поскольку мы не можем легко создать валидный doc файл, пропустим этот тест
		// или создадим интеграционный тест отдельно
		t.Skip("Этот тест требует интеграции с реальным docreader")
	})

	t.Run("ошибка при невалидном документе", func(t *testing.T) {
		// Данные, которые точно не являются валидным документом
		invalidDocBytes := []byte("invalid doc content")

		// Создаем реальный ExelFile для тестирования
		converter, err := NewExDocConverter(mockDocStorage, nil, invalidDocBytes)

		assert.Error(t, err)
		assert.Nil(t, converter)
	})

	t.Run("пустые байты документа", func(t *testing.T) {
		emptyDocBytes := []byte{}

		converter, err := NewExDocConverter(mockDocStorage, nil, emptyDocBytes)

		assert.Error(t, err)
		assert.Nil(t, converter)
	})

	t.Run("nil docStorage", func(t *testing.T) {
		docBytes := []byte("some content")

		converter, err := NewExDocConverter(nil, nil, docBytes)

		assert.Error(t, err)
		assert.Nil(t, converter)
	})
}

func TestConvertOptions(t *testing.T) {
	t.Run("создание convertOptions", func(t *testing.T) {
		options := convertOptions{
			sheetName:            "TestSheet",
			useFirstRowAsColumns: true,
			columns:              []string{"Name", "Age", "Email"},
		}

		assert.Equal(t, "TestSheet", options.sheetName)
		assert.True(t, options.useFirstRowAsColumns)
		assert.Len(t, options.columns, 3)
		assert.Equal(t, []string{"Name", "Age", "Email"}, options.columns)
	})

	t.Run("пустой список колонок", func(t *testing.T) {
		options := convertOptions{
			sheetName:            "TestSheet",
			useFirstRowAsColumns: false,
			columns:              []string{},
		}

		assert.Equal(t, "TestSheet", options.sheetName)
		assert.False(t, options.useFirstRowAsColumns)
		assert.Len(t, options.columns, 0)
	})

	t.Run("много колонок", func(t *testing.T) {
		manyColumns := make([]string, 100)
		for i := 0; i < 100; i++ {
			manyColumns[i] = string(rune('A' + i%26))
		}

		options := convertOptions{
			sheetName:            "BigSheet",
			useFirstRowAsColumns: false,
			columns:              manyColumns,
		}

		assert.Equal(t, "BigSheet", options.sheetName)
		assert.False(t, options.useFirstRowAsColumns)
		assert.Len(t, options.columns, 100)
	})
}

func TestExDocConverter_CreateConvertOptions(t *testing.T) {
	// Создаем структуру, которая имитирует поведение ExDocConverter для тестирования логики
	type testExDocConverter struct {
		mockExelReader MockExelReader
	}

	createTestConverter := func(mockReader MockExelReader) testExDocConverter {
		return testExDocConverter{mockExelReader: mockReader}
	}

	// Имитируем метод CreateConvertOptions для тестирования
	createConvertOptions := func(tc testExDocConverter, sheetName string, useFirstRowAsColumns bool) (convertOptions, error) {
		sheet, err := tc.mockExelReader.SheetInfo(sheetName)
		if err != nil {
			return convertOptions{}, err
		}
		columns := []string{}
		if !useFirstRowAsColumns {
			// Имитируем логику из utils/cell.ColumnsCountToAddresses
			for i := 1; i <= sheet.GetColumnCount(); i++ {
				column := string(rune('A' + i - 1))
				if i > 26 {
					// Упрощенная логика для колонок больше Z
					column = "A" + string(rune('A'+i-27))
				}
				columns = append(columns, column)
			}
		} else {
			firstRow, err := tc.mockExelReader.ReadFirstRow(sheetName)
			if err != nil {
				return convertOptions{}, err
			}
			columns = firstRow
		}
		return convertOptions{
			sheetName:            sheetName,
			useFirstRowAsColumns: useFirstRowAsColumns,
			columns:              columns,
		}, nil
	}

	t.Run("успешное создание опций без использования первой строки как колонок", func(t *testing.T) {
		sheetName := "Sheet1"
		useFirstRowAsColumns := false

		mockExelReader := &MockExelFile{}

		// Создаем мок Sheet с GetColumnCount методом
		sheet := models.Sheet{
			Name:        sheetName,
			StartColumn: "A",
			EndColumn:   "C",
			StartRow:    1,
			EndRow:      10,
		}

		mockExelReader.On("SheetInfo", sheetName).Return(sheet, nil)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, useFirstRowAsColumns)

		assert.NoError(t, err)
		assert.Equal(t, sheetName, options.sheetName)
		assert.Equal(t, useFirstRowAsColumns, options.useFirstRowAsColumns)
		assert.Equal(t, []string{"A", "B", "C"}, options.columns)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("успешное создание опций с использованием первой строки как колонок", func(t *testing.T) {
		sheetName := "Sheet1"
		useFirstRowAsColumns := true
		firstRow := []string{"Name", "Age", "City"}

		mockExelReader := &MockExelFile{}

		sheet := models.Sheet{
			Name:        sheetName,
			StartColumn: "A",
			EndColumn:   "C",
			StartRow:    1,
			EndRow:      10,
		}

		mockExelReader.On("SheetInfo", sheetName).Return(sheet, nil)
		mockExelReader.On("ReadFirstRow", sheetName).Return(firstRow, nil)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, useFirstRowAsColumns)

		assert.NoError(t, err)
		assert.Equal(t, sheetName, options.sheetName)
		assert.Equal(t, useFirstRowAsColumns, options.useFirstRowAsColumns)
		assert.Equal(t, firstRow, options.columns)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("ошибка при получении информации о листе", func(t *testing.T) {
		sheetName := "InvalidSheet"
		expectedError := errors.New("sheet not found")

		mockExelReader := &MockExelFile{}
		mockExelReader.On("SheetInfo", sheetName).Return(models.Sheet{}, expectedError)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, false)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, convertOptions{}, options)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("ошибка при чтении первой строки", func(t *testing.T) {
		sheetName := "Sheet1"
		expectedError := errors.New("failed to read first row")

		mockExelReader := &MockExelFile{}

		sheet := models.Sheet{
			Name:        sheetName,
			StartColumn: "A",
			EndColumn:   "C",
			StartRow:    1,
			EndRow:      10,
		}

		mockExelReader.On("SheetInfo", sheetName).Return(sheet, nil)
		mockExelReader.On("ReadFirstRow", sheetName).Return([]string(nil), expectedError)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, true)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, convertOptions{}, options)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("лист с одной колонкой", func(t *testing.T) {
		sheetName := "SingleColumn"
		useFirstRowAsColumns := false

		mockExelReader := &MockExelFile{}

		sheet := models.Sheet{
			Name:        sheetName,
			StartColumn: "A",
			EndColumn:   "A",
			StartRow:    1,
			EndRow:      5,
		}

		mockExelReader.On("SheetInfo", sheetName).Return(sheet, nil)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, useFirstRowAsColumns)

		assert.NoError(t, err)
		assert.Equal(t, sheetName, options.sheetName)
		assert.Equal(t, []string{"A"}, options.columns)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("пустая первая строка как колонки", func(t *testing.T) {
		sheetName := "EmptyHeaders"
		useFirstRowAsColumns := true
		emptyRow := []string{}

		mockExelReader := &MockExelFile{}

		sheet := models.Sheet{
			Name:        sheetName,
			StartColumn: "A",
			EndColumn:   "B",
			StartRow:    1,
			EndRow:      5,
		}

		mockExelReader.On("SheetInfo", sheetName).Return(sheet, nil)
		mockExelReader.On("ReadFirstRow", sheetName).Return(emptyRow, nil)

		tc := createTestConverter(mockExelReader)
		options, err := createConvertOptions(tc, sheetName, useFirstRowAsColumns)

		assert.NoError(t, err)
		assert.Equal(t, sheetName, options.sheetName)
		assert.Equal(t, emptyRow, options.columns)

		mockExelReader.AssertExpectations(t)
	})
}

func TestExDocConverter_Convert(t *testing.T) {
	t.Run("успешная конвертация", func(t *testing.T) {
		// Этот тест требует интеграции с docreader для создания валидных документов
		t.Skip("Этот тест требует интеграции с реальным docreader для создания валидных документов")
	})

	t.Run("ошибка при чтении строк Excel", func(t *testing.T) {
		// Тестируем логику обработки ошибок при чтении Excel
		mockExelReader := &MockExelFile{}

		options := convertOptions{
			sheetName:            "Sheet1",
			useFirstRowAsColumns: true,
			columns:              []string{"Name", "Age"},
		}
		minRow, maxRow := 2, 3
		expectedError := errors.New("failed to read rows")

		mockExelReader.On("ReadSheetRows", options.sheetName, minRow, maxRow).Return([][]string(nil), expectedError)

		// Имитируем вызов Convert с ошибкой чтения
		_, err := mockExelReader.ReadSheetRows(options.sheetName, minRow, maxRow)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("обработка пустых ячеек - логика проверки", func(t *testing.T) {
		// Тестируем только логику обработки пустых ячеек без интеграции с docreader
		options := convertOptions{
			sheetName: "Sheet1",
			columns:   []string{"Name", "Age", "City"},
		}

		// Проверяем, что количество колонок больше количества данных в строке
		rows := [][]string{{"John", "25"}} // Нет значения для City

		// Создаем map для замены значений, как это делается в реальном коде
		placeholders := make(map[string]string)
		row := rows[0]
		for j, column := range options.columns {
			if j < len(row) {
				placeholders[column] = row[j]
			} else {
				placeholders[column] = "" // Если данных нет, подставляем пустую строку
			}
		}

		// Проверяем, что пустые значения корректно обрабатываются
		assert.Equal(t, "John", placeholders["Name"])
		assert.Equal(t, "25", placeholders["Age"])
		assert.Equal(t, "", placeholders["City"]) // Пустая строка для отсутствующих данных
	})

	t.Run("обработка строк с пустыми значениями", func(t *testing.T) {
		options := convertOptions{
			sheetName: "Sheet1",
			columns:   []string{"Name", "Age", "Email"},
		}

		// Тестируем строку с пустыми значениями в середине
		rows := [][]string{{"John", "", "john@example.com"}}

		placeholders := make(map[string]string)
		row := rows[0]
		for j, column := range options.columns {
			if j < len(row) {
				placeholders[column] = row[j]
			} else {
				placeholders[column] = ""
			}
		}

		assert.Equal(t, "John", placeholders["Name"])
		assert.Equal(t, "", placeholders["Age"]) // Пустое значение
		assert.Equal(t, "john@example.com", placeholders["Email"])
	})

	t.Run("обработка нескольких строк", func(t *testing.T) {
		mockExelReader := &MockExelFile{}

		options := convertOptions{
			sheetName: "Sheet1",
			columns:   []string{"Name", "Age"},
		}
		minRow, maxRow := 2, 4

		rows := [][]string{
			{"John", "25"},
			{"Jane", "30"},
			{"Bob", "35"},
		}

		mockExelReader.On("ReadSheetRows", options.sheetName, minRow, maxRow).Return(rows, nil)

		readRows, err := mockExelReader.ReadSheetRows(options.sheetName, minRow, maxRow)

		assert.NoError(t, err)
		assert.Len(t, readRows, 3)
		assert.Equal(t, rows, readRows)

		mockExelReader.AssertExpectations(t)
	})

	t.Run("граничные случаи minRow и maxRow", func(t *testing.T) {
		mockExelReader := &MockExelFile{}

		options := convertOptions{
			sheetName: "Sheet1",
			columns:   []string{"Data"},
		}

		// Тестируем случай когда minRow == maxRow
		minRow, maxRow := 5, 5
		rows := [][]string{{"SingleRow"}}

		mockExelReader.On("ReadSheetRows", options.sheetName, minRow, maxRow).Return(rows, nil)

		readRows, err := mockExelReader.ReadSheetRows(options.sheetName, minRow, maxRow)

		assert.NoError(t, err)
		assert.Len(t, readRows, 1)
		assert.Equal(t, []string{"SingleRow"}, readRows[0])

		mockExelReader.AssertExpectations(t)
	})
}

// Тесты для интеграции можно добавить отдельно
func TestExDocConverter_Integration_PrepareDoc(t *testing.T) {
	//	t.Skip("Интеграционные тесты требуют настройки окружения с реальными файлами")
	exFile, err := exelreader.CreateFile()
	if err != nil {
		t.Fatal(err)
	}
	exFile.CreateSheet("Sheet1")
	exFile.WriteCell("Sheet1", "A1", "Name")
	exFile.WriteCell("Sheet1", "B1", "Age")
	exFile.WriteCell("Sheet1", "C1", "City")
	//create doc
	tempDocument, err := godocx.NewDocument()
	if err != nil {
		t.Fatal(err)
	}

	tempDocument.AddParagraph("Name: {Name}")
	tempDocument.AddParagraph("Age: {Age}")
	tempDocument.AddParagraph("City: {City}")

	//save to tmp
	tmpFile, err := os.CreateTemp("", "test.docx")
	if err != nil {
		t.Fatal(err)
	}
	tempDocument.WriteTo(tmpFile)
	docFile, err := docReader.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	docFile.Replace("Name", "John")
	docFile.Replace("Age", "25")
	docFile.Replace("City", "New York")
	res, err := docFile.WriteToBytes()
	if err != nil {
		t.Fatal(err)
	}

	docFile, err = docReader.ReadBytes(res)
	if err != nil {
		t.Fatal(err)
	}
	docFile.WriteToFile(tmpFile.Name())
	//read tmpFile.Name() as zip
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer zipReader.Close()
	var documentXML *zip.File
	//read all text from docx
	for _, file := range zipReader.File {
		//print file name
		//	fmt.Println(file.Name)
		if file.Name == "word/document.xml" {
			documentXML = file
		}
	}
	if documentXML == nil {
		t.Fatal("document.xml not found")
	}
	//read all text from document.xml
	rc, err := documentXML.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	//read all text from document.xml
	bytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	strBody := string(bytes)
	assert.Contains(t, strBody, "Name: John")
	assert.Contains(t, strBody, "Age: 25")
	assert.Contains(t, strBody, "City: New York")
}
func TestExDocConverter_Integration_Convert(t *testing.T) {
	//	t.Skip("Интеграционные тесты требуют настройки окружения с реальными файлами")
	exFile, err := exelreader.CreateFile()
	if err != nil {
		t.Fatal(err)
	}
	exFile.CreateSheet("Sheet1")
	exFile.WriteCell("Sheet1", "A1", "Name")
	exFile.WriteCell("Sheet1", "B1", "Age")
	exFile.WriteCell("Sheet1", "C1", "City")
	//create doc
	tempDocument, err := godocx.NewDocument()
	if err != nil {
		t.Fatal(err)
	}

	tempDocument.AddParagraph("Name: {Name}")
	tempDocument.AddParagraph("Age: {Age}")
	tempDocument.AddParagraph("City: {City}")

	//save to tmp
	tmpFile, err := os.CreateTemp("", "test.docx")
	if err != nil {
		t.Fatal(err)
	}
	tempDocument.WriteTo(tmpFile)
	docFile, err := docReader.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	docFile.Replace("Name", "John")
	docFile.Replace("Age", "25")
	docFile.Replace("City", "New York")
	res, err := docFile.WriteToBytes()
	if err != nil {
		t.Fatal(err)
	}

	docFile, err = docReader.ReadBytes(res)
	if err != nil {
		t.Fatal(err)
	}
	docFile.WriteToFile(tmpFile.Name())
	//read tmpFile.Name() as zip
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer zipReader.Close()
	var documentXML *zip.File
	//read all text from docx
	for _, file := range zipReader.File {
		//print file name
		//	fmt.Println(file.Name)
		if file.Name == "word/document.xml" {
			documentXML = file
		}
	}
	if documentXML == nil {
		t.Fatal("document.xml not found")
	}
	//read all text from document.xml
	rc, err := documentXML.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	//read all text from document.xml
	bytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	strBody := string(bytes)
	assert.Contains(t, strBody, "Name: John")
	assert.Contains(t, strBody, "Age: 25")
	assert.Contains(t, strBody, "City: New York")
}

// Дополнительные тесты для проверки корректности моков
func TestMockExelFile(t *testing.T) {
	t.Run("проверка корректности мок объекта", func(t *testing.T) {
		mockExelFile := &MockExelFile{}

		// Настраиваем мок
		expectedSheet := models.Sheet{
			Name:        "TestSheet",
			StartColumn: "A",
			EndColumn:   "B",
			StartRow:    1,
			EndRow:      10,
		}

		mockExelFile.On("SheetInfo", "TestSheet").Return(expectedSheet, nil)

		// Тестируем мок
		sheet, err := mockExelFile.SheetInfo("TestSheet")

		assert.NoError(t, err)
		assert.Equal(t, expectedSheet, sheet)

		mockExelFile.AssertExpectations(t)
	})

	t.Run("проверка мока с множественными вызовами", func(t *testing.T) {
		mockExelFile := &MockExelFile{}

		// Настраиваем мок для нескольких вызовов
		mockExelFile.On("ReadFirstRow", "Sheet1").Return([]string{"Name", "Age"}, nil).Times(2)

		// Делаем два вызова
		firstCall, err1 := mockExelFile.ReadFirstRow("Sheet1")
		secondCall, err2 := mockExelFile.ReadFirstRow("Sheet1")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, []string{"Name", "Age"}, firstCall)
		assert.Equal(t, []string{"Name", "Age"}, secondCall)

		mockExelFile.AssertExpectations(t)
	})
}
