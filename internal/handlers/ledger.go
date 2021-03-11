package handlers

import (
	"bytes"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/karlseguin/ccache"
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
	sharePath     string

	cache *ccache.Cache
}

// NewLedger -
func NewLedger(storage models.GeneralRepository, tokenBalanceRepo tbModel.Repository, sharePath string) *Ledger {
	return &Ledger{
		storage:       storage,
		tokenBalances: tokenBalanceRepo,
		cache:         ccache.New(ccache.Configure().MaxSize(100)),
		sharePath:     sharePath,
	}
}

// Do -
func (ledger *Ledger) Do(model models.Model) (bool, []models.Model, error) {
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return false, nil, nil
	}

	bigMapType, err := ledger.getCachedBigMapType(bmd)
	if err != nil {
		return false, nil, err
	}
	if bigMapType == nil {
		return false, nil, nil
	}

	return ledger.handle(bmd, bigMapType)
}

func (ledger *Ledger) getCachedBigMapType(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	item, err := ledger.cache.Fetch(fmt.Sprintf("%s:%d", bmd.Network, bmd.Ptr), time.Minute*10, func() (interface{}, error) {
		return ledger.findLedgerBigMap(bmd)
	})
	if err != nil {
		if errors.Is(err, ErrNoLedgerKeyInStorage) {
			return nil, nil
		}
		return nil, err
	}
	return item.Value().([]byte), nil
}

func (ledger *Ledger) handle(bmd *bigmapdiff.BigMapDiff, bigMapType []byte) (bool, []models.Model, error) {
	balances, err := ledger.getResultModels(bmd, bigMapType)
	if err != nil {
		if errors.Is(err, tokenbalance.ErrUnknownParser) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, balances, nil
}

func (ledger *Ledger) getResultModels(bmd *bigmapdiff.BigMapDiff, bigMapType []byte) ([]models.Model, error) {
	typ, err := ast.NewTypedAstFromBytes(bigMapType)
	if err != nil {
		return nil, err
	}
	parser, err := tokenbalance.GetParserForBigMap(typ)
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
	if balance[0].IsNFT {
		if err := ledger.tokenBalances.BurnNft(bmd.Network, bmd.Address, balance[0].TokenID); err != nil {
			return nil, err
		}
		if balance[0].Address == "" { // Burn NFT token
			return nil, nil
		}
	}

	return []models.Model{&tbModel.TokenBalance{
		Network:  bmd.Network,
		Address:  balance[0].Address,
		TokenID:  balance[0].TokenID,
		Contract: bmd.Address,
		Value:    balance[0].Value,
	}}, nil
}

func (ledger *Ledger) buildElt(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	b, err := json.Marshal(bmd.Key)
	if err != nil {
		return nil, err
	}

	var s bytes.Buffer
	s.WriteString(`[{"prim":"Elt","args":[`)
	if _, err := s.Write(b); err != nil {
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

func (ledger *Ledger) findLedgerBigMap(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	data, err := fetch.Contract(bmd.Address, bmd.Network, bmd.Protocol, ledger.sharePath)
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

	op := operation.Operation{ID: bmd.OperationID}
	if err := ledger.storage.GetByID(&op); err != nil {
		return nil, err
	}

	var storageData ast.UntypedAST
	if err := json.UnmarshalFromString(op.DeffatedStorage, &storageData); err != nil {
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
	return json.Marshal(bigMap)
}
