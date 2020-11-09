package cerrors

import (
	"io/ioutil"
	"os"
)

var errorDescriptions map[string]Description

// Description -
type Description struct {
	Title       string `json:"title"`
	Description string `json:"descr"`
}

// LoadErrorDescriptions -
func LoadErrorDescriptions(filePath string) (err error) {
	errorDescriptions = make(map[string]Description)

	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	return json.Unmarshal(b, &errorDescriptions)
}
