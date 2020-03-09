package cerrors

import (
	"io/ioutil"
	"os"

	"github.com/tidwall/gjson"
)

var errorDescriptions gjson.Result

// LoadErrorDescriptions -
func LoadErrorDescriptions(filePath string) (err error) {
	f, err := os.Open(filePath)
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
