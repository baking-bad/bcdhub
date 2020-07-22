package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// DefaultParser -
type DefaultParser struct {
	rpc            noderpc.Pool
	es             *elastic.Elastic
	filesDirectory string

	storageParser storage.Parser
}

// NewDefaultParser -
func NewDefaultParser(rpc noderpc.Pool, es *elastic.Elastic, filesDirectory string) *DefaultParser {
	return &DefaultParser{
		rpc:            rpc,
		es:             es,
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *DefaultParser) Parse(opg gjson.Result, network string, head noderpc.Header) ([]elastic.Model, error) {
	parsedModels := make([]elastic.Model, 0)

	for idx, item := range opg.Get("contents").Array() {
		need, err := p.needParse(item, network, idx)
		if err != nil {
			return nil, err
		}
		if !need {
			continue
		}

		hash := opg.Get("hash").String()
		helpers.SetTagSentry("hash", hash)

		resultModels, mainOperation, err := p.parseContent(item, network, hash, head, int64(idx))
		if err != nil {
			return nil, err
		}

		if len(resultModels) > 0 {
			parsedModels = append(parsedModels, resultModels...)
		}

		internalModels, err := p.parseInternalOperations(item, mainOperation, head, int64(idx))
		if err != nil {
			return nil, err
		}
		parsedModels = append(parsedModels, internalModels...)
	}
	return parsedModels, nil
}

func (p *DefaultParser) parseContent(data gjson.Result, network, hash string, head noderpc.Header, contentIdx int64) ([]elastic.Model, models.Operation, error) {
	kind := data.Get("kind").String()
	switch kind {
	case consts.Origination:
		return p.parseOrigination(data, network, hash, head, contentIdx)
	default:
		return p.parseTransaction(data, network, hash, head, contentIdx)
	}
}

func (p *DefaultParser) parseTransaction(data gjson.Result, network, hash string, head noderpc.Header, contentIdx int64) ([]elastic.Model, models.Operation, error) {
	op := models.Operation{
		ID:             helpers.GenerateID(),
		Network:        network,
		Hash:           hash,
		Protocol:       head.Protocol,
		Level:          head.Level,
		Timestamp:      head.Timestamp,
		Kind:           data.Get("kind").String(),
		Source:         data.Get("source").String(),
		Fee:            data.Get("fee").Int(),
		Counter:        data.Get("counter").Int(),
		GasLimit:       data.Get("gas_limit").Int(),
		StorageLimit:   data.Get("storage_limit").Int(),
		Amount:         data.Get("amount").Int(),
		Destination:    data.Get("destination").String(),
		PublicKey:      data.Get("public_key").String(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		BalanceUpdates: p.parseBalanceUpdates(data, "metadata"),
		IndexedTime:    time.Now().UnixNano() / 1000,
		ContentIndex:   contentIdx,
	}
	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors

	additionalModels, err := p.finishParseOperation(data, &op)
	if err != nil {
		return nil, op, err
	}
	transactionModels := []elastic.Model{&op}

	if len(additionalModels) > 0 {
		transactionModels = append(transactionModels, additionalModels...)
	}
	return transactionModels, op, nil
}

func (p *DefaultParser) parseOrigination(data gjson.Result, network, hash string, head noderpc.Header, contentIdx int64) ([]elastic.Model, models.Operation, error) {
	op := models.Operation{
		ID:             helpers.GenerateID(),
		Network:        network,
		Hash:           hash,
		Protocol:       head.Protocol,
		Level:          head.Level,
		Timestamp:      head.Timestamp,
		Kind:           data.Get("kind").String(),
		Source:         data.Get("source").String(),
		Fee:            data.Get("fee").Int(),
		Counter:        data.Get("counter").Int(),
		GasLimit:       data.Get("gas_limit").Int(),
		StorageLimit:   data.Get("storage_limit").Int(),
		Amount:         data.Get("balance").Int(),
		PublicKey:      data.Get("public_key").String(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		Script:         data.Get("script"),
		BalanceUpdates: p.parseBalanceUpdates(data, "metadata"),
		IndexedTime:    time.Now().UnixNano() / 1000,
		ContentIndex:   contentIdx,
	}

	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors
	op.Destination = operationResult.Originated

	protoSymLink, err := meta.GetProtoSymLink(op.Protocol)
	if err != nil {
		return nil, op, err
	}

	originationModels := []elastic.Model{&op}

	if p.isApplied(op) {
		contractModels, err := createNewContract(p.es, op, p.filesDirectory, protoSymLink)
		if err != nil {
			return nil, op, err
		}
		if len(contractModels) > 0 {
			originationModels = append(originationModels, contractModels...)
		}
	}

	additionalModels, err := p.finishParseOperation(data, &op)
	if len(additionalModels) > 0 {
		originationModels = append(originationModels, additionalModels...)
	}
	return originationModels, op, err
}

func (p *DefaultParser) parseBalanceUpdates(item gjson.Result, root string) []models.BalanceUpdate {
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

func (p *DefaultParser) createResult(item gjson.Result, path string) *models.OperationResult {
	result := &models.OperationResult{
		Status:                       item.Get(path + ".status").String(),
		ConsumedGas:                  item.Get(path + ".consumed_gas").Int(),
		StorageSize:                  item.Get(path + ".storage_size").Int(),
		PaidStorageSizeDiff:          item.Get(path + ".paid_storage_size_diff").Int(),
		Originated:                   item.Get(path + ".originated_contracts.0").String(),
		AllocatedDestinationContract: item.Get(path+".allocated_destination_contract").Bool() || item.Get("kind").String() == consts.Origination,
	}
	err := item.Get(path + ".errors")
	result.Errors = cerrors.ParseArray(err)
	return result
}

func (p *DefaultParser) parseMetadata(item gjson.Result) (*models.OperationResult, []models.BalanceUpdate) {
	path := fmt.Sprintf("metadata.operation_result")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}
	return p.createResult(item, path), p.parseBalanceUpdates(item, path)
}

func (p *DefaultParser) finishParseOperation(item gjson.Result, op *models.Operation) ([]elastic.Model, error) {
	if !strings.HasPrefix(op.Destination, "KT") {
		return nil, nil
	}

	metadata, err := meta.GetContractMetadata(p.es, op.Destination)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, nil
		}
		return nil, err
	}

	resultModels := make([]elastic.Model, 0)

	if p.isApplied(*op) {
		rs, err := p.getRichStorage(item, metadata, op)
		if err != nil {
			return nil, err
		}
		if rs.Empty {
			return nil, err
		}
		op.DeffatedStorage = rs.DeffatedStorage

		if len(rs.Models) > 0 {
			resultModels = append(resultModels, rs.Models...)
		}

		if op.Kind == consts.Transaction {
			migration, err := p.findMigration(item, op)
			if err != nil {
				return nil, err
			}

			if migration != nil {
				resultModels = append(resultModels, migration)
			}
		}
	}
	if op.Kind == consts.Transaction {
		return resultModels, p.getEntrypoint(item, metadata, op)
	}
	return resultModels, nil
}

func (p *DefaultParser) findMigration(item gjson.Result, op *models.Operation) (*models.Migration, error) {
	path := fmt.Sprintf("metadata.operation_result.big_map_diff")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result.big_map_diff")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}
	for _, bmd := range item.Get(path).Array() {
		if bmd.Get("action").String() != "update" {
			continue
		}

		value := bmd.Get("value")
		if contractparser.HasLambda(value) {
			logger.Info("[%s] Migration detected: %s", op.Network, op.Destination)
			return &models.Migration{
				ID:          helpers.GenerateID(),
				IndexedTime: time.Now().UnixNano() / 1000,

				Network:   op.Network,
				Level:     op.Level,
				Protocol:  op.Protocol,
				Address:   op.Destination,
				Timestamp: op.Timestamp,
				Hash:      op.Hash,
				Kind:      consts.MigrationLambda,
			}, nil
		}
	}
	return nil, nil
}

func (p *DefaultParser) isApplied(op models.Operation) bool {
	return op.Result != nil && op.Status == "applied"
}

func (p *DefaultParser) getEntrypoint(item gjson.Result, metadata *meta.ContractMetadata, op *models.Operation) error {
	m, err := metadata.Get(consts.PARAMETER, op.Protocol)
	if err != nil {
		return err
	}

	params := item.Get("parameters")
	if params.Exists() {
		ep, err := m.GetByPath(params)
		if err != nil && op.Errors == nil {
			return err
		}
		op.Entrypoint = ep
	} else {
		op.Entrypoint = "default"
	}

	return nil
}

func (p *DefaultParser) parseInternalOperations(item gjson.Result, main models.Operation, head noderpc.Header, contentIdx int64) ([]elastic.Model, error) {
	path := fmt.Sprintf("metadata.internal_operation_results")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("metadata.internal_operations")
		if !item.Get(path).Exists() {
			return nil, nil
		}
	}

	internalModels := make([]elastic.Model, 0)
	for _, op := range item.Get(path).Array() {
		parsedModels, _, err := p.parseContent(op, main.Network, main.Hash, head, contentIdx)
		if err != nil {
			return nil, err
		}
		for j := range parsedModels {
			if internalOperation, ok := parsedModels[j].(*models.Operation); ok {
				internalOperation.Counter = main.Counter
				internalOperation.Hash = main.Hash
				internalOperation.Level = main.Level
				internalOperation.Timestamp = main.Timestamp
				internalOperation.Internal = true

				nonce := op.Get("nonce").Int()
				internalOperation.Nonce = &nonce
			}

			internalModels = append(internalModels, parsedModels[j])
		}
	}
	return internalModels, nil
}

