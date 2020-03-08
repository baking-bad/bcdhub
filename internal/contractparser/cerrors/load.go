package cerrors

import (
	"io/ioutil"
	"os"

	"github.com/tidwall/gjson"
)

var errorDescriptions gjson.Result

func loadErrorDescriptions() (err error) {
	f, err := os.Open("errors.json")
	if err != nil {
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	errorDescriptions = gjson.ParseBytes(b)
	return
}
