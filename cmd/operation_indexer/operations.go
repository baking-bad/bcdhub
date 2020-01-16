package main

import (
	"fmt"
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

const (
	origination = "origination"
	reveal      = "origination"
	delegation  = "delegation"
	transaction = "transaction"
)

type operation struct {
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`

	Level         int64  `json:"level"`
	Kind          string `json:"kind"`
	Source        string `json:"source"`
	Fee           int64  `json:"fee,omitempty"`
	Counter       int64  `json:"counter,omitempty"`
	GasLimit      int64  `json:"gas_limit,omitempty"`
	StorageLimit  int64  `json:"storage_limit,omitempty"`
	Amount        int64  `json:"amount,omitempty"`
	Destination   string `json:"destination,omitempty"`
	PublicKey     string `json:"public_key,omitempty"`
	ManagerPubKey string `json:"manager_pubkey,omitempty"`
	Balance       int64  `json:"balance,omitempty"`
	Delegate      string `json:"delegate,omitempty"`
	Parameters    string `json:"parameters,omitempty"`

	BalanceUpdates []balanceUpdate `json:"balance_updates,omitempty"`
	Result         *result         `json:"result,omitempty"`

	DeffatedStorage string   `json:"deffated_storage,omitempty"`
	BigMapKeyHashes []string `json:"big_map_key_hashes,omitempty"`
}

type balanceUpdate struct {
	Kind     string `json:"kind"`
	Contract string `json:"contract,omitempty"`
	Change   int64  `json:"change"`
	Category string `json:"category,omitempty"`
	Delegate string `json:"delegate,omitempty"`
	Cycle    int    `json:"cycle,omitempty"`
}

type result struct {
	Status              string `json:"status"`
	ConsumedGas         int64  `json:"consumed_gas,omitempty"`
	StorageSize         int64  `json:"storage_size,omitempty"`
	PaidStorageSizeDiff int64  `json:"paid_storage_size_diff,omitempty"`
	Originated          string `json:"-"`
	Errors              string `json:"errors,omitempty"`

	BalanceUpdates []balanceUpdate `json:"balance_updates,omitempty"`
}

func getOperations(rpc *noderpc.NodeRPC, es *elastic.Elastic, block int64, network string, knownContracts map[string]struct{}) ([]operation, error) {
	data, err := rpc.GetOperations(block)
	if err != nil {
		return nil, err
	}
	operations := make([]operation, 0)

	for _, opg := range data.Array() {
		for idx, item := range opg.Get("contents").Array() {
			if !needParse(item, idx, knownContracts) {
				continue
			}

			protocol := opg.Get("protocol").String()
			res := parseContent(item, protocol)
			if res == nil {
				continue
			}
			opgHash := opg.Get("hash").String()
			if err := finishParseOperation(es, rpc, item, protocol, network, opgHash, block, res, knownContracts); err != nil {
				return nil, err
			}

			operations = append(operations, *res)

			internal := parseInternalOperations(es, rpc, item, protocol, network, opgHash, block, knownContracts)
			operations = append(operations, internal...)

		}
	}

	return operations, nil
}

func finishParseOperation(es *elastic.Elastic, rpc *noderpc.NodeRPC, item gjson.Result, protocol, network, hash string, level int64, op *operation, knownContracts map[string]struct{}) error {
	op.Hash = hash
	op.Level = level

	if _, ok := knownContracts[op.Destination]; ok{
	rs, err := getRichStorage(es, rpc, item, level, network, protocol)
	if err != nil {
		return err
	}
	if rs == nil {
		return nil
	}
	op.DeffatedStorage = rs.DeffatedStorage

	if len(rs.BigMapDiffs) > 0 {
		op.BigMapKeyHashes = make([]string, len(rs.BigMapDiffs))
		for i := range rs.BigMapDiffs {
			op.BigMapKeyHashes[i] = rs.BigMapDiffs[i].KeyHash
		}

		if err := es.BulkSaveBigMapDiffs(rs.BigMapDiffs); err != nil {
			return err
		}
	}}

	return nil
}

func needParse(item gjson.Result, idx int, knownContracts map[string]struct{}) bool {
	kind := item.Get("kind").String()
	if kind == origination && item.Get("script").Exists() {
		return true
	}
	source := item.Get("source")
	if source.Exists() {
		if _, ok := knownContracts[source.String()]; ok {
			return true
		}
	}
	destination := item.Get("destination")
	if destination.Exists() {
		if _, ok := knownContracts[destination.String()]; ok {
			return true
		}
	}
	return false
}

func parseContent(item gjson.Result, protocol string) *operation {
	op := operation{
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
	if op.Kind == origination {
		op.Destination = res.Originated
	}

	return &op
}

func parseBalanceUpdates(item gjson.Result, root string) []balanceUpdate {
	filter := fmt.Sprintf("%s.balance_updates.#(kind==\"contract\")#", root)

	contracts := item.Get(filter).Array()
	bu := make([]balanceUpdate, len(contracts))
	for i := range contracts {
		bu[i] = balanceUpdate{
			Kind:     contracts[i].Get("kind").String(),
			Contract: contracts[i].Get("contract").String(),
			Change:   contracts[i].Get("change").Int(),
		}
	}
	return bu
}

func createResult(item gjson.Result, path, protocol string) *result {
	return &result{
		Status:              item.Get(path + ".status").String(),
		ConsumedGas:         item.Get(path + ".consumed_gas").Int(),
		StorageSize:         item.Get(path + ".storage_size").Int(),
		PaidStorageSizeDiff: item.Get(path + ".paid_storage_size_diff").Int(),
		Originated:          item.Get(path + ".originated_contracts.0").String(),
		Errors:              item.Get(path + ".errors").String(),
	}
}

func parseResult(item gjson.Result, protocol string) (*result, []balanceUpdate) {
	path := fmt.Sprintf("metadata.operation_result")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}
	return createResult(item, path, protocol), parseBalanceUpdates(item, path)
}

func parseInternalOperations(es *elastic.Elastic, rpc *noderpc.NodeRPC, item gjson.Result, protocol, network, hash string, level int64, knownContracts map[string]struct{}) []operation {
	path := fmt.Sprintf("metadata.internal_operation_results")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("metadata.internal_operations")
		if !item.Get(path).Exists() {
			return nil
		}
	}

	res := make([]operation, 0)
	for _, op := range item.Get(path).Array() {
		val := parseContent(op, protocol)
		if err := finishParseOperation(es, rpc, op, protocol, network, hash, level, val, knownContracts); err != nil {
			log.Println(err)
			continue
		}
		val.Internal = true
		res = append(res, *val)
	}
	return res
}
