package formattererror

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

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

			data, err := ioutil.ReadFile(dirpath + "code.json")
			if err != nil {
				t.Errorf("error in ioutil.ReadFile(%v%v): %v", dirpath, "code.json", err)
			}

			if !gjson.Valid(string(data)) {
				t.Error("invalid json")
			}

			results, err := ioutil.ReadFile(dirpath + "results.json")
			if err != nil {
				t.Errorf("error in ioutil.ReadFile(%v%v): %v", dirpath, "results.json", err)
			}

			var res Results
			err = json.Unmarshal(results, &res)
			if err != nil {
				t.Error("cant unmarshal results.json file")
			}

			parsedData := gjson.ParseBytes(data)

			row, start, end, err := LocateContractError(parsedData, res.Location)
			if err != nil {
				t.Errorf("err in LocateContractError: %v", err)
			}

			if row != res.Row {
				t.Errorf("wrong row. got %v, expected %v", row, res.Row)
			}

			if start != res.ColStart {
				t.Errorf("wrong start. got %v, expected %v", start, res.ColStart)
			}

			if end != res.ColEnd {
				t.Errorf("wrong end. got %v, expected %v", end, res.ColEnd)
			}

		})
	}

}
