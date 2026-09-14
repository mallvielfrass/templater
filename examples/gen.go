//go:build ignore

package main

import (
	"os"
	"path/filepath"

	"github.com/mallvielfrass/templater/internal/sampledata"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	xlsx, err := sampledata.XLSXBytes(8)
	if err != nil {
		panic(err)
	}
	docx, err := sampledata.DOCXBytes()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "examples/sample.xlsx"), xlsx, 0644); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "examples/sample.docx"), docx, 0644); err != nil {
		panic(err)
	}
}