func (p *DefaultParser) needParse(item gjson.Result, network string, idx int) (bool, error) {
	kind := item.Get("kind").String()
	source := item.Get("source").String()
	destination := item.Get("destination").String()
	prefixCondition := strings.HasPrefix(source, "KT") || strings.HasPrefix(destination, "KT")
	transactionCondition := kind == consts.Transaction && prefixCondition
	originationCondition := kind == consts.Origination && item.Get("script").Exists()
	return originationCondition || transactionCondition, nil
}

func (p *DefaultParser) getRichStorage(data gjson.Result, metadata *meta.ContractMetadata, op *models.Operation) (storage.RichStorage, error) {
	if p.storageParser == nil {
		parser, err := contractparser.MakeStorageParser(p.rpc, p.es, op.Protocol, false)
		if err != nil {
			return storage.RichStorage{Empty: true}, err
		}
		p.storageParser = parser
	}

	protoSymLink, err := meta.GetProtoSymLink(op.Protocol)
	if err != nil {
		return storage.RichStorage{Empty: true}, err
	}

	m, ok := metadata.Storage[protoSymLink]
	if !ok {
		return storage.RichStorage{Empty: true}, fmt.Errorf("Unknown metadata: %s", protoSymLink)
	}

	switch op.Kind {
	case consts.Transaction:
		return p.storageParser.ParseTransaction(data, m, *op)
	case consts.Origination:
		rs, err := p.storageParser.ParseOrigination(data, m, *op)
		if err != nil {
			return rs, err
		}
		storage, err := p.rpc.GetScriptStorageJSON(op.Destination, op.Level)
		if err != nil {
			return rs, err
		}
		rs.DeffatedStorage = storage.String()
		return rs, err
	}
	return storage.RichStorage{Empty: true}, nil
}
