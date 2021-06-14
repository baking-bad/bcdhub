package ledger

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	mock_token_balance "github.com/baking-bad/bcdhub/internal/models/mock/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newDecimal(val string) decimal.Decimal {
	i, _ := decimal.NewFromString(val)
	return i
}

func TestLedger_Parse(t *testing.T) {
	ctrlTokenBalanceRepo := gomock.NewController(t)
	defer ctrlTokenBalanceRepo.Finish()
	tbRepo := mock_token_balance.NewMockRepository(ctrlTokenBalanceRepo)

	tbRepo.
		EXPECT().
		Get(
			gomock.Eq(types.Mainnet),
			gomock.Eq("KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E"),
			gomock.Eq("tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo"),
			gomock.Eq(uint64(0))).
		Return(tbModel.TokenBalance{
			Contract: "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
			Network:  types.Mainnet,
			Address:  "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
			TokenID:  0,
			Balance:  newDecimal("168000"),
		}, nil).
		AnyTimes()

	tests := []struct {
		name          string
		operation     *operation.Operation
		st            *stacktrace.StackTrace
		want          *parsers.Result
		wantTransfers []*transfer.Transfer
		wantErr       bool
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
				Tags:            types.FA12Tag,
				Entrypoint:      consts.TransferEntrypoint,
				Network:         types.Mainnet,
				DeffatedStorage: []byte(`{"int":257}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         257,
						Key:         []byte(`{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}`),
						KeyHash:     "expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf",
						KeyStrings:  []string{"tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"},
						Value:       []byte(`{"int":"1000000"}`),
						OperationID: 1,
						Level:       1269694,
						Contract:    "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
						Network:     types.Mainnet,
						Timestamp:   time.Date(2020, 12, 22, 19, 19, 49, 0, time.UTC),
						ProtocolID:  1,
					},
				},
			},
			want: nil,
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Tags:            types.FA12Tag | types.LedgerTag,
				Entrypoint:      consts.TransferEntrypoint,
				Network:         types.Mainnet,
				DeffatedStorage: []byte(`{"int":257}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         257,
						Key:         []byte(`{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}`),
						KeyHash:     "expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf",
						KeyStrings:  []string{"tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"},
						Value:       []byte(`{"int":"1000000"}`),
						OperationID: 1,
						Level:       1269694,
						Contract:    "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
						Network:     types.Mainnet,
						Timestamp:   time.Date(2020, 12, 22, 19, 19, 49, 0, time.UTC),
						ProtocolID:  1,
					},
				},
			},
			want: nil,
		}, {
			name: "test 3",
			operation: &operation.Operation{
				Tags:            types.FA2Tag | types.LedgerTag,
				Entrypoint:      "burn",
				Kind:            types.OperationKindTransaction,
				Network:         types.Mainnet,
				Destination:     "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
				Hash:            "opNQeUBKfJzBjCNLuo5HkyZynhm5TMe1KEtwioqUrWM1ygmYVDX",
				Level:           1455291,
				DeffatedStorage: []byte(`{"int":2071}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         2071,
						Key:         []byte(`{"bytes":"00016631ce723071ea19a87bd93d7e2f81dd82c18565"}`),
						KeyHash:     "exprvA4NaRxQEqyJad5LzVz7rohGSQDn9B32KR87KxRQ43kWiaqPS4",
						KeyStrings:  []string{"tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo"},
						Value:       []byte(``),
						OperationID: 1,
						Level:       1455291,
						Contract:    "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
						Network:     types.Mainnet,
						Timestamp:   time.Date(2021, 05, 03, 10, 03, 20, 0, time.UTC),
						ProtocolID:  2,
					},
				},
			},
			want: &parsers.Result{
				TokenBalances: []*tbModel.TokenBalance{
					{
						Address:  "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
						Contract: "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
						Network:  types.Mainnet,
						Balance:  newDecimal("0"),
						TokenID:  0,
						IsLedger: true,
					},
				},
			},
			wantTransfers: []*transfer.Transfer{
				{
					From:       "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
					Contract:   "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
					Network:    types.Mainnet,
					Amount:     newDecimal("168000"),
					TokenID:    0,
					Entrypoint: "burn",
					Level:      1455291,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage, err := ast.NewUntypedAST([]byte(`{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`))
			if err != nil {
				t.Errorf("ast.NewUntypedAST() error = %v", err)
				return
			}
			tt.operation.AST = &ast.Script{
				Storage: storage,
			}

			ledger := &Ledger{
				tokenBalances: tbRepo,
			}
			got, err := ledger.Parse(tt.operation, tt.st)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ledger.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)

			assert.Len(t, tt.operation.Transfers, len(tt.wantTransfers))
			assert.Equal(t, tt.wantTransfers, tt.operation.Transfers)
		})
	}
}
