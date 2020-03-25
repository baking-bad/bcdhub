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
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// DefaultParser -
type DefaultParser struct {
	rpc            noderpc.Pool
	es             *elastic.Elastic
	filesDirectory string
	protocols      map[string]string
}

// NewDefaultParser -
func NewDefaultParser(rpc noderpc.Pool, es *elastic.Elastic, filesDirectory string, protocols map[string]string) *DefaultParser {
	return &DefaultParser{
		rpc:            rpc,
		es:             es,
		filesDirectory: filesDirectory,
		protocols:      protocols,
	}
}

// Parse -
func (p *DefaultParser) Parse(opg gjson.Result, network string, level int64) ([]models.Operation, []models.Contract, error) {
	ts, err := p.rpc.GetLevelTime(int(level))
	if err != nil {
		return nil, nil, err
	}

	operations := make([]models.Operation, 0)
	contracts := make([]models.Contract, 0)
	for idx, item := range opg.Get("contents").Array() {
		if !p.needParse(item, idx) {
			continue
		}

		protocol := opg.Get("protocol").String()
		op, contract, err := p.parseContent(item, network, protocol)
		if err != nil {
			return nil, nil, err
		}

		op.Hash = opg.Get("hash").String()
		op.Level = level
		op.Timestamp = ts
		op.IndexedTime = time.Now().UnixNano() / 1000

		if err := p.finishParseOperation(item, &op); err != nil {
			return nil, nil, err
		}

		operations = append(operations, op)
		if contract != nil {
			contracts = append(contracts, *contract)
		}

		internal, internalContracts, err := p.parseInternalOperations(item, op)
		if err != nil {
			return nil, nil, err
		}
		operations = append(operations, internal...)
		contracts = append(contracts, internalContracts...)
	}
	return operations, contracts, nil
}

func (p *DefaultParser) parseContent(data gjson.Result, network, protocol string) (models.Operation, *models.Contract, error) {
	kind := data.Get("kind").String()
	switch kind {
	case consts.Origination:
		return p.parseOrigination(data, network, protocol)
	default:
		return p.parseTransaction(data, network, protocol), nil, nil
	}
}

func (p *DefaultParser) parseTransaction(data gjson.Result, network, protocol string) models.Operation {
	op := models.Operation{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:        network,
		Protocol:       protocol,
		Kind:           data.Get("kind").String(),
		Source:         data.Get("source").String(),
		Fee:            data.Get("fee").Int(),
		Counter:        data.Get("counter").Int(),
		GasLimit:       data.Get("gas_limit").Int(),
		StorageLimit:   data.Get("storage_limit").Int(),
		Amount:         data.Get("amount").Int(),
		Destination:    data.Get("destination").String(),
		PublicKey:      data.Get("public_key").String(),
		Balance:        data.Get("balance").Int(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		BalanceUpdates: p.parseBalanceUpdates(data, "metadata"),
	}

	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors
	return op
}

func (p *DefaultParser) parseOrigination(data gjson.Result, network, protocol string) (models.Operation, *models.Contract, error) {
	op := models.Operation{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:        network,
		Protocol:       protocol,
		Kind:           data.Get("kind").String(),
		Source:         data.Get("source").String(),
		Fee:            data.Get("fee").Int(),
		Counter:        data.Get("counter").Int(),
		GasLimit:       data.Get("gas_limit").Int(),
		StorageLimit:   data.Get("storage_limit").Int(),
		Amount:         data.Get("balance").Int(),
		PublicKey:      data.Get("public_key").String(),
		Balance:        data.Get("balance").Int(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		Script:         data.Get("script"),
		BalanceUpdates: p.parseBalanceUpdates(data, "metadata"),
	}

	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors
	op.Destination = operationResult.Originated

	protoSymLink, ok := p.protocols[op.Protocol]
	if !ok {
		return op, nil, fmt.Errorf("[%s] Unknown protocol: %s", op.Network, op.Protocol)
	}

	if !contractparser.IsDelegateContract(op.Script) {
		contract, err := createNewContract(p.es, op, p.filesDirectory, protoSymLink)
		return op, contract, err
	}
	return op, nil, nil
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
		AllocatedDestinationContract: item.Get(path + ".allocated_destination_contract").Bool(),
	}
	result.Errors = cerrors.ParseArray(item.Get(path + ".errors"))
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

func (p *DefaultParser) finishParseOperation(item gjson.Result, op *models.Operation) error {
	if strings.HasPrefix(op.Destination, "KT") && op.Kind == consts.Transaction {
		metadata, err := meta.GetMetadata(p.es, op.Destination, op.Network, "parameter", op.Protocol)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				return nil
			}
			return err
		}

		if p.isApplied(op) {
			rs, err := p.getRichStorage(item, metadata, op)
			if err != nil {
				return err
			}
			if rs.Empty {
				return nil
			}
			op.DeffatedStorage = rs.DeffatedStorage

			if len(rs.BigMapDiffs) > 0 {
				if err := p.es.BulkSaveBigMapDiffs(rs.BigMapDiffs); err != nil {
					return err
				}
			}
		}

		if err := p.getEntrypoint(item, metadata, op); err != nil {
			return err
		}
	}

	return nil
}

