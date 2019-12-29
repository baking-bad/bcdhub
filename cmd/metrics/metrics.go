package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/db/account"
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/db/project"
	"github.com/aopoltorzhicky/bcdhub/internal/db/relation"
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

func computeMetrics(db *gorm.DB, c contract.Contract) (upd contract.Contract, err error) {
	script, err := contractparser.New(c.Script.RawMessage, defaultLabels)
	if err != nil {
		return
	}
	if err = script.Parse(); err != nil {
		return
	}
	// Detect language
	upd.Language = script.Language()
	// Compute hash code
	upd.HashCode = script.Code.HashCode
	// Set project ID
	proj, err := project.Search(db, upd.HashCode)
	if err == nil {
		upd.ProjectID = proj.ID
	}

	// Set kind
	upd.Kind = script.Kind()

	if err = saveTags(script.Tags, c); err != nil {
		return
	}

	if err = saveHardcodedAddresses(script.HardcodedAddresses, c); err != nil {
		return
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

func saveTags(scriptTags map[string]struct{}, c contract.Contract) (err error) {
	query := ""
	for tag := range scriptTags {
		query += fmt.Sprintf(" ('%s', '%d'),", tag, c.ID)
	}
	if len(query) > 0 {
		// Remove last comma
		query = query[:len(query)-1]

		if err := ms.DB.Exec(fmt.Sprintf("INSERT INTO tags (tag, contract_id) VALUES %s", query)).Error; err != nil {
			return err
		}
	}
	return
}

func saveHardcodedAddresses(addresses []string, c contract.Contract) error {
	relationValues := ""
	for _, a := range addresses {
		var acc account.Account
		if err := ms.DB.FirstOrCreate(&acc, account.Account{
			Network: c.Network,
			Address: a,
		}).Error; err != nil {
			log.Println(err)
			continue
		}
		relationValues = fmt.Sprintf(" ('%d', '%d', '%s'),", c.AddressID, acc.ID, relation.Hardcoded)
	}

	if len(relationValues) > 0 {
		// Remove last comma
		relationValues = relationValues[:len(relationValues)-1]

		if err := ms.DB.Exec(fmt.Sprintf("INSERT INTO relations (account_id, relation_id, type) VALUES %s", relationValues)).Error; err != nil {
			return err
		}
	}
	return nil
}
