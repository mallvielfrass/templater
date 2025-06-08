package docReader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindAndReplace(t *testing.T) {
	doc, err := CreateMockDoc()
	require.NoError(t, err)

	err = doc.Replace("{test}", "toster")
	require.NoError(t, err)

	doc.WriteToFile("test.docx")
}
