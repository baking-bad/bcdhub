package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

func getContractProjectID(c models.Contract, buckets []models.Contract) (string, error) {
	for i := len(buckets) - 1; i > -1; i-- {
		ok, err := compare(c, buckets[i])
		if err != nil {
			return "", err
		}

		if ok {
			return buckets[i].ProjectID, nil
		}
	}

	return strings.ReplaceAll(uuid.New().String(), "-", ""), nil
}

func parseContract(contract models.Contract) error {
	if contract.Alias == "" {
		if err := setAlias(&contract); err != nil {
			return err
		}
	}

	if contract.ProjectID == "" {
		buckets, err := ctx.ES.GetLastProjectContracts()
		if err != nil {
			return err
		}
		projID, err := getContractProjectID(contract, buckets)
		if err != nil {
			return err
		}
		contract.ProjectID = projID
	}

	logger.Info("Contract %s to project %s", contract.Address, contract.ProjectID)

	if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, contract.ID, contract); err != nil {
		return err
	}
	return nil
}

func getContract(data amqp.Delivery) error {
	var contractID string
	if err := json.Unmarshal(data.Body, &contractID); err != nil {
		return fmt.Errorf("[getContract] Unmarshal message body error: %s", err)
	}

	c, err := ctx.ES.GetContractByID(contractID)
	if err != nil {
		return fmt.Errorf("[getContract] Find contract error: %s", err)
	}

	if err := parseContract(c); err != nil {
		return fmt.Errorf("[getContract] Compute error message: %s", err)
	}
	return nil
}
