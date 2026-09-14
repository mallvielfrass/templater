package exdocconverter

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/sampledata"
	"github.com/stretchr/testify/require"
)

type memDocs struct {
	files map[string][]byte
}

func (m *memDocs) SaveDoc(docBytes []byte) (string, error) {
	sum := sha256.Sum256(docBytes)
	hash := hex.EncodeToString(sum[:])
	if m.files == nil {
		m.files = map[string][]byte{}
	}
	m.files[hash] = docBytes
	return hash, nil
}

func TestConvertFillsPlaceholders(t *testing.T) {
	xlsx, err := sampledata.XLSXBytes(2)
	require.NoError(t, err)
	docx, err := sampledata.DOCXBytes()
	require.NoError(t, err)
	book, err := exelreader.ReadBuffer("sample.xlsx", xlsx)
	require.NoError(t, err)
	store := &memDocs{}
	conv, err := NewExDocConverter(store, &book, docx)
	require.NoError(t, err)
	opts, err := conv.CreateConvertOptions(sampledata.SheetName, true)
	require.NoError(t, err)
	hashes, err := conv.Convert(opts, 2, 3)
	require.NoError(t, err)
	require.Len(t, hashes, 2)

	data := store.files[hashes[0]]
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var xml string
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		rc.Close()
		require.NoError(t, err)
		xml = string(b)
	}
	require.Contains(t, xml, "Demo-1")
	require.Contains(t, xml, "item-1.example")
	require.Contains(t, xml, "2020")
}

func TestConvertInvalidTemplate(t *testing.T) {
	xlsx, err := sampledata.XLSXBytes(1)
	require.NoError(t, err)
	book, err := exelreader.ReadBuffer("sample.xlsx", xlsx)
	require.NoError(t, err)
	_, err = NewExDocConverter(&memDocs{}, &book, []byte("not-a-docx"))
	require.Error(t, err)
}
