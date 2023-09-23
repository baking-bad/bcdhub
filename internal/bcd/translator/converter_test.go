package translator

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConverter_FromFile(t *testing.T) {
	files, err := os.ReadDir("./tests/")
	require.NoError(t, err)

	c, err := NewConverter()
	require.NoError(t, err)

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			resultFilename := fmt.Sprintf("tests/%s/code.json", file.Name())
			resultBytes, err := os.ReadFile(resultFilename)
			require.NoError(t, err)

			filename := fmt.Sprintf("tests/%s/code.tz", file.Name())
			got, err := c.FromFile(filename)
			require.NoError(t, err)

			require.JSONEq(t, string(resultBytes), got, "JSON comparing")
		})
	}
}
