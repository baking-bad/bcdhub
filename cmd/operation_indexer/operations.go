package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/cerrors"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/entrypoint"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func getOperations(rpc noderpc.Pool, es *elastic.Elastic, block int64, network string, contracts map[string]struct{}) ([]models.Operation, error) {
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

			res := parseContent(item)
			if res == nil {
				continue
			}

			opgHash := opg.Get("hash").String()
			protocol := opg.Get("protocol").String()
			if err := finishParseOperation(es, rpc, item, protocol, network, opgHash, block, contracts, res); err != nil {
				return nil, err
			}

			operations = append(operations, *res)

			internal := parseInternalOperations(es, rpc, item, res, contracts)
			operations = append(operations, internal...)

		}
	}

	return operations, nil
}

func finishParseOperation(es *elastic.Elastic, rpc noderpc.Pool, item gjson.Result, protocol, network, hash string, level int64, contracts map[string]struct{}, op *models.Operation) error {
	op.Hash = hash
	op.Level = level
	op.Network = network
	op.Protocol = protocol
	op.IndexedTime = time.Now().UnixNano() / 1000

	if isContract(contracts, op.Destination) {
		if isApplied(op) {
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

		if err := getEntrypoint(es, item, op); err != nil {
			return err
		}
	}

	return nil
}

func getEntrypoint(es *elastic.Elastic, item gjson.Result, op *models.Operation) error {
	if op.Parameters != "" && strings.HasPrefix(op.Destination, "KT") && !contractparser.IsParametersError(op.Result.Errors) {
		metadata, err := meta.GetMetadata(es, op.Destination, op.Network, "parameter", op.Protocol)
		if err != nil {
			return err
		}

		params := item.Get("parameters")
		ep, err := entrypoint.Get(params, metadata)
		if err != nil {
			return err
		}
		op.Entrypoint = ep
	}
	return nil
}

func isApplied(op *models.Operation) bool {
	return op.Result != nil && op.Result.Status == "applied"
}

func needParse(item gjson.Result, idx int) bool {
	kind := item.Get("kind").String()
	originationCondition := kind == consts.Origination && item.Get("script").Exists()
	prefixCondition := strings.HasPrefix(item.Get("source").String(), "KT") || strings.HasPrefix(item.Get("destination").String(), "KT")
	transactionCondition := kind == consts.Transaction && prefixCondition
	return originationCondition || transactionCondition
}

func parseContent(item gjson.Result) *models.Operation {
	op := models.Operation{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
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
	res, bu := parseResult(item)
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

func createResult(item gjson.Result, path string) *models.OperationResult {
	result := &models.OperationResult{
		Status:                       item.Get(path + ".status").String(),
		ConsumedGas:                  item.Get(path + ".consumed_gas").Int(),
		StorageSize:                  item.Get(path + ".storage_size").Int(),
		PaidStorageSizeDiff:          item.Get(path + ".paid_storage_size_diff").Int(),
		Originated:                   item.Get(path + ".originated_contracts.0").String(),
		AllocatedDestinationContract: item.Get(path + ".allocated_destination_contract").Bool(),
	}
	result.Errors = cerrors.ParseArray(data.Get("errors"))
	return result
}

func parseResult(item gjson.Result) (*models.OperationResult, []models.BalanceUpdate) {
	path := fmt.Sprintf("metadata.operation_result")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}
	return createResult(item, path), parseBalanceUpdates(item, path)
}

func parseInternalOperations(es *elastic.Elastic, rpc noderpc.Pool, item gjson.Result, main *models.Operation, contracts map[string]struct{}) []models.Operation {
	path := fmt.Sprintf("metadata.internal_operation_results")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("metadata.internal_operations")
		if !item.Get(path).Exists() {
			return nil
		}
	}

	res := make([]models.Operation, 0)
	for _, op := range item.Get(path).Array() {
		val := parseContent(op)
		val.Counter = main.Counter
		if err := finishParseOperation(es, rpc, op, main.Protocol, main.Network, main.Hash, main.Level, contracts, val); err != nil {
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
