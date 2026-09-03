package docReader

import (
	"bytes"

	"github.com/gomutex/godocx"
)

func EmptyBytes() ([]byte, error) {
	doc, err := godocx.NewDocument()
	if err != nil {
		return nil, err
	}
	doc.AddParagraph("")
	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
