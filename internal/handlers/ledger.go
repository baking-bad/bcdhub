package handlers

import (
	"fmt"
	"strings"
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
	"github.com/tidwall/gjson"
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
	storage   models.GeneralRepository
	sharePath string

	cache *ccache.Cache
}

// NewLedger -
func NewLedger(storage models.GeneralRepository, sharePath string) *Ledger {
	return &Ledger{
		storage:   storage,
		cache:     ccache.New(ccache.Configure().MaxSize(100)),
		sharePath: sharePath,
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
	balance, err := ledger.getTokenBalance(bmd, bigMapType)
	if err != nil {
		if errors.Is(err, tokenbalance.ErrUnknownParser) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, []models.Model{balance}, nil
}

func (ledger *Ledger) getTokenBalance(bmd *bigmapdiff.BigMapDiff, bigMapType []byte) (*tbModel.TokenBalance, error) {
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

	return &tbModel.TokenBalance{
		Network:  bmd.Network,
		Address:  balance.Address,
		TokenID:  balance.TokenID,
		Contract: bmd.Address,
		Value:    balance.Value,
	}, nil
}

func (ledger *Ledger) buildElt(bmd *bigmapdiff.BigMapDiff) (gjson.Result, error) {
	b, err := json.Marshal(bmd.Key)
	if err != nil {
		return gjson.Result{}, err
	}

	var s strings.Builder
	s.WriteString(`{"prim":"Elt","args":[`)
	if _, err := s.Write(b); err != nil {
		return gjson.Result{}, err
	}
	s.WriteByte(',')
	if len(bmd.ValueBytes()) != 0 {
		if _, err := s.Write(bmd.ValueBytes()); err != nil {
			return gjson.Result{}, err
		}
	} else {
		s.WriteString(`{"int":"0"}`)
	}
	s.WriteString(`]}`)
	return gjson.Parse(s.String()), nil
}

func (ledger *Ledger) findLedgerBigMap(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	data, err := fetch.Contract(bmd.Address, bmd.Network, bmd.Protocol, ledger.sharePath)
	if err != nil {
		return nil, err
	}
	script := gjson.ParseBytes(data)

	storageType := script.Get(`#(prim=="storage").args.0`)
	tree, err := ast.NewTypedAstFromString(storageType.Raw)
	if err != nil {
		return nil, err
	}

	node := tree.FindByName(ledgerStorageKey)
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
