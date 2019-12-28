package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/db/account"
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/db/relation"
)

// Handler -
type Handler func(contract.Contract) error

var (
	handlers = []Handler{findTags, findRelations}
)

func initCompute() {
	var contracts []contract.Contract
	if err := ms.DB.Find(&contracts).Error; err != nil {
		log.Println(err)
		return
	}

	for _, c := range contracts {
		log.Printf("Compute for %s", c.Address.Address)
		if err := compute(c); err != nil {
			log.Println(err)
		}
	}
}

func getScriptString(c contract.Contract) (string, error) {
	bScript, err := c.Script.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(bScript), nil
}

func compute(c contract.Contract) error {
	for _, handler := range handlers {
		if err := handler(c); err != nil {
			return err
		}
	}

	return nil
}

func hasViewMethod(script string) bool {
	return strings.Contains(script, "\"prim\": \"contract\"")
}

func hasContractFactory(script string) bool {
	return strings.Contains(script, "CREATE_CONTRACT")
}

func hasDelegatable(script string) bool {
	return strings.Contains(script, "SET_DELEGATE")
}

func hasChainAware(script string) bool {
	return strings.Contains(script, "CHAIN_ID") || strings.Contains(script, "chain_id")
}

func hasCheckSig(script string) bool {
	return strings.Contains(script, "CHECK_SIGNATURE")
}

func findTags(c contract.Contract) error {
	script, err := getScriptString(c)
	if err != nil {
		return err
	}
	tags := ""

	if hasViewMethod(script) {
		log.Printf("%s has tag '%s'", c.Address.Address, ViewMethodTag)
		tags += fmt.Sprintf(" ('%s', '%d'),", ViewMethodTag, c.ID)
	}
	if hasContractFactory(script) {
		log.Printf("%s has tag '%s'", c.Address.Address, ContractFactoryTag)
		tags += fmt.Sprintf(" ('%s', '%d'),", ContractFactoryTag, c.ID)
	}
	if hasDelegatable(script) {
		log.Printf("%s has tag '%s'", c.Address.Address, DelegatableTag)
		tags += fmt.Sprintf(" ('%s', '%d'),", DelegatableTag, c.ID)
	}
	if hasChainAware(script) {
		log.Printf("%s has tag '%s'", c.Address.Address, ChainAwareTag)
		tags += fmt.Sprintf(" ('%s', '%d'),", ChainAwareTag, c.ID)
	}
	if hasCheckSig(script) {
		log.Printf("%s has tag '%s'", c.Address.Address, CheckSigTag)
		tags += fmt.Sprintf(" ('%s', '%d'),", CheckSigTag, c.ID)
	}

	if len(tags) > 0 {
		// Remove last comma
		tags = tags[:len(tags)-1]

		if err := ms.DB.Exec(fmt.Sprintf("INSERT INTO tags (tag, contract_id) VALUES %s", tags)).Error; err != nil {
			return err
		}
	}
	return nil
}

func findRelations(c contract.Contract) error {
	script, err := getScriptString(c)
	if err != nil {
		return err
	}
	addresses := findHardcodedAddresses(script)

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

func findHardcodedAddresses(script string) []string {
	regexString := "(tz1|KT1)[0-9A-Za-z]{33}"
	re := regexp.MustCompile(regexString)
	return re.FindAllString(script, -1)
}
