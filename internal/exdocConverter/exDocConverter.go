package exdocconverter

import (
	docreader "github.com/mallvielfrass/templater/internal/docReader"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/utils/cell"
)

type ExDocConverter struct {
	docStorage docStorage
	exelReader *exelreader.ExelFile
	docBytes   []byte
}
type docStorage interface {
	SaveDoc(docBytes []byte) (hash string, err error)
	//GetDoc(hash string) (docBytes []byte, err error)
}

func NewExDocConverter(docStorage docStorage, exelReader *exelreader.ExelFile, docBytes []byte) (*ExDocConverter, error) {
	//read doc file for validation doc file
	_, err := docreader.ReadBytes(docBytes)
	if err != nil {
		return nil, err
	}
	return &ExDocConverter{docStorage: docStorage, exelReader: exelReader, docBytes: docBytes}, nil
}

type convertOptions struct {
	sheetName            string
	useFirstRowAsColumns bool
	columns              []string
}

func (e *ExDocConverter) CreateConvertOptions(sheetName string, useFirstRowAsColumns bool) (convertOptions, error) {
	sheet, err := e.exelReader.SheetInfo(sheetName)
	if err != nil {
		return convertOptions{}, err
	}
	columns := []string{}
	if !useFirstRowAsColumns {
		columns = cell.ColumnsCountToAddresses(sheet.GetColumnCount())
	} else {
		firstRow, err := e.exelReader.ReadFirstRow(sheetName)
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
func (e *ExDocConverter) Convert(options convertOptions, minRow int, maxRow int) ([]string, error) {
	docHashes := make([]string, maxRow-minRow+1)

	// Читаем строки из Excel файла
	rows, err := e.exelReader.ReadSheetRows(options.sheetName, minRow, maxRow)
	if err != nil {
		return nil, err
	}

	// Для каждой строки создаем документ
	for i, row := range rows {
		// Создаем копию шаблона документа
		doc, err := docreader.ReadBytes(e.docBytes)
		if err != nil {
			return nil, err
		}

		// Создаем map для замены значений
		placeholders := make(map[string]string)
		for j, column := range options.columns {
			if j < len(row) {
				placeholders[column] = row[j]
			} else {
				placeholders[column] = "" // Если данных нет, подставляем пустую строку
			}
		}

		// Заменяем все плейсхолдеры в документе
		err = doc.ReplaceAll(placeholders)
		if err != nil {
			return nil, err
		}

		// Получаем байты документа
		docBytes, err := doc.WriteToBytes()
		if err != nil {
			return nil, err
		}

		// Сохраняем документ в storage
		hash, err := e.docStorage.SaveDoc(docBytes)
		if err != nil {
			return nil, err
		}

		// Добавляем хеш в результат
		docHashes[i] = hash
	}

	return docHashes, nil
}
