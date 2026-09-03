package docReader

import (
	"os"

	"github.com/gomutex/godocx"
	docx "github.com/lukasjarosch/go-docx"
)

func CreateMockDoc() (*docx.Document, error) {
	tempDocument, err := godocx.NewDocument()
	if err != nil {
		return nil, err
	}

	tempDocument.AddParagraph("{test}")

	//save to tmp
	tmpFile, err := os.CreateTemp("", "test.docx")
	if err != nil {
		return nil, err
	}
	tempDocument.WriteTo(tmpFile)

	document, err := docx.Open(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	return document, nil
}
