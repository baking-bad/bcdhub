package operations

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/normalize"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Transaction -
type Transaction struct {
	*ParseParams
}

// NewTransaction -
func NewTransaction(params *ParseParams) Transaction {
	return Transaction{params}
}

// Parse -
func (p Transaction) Parse(data gjson.Result) ([]models.Model, error) {
	source := data.Get("source").String()
	tx := operation.Operation{
		ID:            helpers.GenerateID(),
		Network:       p.network,
		Hash:          p.hash,
		Protocol:      p.head.Protocol,
		Level:         p.head.Level,
		Timestamp:     p.head.Timestamp,
		Kind:          data.Get("kind").String(),
		Initiator:     source,
		Source:        source,
		Fee:           data.Get("fee").Int(),
		Counter:       data.Get("counter").Int(),
		GasLimit:      data.Get("gas_limit").Int(),
		StorageLimit:  data.Get("storage_limit").Int(),
		Amount:        data.Get("amount").Int(),
		Destination:   data.Get("destination").String(),
		PublicKey:     data.Get("public_key").String(),
		ManagerPubKey: data.Get("manager_pubkey").String(),
		Delegate:      data.Get("delegate").String(),
		Parameters:    data.Get("parameters").String(),
		IndexedTime:   time.Now().UnixNano() / 1000,
		ContentIndex:  p.contentIdx,
	}

	p.fillInternal(&tx)

	if data.Get("nonce").Exists() {
		nonce := data.Get("nonce").Int()
		tx.Nonce = &nonce
	}

	txMetadata := parseMetadata(data, tx)
	tx.Result = &txMetadata.Result
	tx.Status = tx.Result.Status
	tx.Errors = tx.Result.Errors

	tx.SetBurned(p.constants)
	txModels := []models.Model{&tx}

	if tx.IsApplied() {
		for i := range txMetadata.BalanceUpdates {
			txModels = append(txModels, txMetadata.BalanceUpdates[i])
		}

		appliedModels, err := p.appliedHandler(data, &tx)
		if err != nil {
			return nil, err
		}
		txModels = append(txModels, appliedModels...)
	}

	if err := p.tagTransaction(&tx); err != nil {
		return nil, err
	}

	p.stackTrace.Add(tx)

	transfers, err := p.transferParser.Parse(tx, txModels)
	if err != nil {
		if !errors.Is(err, events.ErrNodeReturn) {
			return nil, err
		}
		logger.Error(err)
	}
	for i := range transfers {
		txModels = append(txModels, transfers[i])
	}

	if err := transferParsers.UpdateTokenBalances(p.TokenBalances, transfers); err != nil {
		return nil, err
	}

	return txModels, nil
}

func (p Transaction) normalizeOperationData(contract gjson.Result, operation *operation.Operation) error {
	params := gjson.Parse(operation.Parameters)

	var value gjson.Result
	var entrypointName string

	entrypoint := params.Get("entrypoint")
	if entrypoint.Exists() {
		value = params.Get("value")
		entrypointName = entrypoint.String()
	} else {
		value = params
		entrypointName = consts.DefaultEntrypoint
	}

	paramsType := contract.Get("code.#(prim==\"parameter\").args.0")
	typ, err := findByFieldName(entrypointName, paramsType)
	if err != nil && !errors.Is(err, ErrAnnotsIsNotFound) {
		return err
	}
	data, err := normalize.Data(value, typ)
	if err != nil {
		return err
	}
	if entrypoint.Exists() {
		p, err := sjson.SetRaw(params.Raw, "value", data.Raw)
		if err != nil {
			return err
		}
		operation.Parameters = p
	} else {
		operation.Parameters = data.String()
	}
	return nil
}

// errors
var (
	ErrInvalidJSONType  = errors.New("Invalid JSON type")
	ErrAnnotsIsNotFound = errors.New("Annot is not found")
)

func findByFieldName(fieldName string, data gjson.Result) (gjson.Result, error) {
	switch {
	case data.IsArray():
		for _, item := range data.Array() {
			res, err := findByFieldName(fieldName, item)
			if err != nil {
				if errors.Is(err, ErrAnnotsIsNotFound) {
					continue
				}
				return gjson.Result{}, err
			}
			return res, nil
		}
	case data.IsObject():
		for _, item := range data.Get("annots").Array() {
			if item.String() == fmt.Sprintf("%%%s", fieldName) {
				return data, nil
			}
		}
		args := data.Get("args")
		if args.Exists() {
			return findByFieldName(fieldName, args)
		}
	default:
		return data, ErrInvalidJSONType
	}
	return data, ErrAnnotsIsNotFound
}

func (p Transaction) fillInternal(tx *operation.Operation) {
	if p.main == nil {
		p.main = tx
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}

func (p Transaction) appliedHandler(item gjson.Result, op *operation.Operation) ([]models.Model, error) {
	if !helpers.IsContract(op.Destination) || !op.IsApplied() {
		return nil, nil
	}

	schema, err := meta.GetContractSchema(p.Schema, op.Destination)
	if err != nil {
		if p.Storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	contract, err := contractparser.GetContract(p.rpc, op.Destination, op.Network, op.Protocol, p.shareDir, op.Level)
	if err != nil {
		return nil, err
	}

	resultModels := make([]models.Model, 0)

	rs, err := p.storageParser.Parse(item, contract, schema, op)
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	op.DeffatedStorage = rs.DeffatedStorage

	resultModels = append(resultModels, rs.Models...)

	migration := NewMigration(op).Parse(item)
	if migration != nil {
		resultModels = append(resultModels, migration)
	}

	if err := p.normalizeOperationData(contract, op); err != nil {
		return nil, err
	}

	bu := NewBalanceUpdate("metadata", *op).Parse(item)
	for i := range bu {
		resultModels = append(resultModels, bu[i])
	}
	return resultModels, p.getEntrypoint(item, schema, op)
}

func (p Transaction) getEntrypoint(item gjson.Result, metadata *meta.ContractSchema, op *operation.Operation) error {
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
		op.Entrypoint = consts.DefaultEntrypoint
	}

	return nil
}

func (p Transaction) tagTransaction(tx *operation.Operation) error {
	if !helpers.IsContract(tx.Destination) {
		return nil
	}

	contract := contract.NewEmptyContract(tx.Network, tx.Destination)
	if err := p.Storage.GetByID(&contract); err != nil {
		if p.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	tx.Tags = make([]string, 0)
	for _, tag := range contract.Tags {
		if helpers.StringInArray(tag, []string{
			consts.FA12Tag, consts.FA2Tag,
		}) {
			tx.Tags = append(tx.Tags, tag)
		}
	}
	return nil
}
