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
	"github.com/baking-bad/bcdhub/internal/logger"
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
func (p *DefaultParser) Parse(opg gjson.Result, network string, head noderpc.Header) ([]models.Operation, []models.Contract, error) {
	operations := make([]models.Operation, 0)
	contracts := make([]models.Contract, 0)
	for idx, item := range opg.Get("contents").Array() {
		if !p.needParse(item, idx) {
			continue
		}

		hash := opg.Get("hash").String()
		op, contract, err := p.parseContent(item, network, hash, head)
		if err != nil {
			return nil, nil, err
		}

		operations = append(operations, op)
		if contract != nil {
			contracts = append(contracts, *contract)
		}

		internal, internalContracts, err := p.parseInternalOperations(item, op, head)
		if err != nil {
			return nil, nil, err
		}
		operations = append(operations, internal...)
		contracts = append(contracts, internalContracts...)
	}
	return operations, contracts, nil
}

func (p *DefaultParser) parseContent(data gjson.Result, network, hash string, head noderpc.Header) (models.Operation, *models.Contract, error) {
	kind := data.Get("kind").String()
	switch kind {
	case consts.Origination:
		return p.parseOrigination(data, network, hash, head)
	default:
		op, err := p.parseTransaction(data, network, hash, head)
		return op, nil, err
	}
}

func (p *DefaultParser) parseTransaction(data gjson.Result, network, hash string, head noderpc.Header) (op models.Operation, err error) {
	op.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	op.Network = network
	op.Hash = hash
	op.Protocol = head.Protocol
	op.Level = head.Level
	op.Timestamp = head.Timestamp
	op.Kind = data.Get("kind").String()
	op.Source = data.Get("source").String()
	op.Fee = data.Get("fee").Int()
	op.Counter = data.Get("counter").Int()
	op.GasLimit = data.Get("gas_limit").Int()
	op.StorageLimit = data.Get("storage_limit").Int()
	op.Amount = data.Get("amount").Int()
	op.Destination = data.Get("destination").String()
	op.PublicKey = data.Get("public_key").String()
	op.Balance = data.Get("balance").Int()
	op.ManagerPubKey = data.Get("manager_pubkey").String()
	op.Delegate = data.Get("delegate").String()
	op.Parameters = data.Get("parameters").String()
	op.BalanceUpdates = p.parseBalanceUpdates(data, "metadata")
	op.IndexedTime = time.Now().UnixNano() / 1000

	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors

	err = p.finishParseOperation(data, &op)
	return
}

func (p *DefaultParser) parseOrigination(data gjson.Result, network, hash string, head noderpc.Header) (models.Operation, *models.Contract, error) {
	op := models.Operation{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
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
		Balance:        data.Get("balance").Int(),
		ManagerPubKey:  data.Get("manager_pubkey").String(),
		Delegate:       data.Get("delegate").String(),
		Parameters:     data.Get("parameters").String(),
		Script:         data.Get("script"),
		BalanceUpdates: p.parseBalanceUpdates(data, "metadata"),
		IndexedTime:    time.Now().UnixNano() / 1000,
	}

	operationResult, balanceUpdates := p.parseMetadata(data)
	op.Result = operationResult
	op.BalanceUpdates = append(op.BalanceUpdates, balanceUpdates...)
	op.Status = op.Result.Status
	op.Errors = op.Result.Errors
	op.Destination = operationResult.Originated

	protoSymLink, err := meta.GetProtoSymLink(op.Protocol)
	if err != nil {
		return op, nil, err
	}

	var contract *models.Contract
	if !contractparser.IsDelegateContract(op.Script) && p.isApplied(op) {
		contract, err = createNewContract(p.es, op, p.filesDirectory, protoSymLink)
	}
	err = p.finishParseOperation(data, &op)
	return op, contract, err
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
	if !strings.HasPrefix(op.Destination, "KT") {
		return nil
	}

	metadata, err := meta.GetContractMetadata(p.es, op.Destination)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil
		}
		return err
	}
	if p.isApplied(*op) {
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

		if op.Kind == consts.Transaction {
			if err := p.findMigration(item, op); err != nil {
				return err
			}
		}
	}
	if op.Kind == consts.Transaction {
		return p.getEntrypoint(item, metadata, op)
	}
	return nil
}

func (p *DefaultParser) findMigration(item gjson.Result, op *models.Operation) error {
	path := fmt.Sprintf("metadata.operation_result.big_map_diff")
	if !item.Get(path).Exists() {
		path = fmt.Sprintf("result.big_map_diff")
		if !item.Get(path).Exists() {
			return nil
		}
	}
	for _, bmd := range item.Get(path).Array() {
		if bmd.Get("action").String() != "update" {
			continue
		}

		value := bmd.Get("value")
		if contractparser.HasLambda(value) {
			logger.Info("[%s] Migration detected: %s", op.Network, op.Destination)
			migration := models.Migration{
				ID:          strings.ReplaceAll(uuid.New().String(), "-", ""),
				IndexedTime: time.Now().UnixNano() / 1000,

				Network:   op.Network,
				Level:     op.Level,
				Protocol:  op.Protocol,
				Address:   op.Destination,
				Timestamp: op.Timestamp,
				Hash:      op.Hash,
			}
			if _, err := p.es.AddDocumentWithID(migration, elastic.DocMigrations, migration.ID); err != nil {
				return err
			}
			break
		}
	}
	return nil
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

func (p *DefaultParser) parseInternalOperations(item gjson.Result, main models.Operation, head noderpc.Header) ([]models.Operation, []models.Contract, error) {
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
		internalOperation, contract, err := p.parseContent(op, main.Network, main.Hash, head)
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
		internalOperation.Internal = true
		internalOperation.InternalIndex = int64(i + 1)
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

func (p *DefaultParser) getRichStorage(data gjson.Result, metadata *meta.ContractMetadata, op *models.Operation) (storage.RichStorage, error) {
	protoSymLink, err := meta.GetProtoSymLink(op.Protocol)
	if err != nil {
		return storage.RichStorage{Empty: true}, err
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

	m, ok := metadata.Storage[protoSymLink]
	if !ok {
		return storage.RichStorage{Empty: true}, fmt.Errorf("Unknown metadata: %s", protoSymLink)
	}
	switch op.Kind {
	case consts.Transaction:
		return parser.ParseTransaction(data, m, *op)
	case consts.Origination:
		return parser.ParseOrigination(data, m, *op)
	}
	return storage.RichStorage{Empty: true}, nil
}
