package ledger

import (
	"bytes"

	"github.com/shopspring/decimal"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/pkg/errors"
)

const (
	ledgerStorageKey = "ledger"
)

// errors
var (
	ErrNoLedgerKeyInStorage = errors.New("No ledger key in storage")
)

// Ledger -
type Ledger struct {
	tokenBalances tbModel.Repository
}

// New -
func New(tokenBalanceRepo tbModel.Repository) *Ledger {
	return &Ledger{
		tokenBalances: tokenBalanceRepo,
	}
}

// Do -
func (ledger *Ledger) Parse(operation *operation.Operation, st *stacktrace.StackTrace) (*parsers.Result, error) {
	if operation == nil || len(operation.BigMapDiffs) == 0 || !operation.Tags.Has(types.LedgerTag) {
		return nil, nil
	}

	isFATransfer := (operation.Tags.Has(types.FA2Tag) || operation.Tags.Has(types.FA12Tag)) && operation.IsEntrypoint(consts.TransferEntrypoint)
	if isFATransfer {
		return nil, nil
	}

	storage, err := operation.AST.Storage.ToTypedAST()
	if err != nil {
		return nil, err
	}

	if err := storage.SettleFromBytes(operation.DeffatedStorage); err != nil {
		return nil, err
	}

	bigMap, err := ledger.getLedgerBigMap(storage)
	switch {
	case errors.Is(err, ErrNoLedgerKeyInStorage):
		return nil, nil
	case err != nil:
		return nil, err
	case bigMap == nil:
		return nil, nil
	}

	result := new(parsers.Result)
	for _, bmd := range operation.BigMapDiffs {
		if bmd.Ptr != *bigMap.Ptr {
			continue
		}
		balances, err := ledger.handle(bmd, bigMap, st, operation)
		if err != nil {
			return nil, err
		}
		result.TokenBalances = append(result.TokenBalances, balances...)
	}
	return result, nil
}

func (ledger *Ledger) handle(bmd *bigmapdiff.BigMapDiff, bigMapType *ast.BigMap, st *stacktrace.StackTrace, op *operation.Operation) ([]*tbModel.TokenBalance, error) {
	balances, err := ledger.getResultModels(bmd, bigMapType, st, op)
	switch {
	case errors.Is(err, tokenbalance.ErrUnknownParser):
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return balances, nil
	}
}

func (ledger *Ledger) getResultModels(bmd *bigmapdiff.BigMapDiff, bigMapType *ast.BigMap, st *stacktrace.StackTrace, op *operation.Operation) ([]*tbModel.TokenBalance, error) {
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
		Balance:  balance[0].Value,
		IsLedger: true,
	}

	balances := []*tbModel.TokenBalance{tb}

	t := ledger.makeTransfer(balance[0], st, op)
	if t != nil {
		op.Transfers = append(op.Transfers, t)

		if balance[0].IsExclusiveNFT {
			holders, err := ledger.tokenBalances.GetHolders(tb.Network, tb.Contract, tb.TokenID)
			if err != nil {
				return nil, err
			}
			for i := range holders {
				holders[i].Balance = decimal.Zero
				holders[i].IsLedger = true
				t.From = holders[i].Address
				balances = append(balances, &holders[i])
			}
		}
	}

	return balances, nil
}

func (ledger *Ledger) makeTransfer(tb tokenbalance.TokenBalance, st *stacktrace.StackTrace, op *operation.Operation) *transfer.Transfer {

	balance, err := ledger.tokenBalances.Get(op.Network, op.Destination, tb.Address, tb.TokenID)
	if err != nil {
		logger.Error(err)
		return nil
	}

	t := op.EmptyTransfer()

	switch balance.Balance.Cmp(tb.Value) {
	case 1:
		t.From = tb.Address
	case -1:
		t.To = tb.Address
	default:
		return nil
	}

	t.Amount = tb.Value.Sub(balance.Balance).Abs()
	t.TokenID = tb.TokenID

	if op.Nonce != nil && st != nil && !st.Empty() {
		item := st.Get(*op)
		if item != nil && item.ParentID > -1 {
			parent := st.GetByID(item.ParentID)
			if parent != nil {
				t.Parent = parent.Entrypoint
			}
		}
	}

	return t
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

func (ledger *Ledger) getLedgerBigMap(storage *ast.TypedAst) (*ast.BigMap, error) {
	node := storage.FindByName(ledgerStorageKey, false)
	if node == nil {
		return nil, ErrNoLedgerKeyInStorage
	}
	bigMap, ok := node.(*ast.BigMap)
	if !ok {
		return nil, ErrNoLedgerKeyInStorage
	}
	if bigMap.Ptr == nil {
		return nil, ErrNoLedgerKeyInStorage
	}
	return bigMap, nil
}
