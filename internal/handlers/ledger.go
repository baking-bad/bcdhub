package handlers

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ledgerStorageKey = "ledger"
)

// errors
var (
	ErrNoLedgerKeyInStorage = errors.New("No ledger key in storage")
	ErrNoRPCNetwork         = errors.New("Unknown rpc")
)

// Ledger -
type Ledger struct {
	storage       models.GeneralRepository
	tokenBalances tbModel.Repository
	operations    operation.Repository
	sharePath     string
}

// NewLedger -
func NewLedger(storage models.GeneralRepository, opRepo operation.Repository, tokenBalanceRepo tbModel.Repository, sharePath string) *Ledger {
	return &Ledger{
		storage:       storage,
		operations:    opRepo,
		tokenBalances: tokenBalanceRepo,
		sharePath:     sharePath,
	}
}

// Do -
func (ledger *Ledger) Do(bmd *bigmapdiff.BigMapDiff) (bool, []models.Model, error) {
	bigMapType, err := ledger.findLedgerBigMap(bmd)
	if err != nil {
		if errors.Is(err, ErrNoLedgerKeyInStorage) {
			return false, nil, nil
		}
		return false, nil, err
	}
	if bigMapType == nil {
		return false, nil, nil
	}

	success, models, err := ledger.handle(bmd, bigMapType)
	if err != nil {
		return false, nil, err
	}
	if success {
		return success, nil, ledger.tokenBalances.Update(models)
	}
	return success, nil, nil
}

func (ledger *Ledger) handle(bmd *bigmapdiff.BigMapDiff, bigMapType *ast.BigMap) (bool, []*tbModel.TokenBalance, error) {
	balances, err := ledger.getResultModels(bmd, bigMapType)
	if err != nil {
		if errors.Is(err, tokenbalance.ErrUnknownParser) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, balances, nil
}

func (ledger *Ledger) getResultModels(bmd *bigmapdiff.BigMapDiff, bigMapType *ast.BigMap) ([]*tbModel.TokenBalance, error) {
	parser, err := tokenbalance.GetParserForBigMap(bigMapType)
	if err != nil {
		return nil, err
	}
	elt, err := ledger.buildElt(bmd)
	if err != nil {
		return nil, err
	}
	balance, err := parser.Parse(elt)
	if err != nil {
		return nil, err
	}
	if len(balance) == 0 {
		return nil, nil
	}

	tb := &tbModel.TokenBalance{
		Network:  bmd.Network,
		Address:  balance[0].Address,
		TokenID:  balance[0].TokenID,
		Contract: bmd.Contract,
		Value:    balance[0].Value,
	}

	items := []models.Model{tb}

	if balance[0].IsExclusiveNFT {
		holders, err := ledger.tokenBalances.NFTHolders(tb.Network, tb.Contract, tb.TokenID)
		if err != nil {
			return nil, err
		}
		for i := range holders {
			holders[i].Value.SetInt64(0)
			items = append(items, &holders[i])
		}
	}

	return items, nil
}

func (ledger *Ledger) buildElt(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	var s bytes.Buffer
	s.WriteString(`[{"prim":"Elt","args":[`)
	if _, err := s.Write(bmd.Key); err != nil {
		return nil, err
	}
	s.WriteByte(',')
	if len(bmd.ValueBytes()) != 0 {
		if _, err := s.Write(bmd.ValueBytes()); err != nil {
			return nil, err
		}
	} else {
		s.WriteString(`{"int":"0"}`)
	}
	s.WriteString(`]}]`)
	return s.Bytes(), nil
}

func (ledger *Ledger) findLedgerBigMap(bmd *bigmapdiff.BigMapDiff) (*ast.BigMap, error) {
	data, err := fetch.Contract(bmd.Contract, bmd.Network, bmd.Protocol, ledger.sharePath)
	if err != nil {
		return nil, err
	}
	var script ast.Script
	if err := json.Unmarshal(data, &script); err != nil {
		return nil, err
	}
	tree, err := script.StorageType()
	if err != nil {
		return nil, err
	}

	node := tree.FindByName(ledgerStorageKey, false)
	if err != nil {
		return nil, ErrNoLedgerKeyInStorage
	}

	op, err := ledger.operations.GetOne(bmd.OperationHash, bmd.OperationCounter, bmd.OperationNonce)
	if err != nil {
		if ledger.storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var storageData ast.UntypedAST
	if err := json.Unmarshal(op.DeffatedStorage, &storageData); err != nil {
		return nil, err
	}
	if err := tree.Settle(storageData); err != nil {
		return nil, err
	}

	bigMap, ok := node.(*ast.BigMap)
	if !ok {
		return nil, ErrNoLedgerKeyInStorage
	}
	if *bigMap.Ptr != bmd.Ptr {
		return nil, ErrNoLedgerKeyInStorage
	}
	return bigMap, nil
}
