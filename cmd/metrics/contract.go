package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
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

	projID := strings.ReplaceAll(uuid.New().String(), "-", "")
	proj := models.Project{
		ID:    projID,
		Alias: projID,
	}

	if _, err := ctx.ES.AddDocumentWithID(proj, elastic.DocProjects, projID); err != nil {
		return "", err
	}

	return projID, nil
}

func parseContract(contract models.Contract) error {
	buckets, err := ctx.ES.GetLastProjectContracts()
	if err != nil {
		return err
	}
	projID, err := getContractProjectID(contract, buckets)
	if err != nil {
		return err
	}
	contract.ProjectID = projID

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
