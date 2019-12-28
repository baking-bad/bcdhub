package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/db/project"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/jinzhu/gorm"
)

var defaultLabels = map[string]int{}

const labelsFile = "labels.json"

func getLabels() error {
	if err := jsonload.StructFromFile(labelsFile, &defaultLabels); err != nil {
		return err
	}
	return nil
}

func saveLabels() error {
	f, err := os.Create(labelsFile)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(defaultLabels)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(b))
	return err
}

func getCodeHash(script map[string]interface{}) (string, error) {
	code, ok := script["code"]
	if !ok {
		return "", fmt.Errorf("Invalid script: can`t find 'code' tag")
	}
	value, ok := code.([]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid script: can`t convert 'code' tag to map")
	}
	if len(value) < 3 {
		return "", fmt.Errorf("Invalid script: 'code' if %d length", len(value))
	}
	codeValue := value[2]
	return getHash(codeValue, defaultLabels)
}

func getHashPrimitive(value map[string]interface{}, labels map[string]int) (hash string, err error) {
	prim, primOK := value["prim"]
	if !primOK {
		return "", nil
	}
	sPrim := prim.(string)
	args, argsOK := value["args"].([]interface{})

	label, labelOK := labels[sPrim]
	if !labelOK {
		label = len(labels)
		labels[sPrim] = label
	}

	if argsOK {
		h, e := getHash(args, labels)
		if e != nil {
			return "", e
		}
		hash = fmt.Sprintf("%x%s", label, h)
	} else {
		hash = fmt.Sprintf("%x", label)
	}
	return
}

func getHash(value interface{}, labels map[string]int) (string, error) {
	hash := ""
	switch t := value.(type) {
	case []interface{}:
		for _, arg := range t {
			h, err := getHash(arg, labels)
			if err != nil {
				return "", err
			}
			hash += h
		}
	case map[string]interface{}:
		h, err := getHashPrimitive(t, labels)
		if err != nil {
			return "", err
		}
		hash = h
	default:
		return "", fmt.Errorf("Unknown value type: %T", t)
	}
	return hash, nil
}

func computeMetrics(rpc *noderpc.NodeRPC, db *gorm.DB, c contract.Contract) (upd contract.Contract, err error) {
	// Detect language
	upd.Language, err = detectLanguage(c.Script)
	if err != nil {
		log.Println(err)
	}

	var scriptMap map[string]interface{}
	if err = json.Unmarshal(c.Script.RawMessage, &scriptMap); err == nil {
		// Compute hash code
		upd.HashCode, err = getCodeHash(scriptMap)
		if err != nil {
			return
		}

		// Set project ID
		proj, err := project.Search(db, upd.HashCode)
		if err == nil {
			upd.ProjectID = proj.ID
		}

		// Set kind
		upd.Kind = detectKind(upd.HashCode)
	}

	return
}

func computeEmptyMetrics(rpcs map[string]*noderpc.NodeRPC, db *gorm.DB) error {
	var contracts []contract.Contract
	if err := db.Where("script = ''").Find(&contracts).Error; err != nil {
		return err
	}

	for _, c := range contracts {
		if err := compute(c); err != nil {
			log.Println(err)
		}
	}

	return nil
}
