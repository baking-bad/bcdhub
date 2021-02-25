package operations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

	txMetadata := parseMetadata(data)
	tx.Result = &txMetadata.Result
	tx.Status = tx.Result.Status
	tx.Errors = tx.Result.Errors

	tx.SetBurned(p.constants)
	txModels := []models.Model{&tx}

	if tx.IsApplied() {
		appliedModels, err := p.appliedHandler(data, &tx)
		if err != nil {
			return nil, err
		}
		txModels = append(txModels, appliedModels...)
	}

	if len(tx.Parameters) > 0 && !tezerrors.HasParametersError(tx.Errors) {
		if err := p.getEntrypoint(&tx); err != nil {
			return nil, err
		}
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
	if !bcd.IsContract(op.Destination) || !op.IsApplied() {
		return nil, nil
	}

	script, err := fetch.Contract(op.Destination, op.Network, op.Protocol, p.shareDir)
	if err != nil {
		return nil, err
	}
	op.Script = gjson.ParseBytes(script)

	resultModels := make([]models.Model, 0)

	rs, err := p.storageParser.Parse(item, op)
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	op.DeffatedStorage = rs.DeffatedStorage

	resultModels = append(resultModels, rs.Models...)

	migration, err := NewMigration().Parse(item, op)
	if err != nil {
		return nil, err
	}
	if migration != nil {
		resultModels = append(resultModels, migration)
	}

	return resultModels, nil
}

func (p Transaction) getEntrypoint(op *operation.Operation) error {
	if op.Script.Raw == "" {
		script, err := fetch.Contract(op.Destination, op.Network, op.Protocol, p.shareDir)
		if err != nil {
			return err
		}
		op.Script = gjson.ParseBytes(script)
	}

	parameter := op.GetScriptSection(consts.PARAMETER)
	param, err := ast.NewTypedAstFromString(parameter.Raw)
	if err != nil {
		return err
	}

	params := types.NewParameters([]byte(op.Parameters))

	subTree, err := param.FromParameters(params)
	if err != nil {
		return err
	}
	op.Entrypoint = params.Entrypoint

	if len(subTree.Nodes) == 0 {
		return nil
	}

	op.Entrypoint = subTree.Nodes[0].GetName()

	return nil
}

func (p Transaction) tagTransaction(tx *operation.Operation) error {
	if !bcd.IsContract(tx.Destination) {
		return nil
	}

	c := contract.NewEmptyContract(tx.Network, tx.Destination)
	if err := p.Storage.GetByID(&c); err != nil {
		if p.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	tx.Tags = make([]string, 0)
	for _, tag := range c.Tags {
		if helpers.StringInArray(tag, []string{
			consts.FA12Tag, consts.FA2Tag,
		}) {
			tx.Tags = append(tx.Tags, tag)
		}
	}
	return nil
}
