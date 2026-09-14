package sampledata

import (
	"bytes"
	"fmt"

	"github.com/gomutex/godocx"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
)

const (
	SheetName = "Sheet1"
	ColSite   = "site"
	ColName   = "name"
	ColYear   = "year"
)

func XLSXBytes(rows int) ([]byte, error) {
	if rows < 1 {
		rows = 1
	}
	book, err := exelreader.CreateFile()
	if err != nil {
		return nil, err
	}
	if err := book.CreateSheet(SheetName); err != nil {
		return nil, err
	}
	if err := book.WriteCell(SheetName, "A1", ColSite); err != nil {
		return nil, err
	}
	if err := book.WriteCell(SheetName, "B1", ColName); err != nil {
		return nil, err
	}
	if err := book.WriteCell(SheetName, "C1", ColYear); err != nil {
		return nil, err
	}
	for i := 0; i < rows; i++ {
		r := i + 2
		if err := book.WriteCell(SheetName, fmt.Sprintf("A%d", r), fmt.Sprintf("item-%d.example", i+1)); err != nil {
			return nil, err
		}
		if err := book.WriteCell(SheetName, fmt.Sprintf("B%d", r), fmt.Sprintf("Demo-%d", i+1)); err != nil {
			return nil, err
		}
		if err := book.WriteCell(SheetName, fmt.Sprintf("C%d", r), 2020+i%6); err != nil {
			return nil, err
		}
	}
	buf, err := book.WriteBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DOCXBytes() ([]byte, error) {
	doc, err := godocx.NewDocument()
	if err != nil {
		return nil, err
	}
	doc.AddParagraph("Certificate for {" + ColName + "}")
	doc.AddParagraph("Site: {" + ColSite + "}")
	doc.AddParagraph("Year: {" + ColYear + "}")
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
