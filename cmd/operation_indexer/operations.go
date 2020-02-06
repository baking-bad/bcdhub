package main

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func getOperations(rpc *noderpc.NodeRPC, es *elastic.Elastic, block int64, network string, contracts map[string]struct{}) ([]models.Operation, error) {
	data, err := rpc.GetOperations(block)
	if err != nil {
		return nil, err
	}
	operations := make([]models.Operation, 0)

	for _, opg := range data.Array() {
		for idx, item := range opg.Get("contents").Array() {
			if !needParse(item, idx) {
				continue
			}

			protocol := opg.Get("protocol").String()
			res := parseContent(item, protocol)
			if res == nil {
				continue
			}
			opgHash := opg.Get("hash").String()
			if err := finishParseOperation(es, rpc, item, protocol, network, opgHash, block, contracts, res); err != nil {
				return nil, err
			}

			operations = append(operations, *res)

			internal := parseInternalOperations(es, rpc, item, protocol, network, opgHash, block, contracts)
			operations = append(operations, internal...)

		}
	}

	return operations, nil
}

func finishParseOperation(es *elastic.Elastic, rpc *noderpc.NodeRPC, item gjson.Result, protocol, network, hash string, level int64, contracts map[string]struct{}, op *models.Operation) error {
	op.Hash = hash
	op.Level = level
	op.Network = network

	if isContract(contracts, op.Destination) {
		rs, err := getRichStorage(es, rpc, item, level, protocol, op.ID)
		if err != nil {
			return err
		}
		if rs.Empty {
			return nil
		}
		op.DeffatedStorage = rs.DeffatedStorage

		if len(rs.BigMapDiffs) > 0 {
			if err := es.BulkSaveBigMapDiffs(rs.BigMapDiffs); err != nil {
				return err
			}
		}
	}
	return nil
}

func needParse(item gjson.Result, idx int) bool {
	kind := item.Get("kind").String()
	return (kind == consts.Origination && item.Get("script").Exists()) || kind == consts.Transaction
}

func parseContent(item gjson.Result, protocol string) *models.Operation {
	op := models.Operation{
		ID:             uuid.New().String(),
		Protocol:       protocol,
		Kind:           item.Get("kind").String(),
		Source:         item.Get("source").String(),
		Fee:            item.Get("fee").Int(),
		Counter:        item.Get("counter").Int(),
		GasLimit:       item.Get("gas_limit").Int(),
		StorageLimit:   item.Get("storage_limit").Int(),
		Amount:         item.Get("amount").Int(),
		Destination:    item.Get("destination").String(),
		PublicKey:      item.Get("public_key").String(),
		Balance:        item.Get("balance").Int(),
		ManagerPubKey:  item.Get("manager_pubkey").String(),
		Delegate:       item.Get("delegate").String(),
		Parameters:     item.Get("parameters").String(),
		BalanceUpdates: parseBalanceUpdates(item, "metadata"),
	}
	res, bu := parseResult(item, protocol)
	op.Result = res
	op.BalanceUpdates = append(op.BalanceUpdates, bu...)
	if op.Kind == consts.Origination {
		op.Destination = res.Originated
	}

	return &op
}

func parseBalanceUpdates(item gjson.Result, root string) []models.BalanceUpdate {
	filter := fmt.Sprintf("%s.balance_updates.#(kind==\"contract\")#", root)

	contracts := item.Get(filter).Array()
	bu := make([]models.BalanceUpdate, len(contracts))
	for i := range contracts {
		bu[i] = models.BalanceUpdate{
			Kind:     contracts[i].Get("kind").String(),
			Contract: contracts[i].Get("contract").String(),
			Change:   contracts[i].Get("change").Int(),
		}
	}
	return bu
}

func createResult(item gjson.Result, path, protocol string) *models.OperationResult {
	return &models.OperationResult{
		Status:              item.Get(path + ".status").String(),
		ConsumedGas:         item.Get(path + ".consumed_gas").Int(),
		StorageSize:         item.Get(path + ".storage_size").Int(),
		PaidStorageSizeDiff: item.Get(path + ".paid_storage_size_diff").Int(),
		Originated:          item.Get(path + ".originated_contracts.0").String(),
		Errors:              item.Get(path + ".errors").String(),
	}
}

func parseResult(item gjson.Result, protocol string) (*models.OperationResult, []models.BalanceUpdate) {
	path := fmt.Sprintf("metadata.operation_result")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}
	return createResult(item, path, protocol), parseBalanceUpdates(item, path)
}

func parseInternalOperations(es *elastic.Elastic, rpc *noderpc.NodeRPC, item gjson.Result, protocol, network, hash string, level int64, contracts map[string]struct{}) []models.Operation {
	path := fmt.Sprintf("metadata.internal_operation_results")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("metadata.internal_operations")
		if !item.Get(path).Exists() {
			return nil
		}
	}

	res := make([]models.Operation, 0)
	for _, op := range item.Get(path).Array() {
		val := parseContent(op, protocol)
		if err := finishParseOperation(es, rpc, op, protocol, network, hash, level, contracts, val); err != nil {
			logger.Error(err)
			continue
		}
		val.Internal = true
		res = append(res, *val)
	}
	return res
}

func isContract(contracts map[string]struct{}, address string) bool {
	_, ok := contracts[address]
	return ok
}
