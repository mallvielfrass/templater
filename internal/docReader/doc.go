package docReader

import (
	"bytes"
	"io"
	"sync"

	docx "github.com/lukasjarosch/go-docx"
)

type Doc struct {
	doc *docx.Document
	m   sync.Mutex
}

func ReadBytes(data []byte) (*Doc, error) {
	//copy data
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	doc, err := docx.OpenBytes(dataCopy)
	if err != nil {
		return nil, err
	}
	return &Doc{
		doc: doc,
	}, nil
}
func ReadFile(filename string) (*Doc, error) {
	doc, err := docx.Open(filename)
	if err != nil {
		return nil, err
	}
	return &Doc{
		doc: doc,
	}, nil
}

func (d *Doc) Replace(placeholder, value string) error {
	d.m.Lock()
	defer d.m.Unlock()
	return d.doc.Replace(placeholder, value)
}
func (d *Doc) ReplaceAll(placeholders map[string]string) error {
	d.m.Lock()
	defer d.m.Unlock()
	placeholderMap := make(docx.PlaceholderMap)
	for key, value := range placeholders {
		placeholderMap["{"+key+"}"] = value
	}
	return d.doc.ReplaceAll(placeholderMap)
}

func (d *Doc) WriteToFile(filename string) error {
	d.m.Lock()
	defer d.m.Unlock()
	return d.doc.WriteToFile(filename)
}

func (d *Doc) WriteToWriter(writer io.Writer) error {
	d.m.Lock()
	defer d.m.Unlock()
	return d.doc.Write(writer)
}

func (d *Doc) WriteToBytes() ([]byte, error) {
	d.m.Lock()
	defer d.m.Unlock()
	var buf bytes.Buffer
	err := d.doc.Write(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
