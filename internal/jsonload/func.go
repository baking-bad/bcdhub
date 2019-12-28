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
		log.Println("Can't open app_config.json file. Loading default config.")
		return err
	}
	defer jsonFile.Close()

	jsonParser := json.NewDecoder(jsonFile)
	if err = jsonParser.Decode(&response); err != nil {
		log.Println("Can't unmarshal data from app_config.json file. Loading default config.")
		return err
	}
	return nil
}
