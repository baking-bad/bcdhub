package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/streadway/amqp"
)

func operationMetrics(op models.Operation, contract *models.Contract) error {
	stats, err := ctx.ES.GetContractStats(contract.Address, contract.Network)
	if err != nil {
		return err
	}
	contract.TxCount = stats.TxCount
	contract.LastAction = models.BCDTime{
		Time: stats.LastAction,
	}
	contract.SumTxAmount = stats.SumTxAmount
	contract.MedianConsumedGas = stats.MedianConsumedGas

	rpc, err := ctx.getRPC(contract.Network)
	if err != nil {
		return err
	}
	balance, err := rpc.GetContractBalance(contract.Address, op.Level)
	if err != nil {
		return err
	}
	contract.Balance = balance
	return nil
}

func getOperation(data amqp.Delivery) error {
	var operationID string
	if err := json.Unmarshal(data.Body, &operationID); err != nil {
		return fmt.Errorf("[parseOperation] Unmarshal message body error: %s", err)
	}
	op, err := ctx.ES.GetOperationByID(operationID)
	if err != nil {
		return fmt.Errorf("[parseOperation] Find operation error: %s", err)
	}

	if err := setOperationAliases(&op); err != nil {
		return fmt.Errorf("[parseOperation] Error during set operation alias: %s", err)
	}

	if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, op.ID, op); err != nil {
		return err
	}

	for _, address := range []string{op.Source, op.Destination} {
		if !strings.HasPrefix(address, "KT") {
			continue
		}
		c, err := ctx.ES.GetContract(map[string]interface{}{
			"network": op.Network,
			"address": address,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Unknown contract") {
				continue
			}
			return fmt.Errorf("[parseOperation] Find contract error: %s", err)
		}
		if err := operationMetrics(op, &c); err != nil {
			return fmt.Errorf("[parseOperation] Compute error message: %s", err)
		}

		if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
			return err
		}
	}

	logger.Info("Operation %s processed", operationID)
	return nil
}
