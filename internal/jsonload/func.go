package jsonload

import (
	"encoding/json"
	"log"
	"os"
)

// StructFromFile - loads JSON file to struct
func StructFromFile(path string, response interface{}) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Printf("Can't open %v file.\n", path)
		return err
	}
	defer jsonFile.Close()

	jsonParser := json.NewDecoder(jsonFile)
	if err = jsonParser.Decode(&response); err != nil {
		log.Printf("Can't unmarshal data from %v file.\n", path)
		return err
	}
	return nil
}
