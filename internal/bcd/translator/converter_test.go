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
// the input from its start, so parsing was quadratic: these inputs took about
// 47 seconds.
func TestConverter_FromString_LargeInputIsFast(t *testing.T) {
	c, err := NewConverter()
	require.NoError(t, err)

	for _, input := range []string{
		"{" + strings.Repeat(" UNIT ;", 60000) + " UNIT }",
		"Pair" + strings.Repeat(` "a"`, 60000),
	} {
		start := time.Now()
		_, err := c.FromString(input)
		require.NoError(t, err)
		require.Less(t, time.Since(start), perfTestLimit)
	}
}

// perfTestLimit bounds performance regression tests. It is generous on
// purpose: the linear code takes a fraction of it even on a slow runner, while
// the quadratic code these tests guard against takes tens of seconds.
const perfTestLimit = 5 * time.Second
