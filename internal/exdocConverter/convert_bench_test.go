package exdocconverter

import (
	"fmt"
	"testing"

	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/sampledata"
)

type discardDocs struct {
	n int
}

func (d *discardDocs) SaveDoc([]byte) (string, error) {
	d.n++
	return fmt.Sprintf("%d", d.n), nil
}

func BenchmarkConvert(b *testing.B) {
	docx, err := sampledata.DOCXBytes()
	if err != nil {
		b.Fatal(err)
	}
	for _, rows := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("%d_rows", rows), func(b *testing.B) {
			xlsx, err := sampledata.XLSXBytes(rows)
			if err != nil {
				b.Fatal(err)
			}
			book, err := exelreader.ReadBuffer("bench.xlsx", xlsx)
			if err != nil {
				b.Fatal(err)
			}
			conv, err := NewExDocConverter(&discardDocs{}, &book, docx)
			if err != nil {
				b.Fatal(err)
			}
			opts, err := conv.CreateConvertOptions(sampledata.SheetName, true)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := conv.Convert(opts, 2, rows+1); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