func (p *DefaultParser) isApplied(op *models.Operation) bool {
	return op.Result != nil && op.Status == "applied"
}

func (p *DefaultParser) getEntrypoint(item gjson.Result, metadata meta.Metadata, op *models.Operation) error {
	params := item.Get("parameters")
	if params.Exists() {
		ep, err := metadata.GetByPath(params)
		if err != nil && op.Errors == nil {
			return err
		}
		op.Entrypoint = ep
	} else {
		op.Entrypoint = "default"
	}

	return nil
}

func (p *DefaultParser) parseInternalOperations(item gjson.Result, main models.Operation) ([]models.Operation, []models.Contract, error) {
	path := fmt.Sprintf("metadata.internal_operation_results")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("metadata.internal_operations")
		if !item.Get(path).Exists() {
			return nil, nil, nil
		}
	}

	res := make([]models.Operation, 0)
	contracts := make([]models.Contract, 0)
	for i, op := range item.Get(path).Array() {
		internalOperation, contract, err := p.parseContent(op, main.Network, main.Protocol)
		if err != nil {
			return nil, nil, err
		}
		if contract != nil {
			contracts = append(contracts, *contract)
		}

		internalOperation.Counter = main.Counter
		internalOperation.Hash = main.Hash
		internalOperation.Level = main.Level
		internalOperation.Timestamp = main.Timestamp
		internalOperation.IndexedTime = time.Now().UnixNano() / 1000
		internalOperation.Internal = true
		internalOperation.InternalIndex = int64(i + 1)

		if err := p.finishParseOperation(op, &internalOperation); err != nil {
			return nil, nil, err
		}
		res = append(res, internalOperation)
	}
	return res, contracts, nil
}

func (p *DefaultParser) needParse(item gjson.Result, idx int) bool {
	kind := item.Get("kind").String()
	originationCondition := kind == consts.Origination && item.Get("script").Exists()
	prefixCondition := strings.HasPrefix(item.Get("source").String(), "KT") || strings.HasPrefix(item.Get("destination").String(), "KT")
	transactionCondition := kind == consts.Transaction && prefixCondition
	return originationCondition || transactionCondition
}

func (p *DefaultParser) getRichStorage(data gjson.Result, metadata meta.Metadata, op *models.Operation) (storage.RichStorage, error) {
	protoSymLink, ok := p.protocols[op.Protocol]
	if !ok {
		return storage.RichStorage{Empty: true}, fmt.Errorf("Unknown protocol: %s", op.Protocol)
	}

	var parser storage.Parser
	switch protoSymLink {
	case consts.MetadataBabylon:
		parser = storage.NewBabylon(p.rpc)
	case consts.MetadataAlpha:
		parser = storage.NewAlpha()
	default:
		return storage.RichStorage{Empty: true}, fmt.Errorf("Unknown protocol: %s", op.Protocol)
	}

	switch op.Kind {
	case consts.Transaction:
		return parser.ParseTransaction(data, metadata, op.Level, op.ID)
	case consts.Origination:
		return parser.ParseOrigination(data, metadata, op.Level, op.ID)
	}
	return storage.RichStorage{Empty: true}, nil
}
