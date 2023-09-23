package formattererror

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestLocateContractError(t *testing.T) {
	type Results struct {
		Location int `json:"location"`
		Row      int `json:"row"`
		ColStart int `json:"colStart"`
		ColEnd   int `json:"colEnd"`
	}

	contracts := []string{
		"KT1X5bt2nmHJfAMHX9vzjNDE5sH7sP5YuYfF",
		"KT1CaqojNPNJPKZGjLk7SSZWpeSovyimBWsx",
		"KT1ERaFkZ8F8xUnPtziVqyk67sDHYpebnvmp",
		"KT1PbQ8FAk5RHtkAyuU49ddbm1nyVoxm8J1E",
		"KT1JNJzUQPME5VccHhXcCZZaDH8HvcxFwqsj",
	}

	for _, c := range contracts {
		t.Run(c, func(t *testing.T) {
			dirpath := fmt.Sprintf("formatter_error_tests/%v/", c)

			data, err := os.ReadFile(dirpath + "code.json")
			require.NoError(t, err)
			require.True(t, gjson.Valid(string(data)))

			results, err := os.ReadFile(dirpath + "results.json")
			require.NoError(t, err)

			var res Results
			err = json.Unmarshal(results, &res)
			require.NoError(t, err)

			parsedData := gjson.ParseBytes(data)

			row, start, end, err := LocateContractError(parsedData, res.Location)
			require.NoError(t, err)

			require.Equal(t, res.Row, row)
			require.Equal(t, res.ColStart, start)
			require.Equal(t, res.ColEnd, end)
		})
	}

}
