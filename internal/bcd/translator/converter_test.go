package translator

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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

// go-peg's EnableAst computed line/column info for every AST node by scanning
// the input from its start, so parsing was quadratic: 20k instructions took
// about 5 seconds.
func TestConverter_FromString_LargeInputIsFast(t *testing.T) {
	c, err := NewConverter()
	require.NoError(t, err)

	for _, input := range []string{
		"{" + strings.Repeat(" UNIT ;", 20000) + " UNIT }",
		"Pair" + strings.Repeat(` "a"`, 20000),
	} {
		start := time.Now()
		_, err := c.FromString(input)
		require.NoError(t, err)
		require.Less(t, time.Since(start), time.Second)
	}
}
