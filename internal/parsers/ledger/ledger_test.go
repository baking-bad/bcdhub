package ledger

import (
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	mock_accounts "github.com/baking-bad/bcdhub/internal/models/mock/account"
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

func TestLedger_Parse(t *testing.T) {
	ctrlTokenBalanceRepo := gomock.NewController(t)
	defer ctrlTokenBalanceRepo.Finish()
	tbRepo := mock_token_balance.NewMockRepository(ctrlTokenBalanceRepo)

	ctrlAccountsRepo := gomock.NewController(t)
	defer ctrlAccountsRepo.Finish()
	accountsRepo := mock_accounts.NewMockRepository(ctrlAccountsRepo)

	accountsRepo.
		EXPECT().
		Get(gomock.Eq("tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo")).
		Return(account.Account{
			Address: "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
			Type:    types.AccountTypeTz,
		}, nil).
		AnyTimes()

	tbRepo.
		EXPECT().
		Get(
			gomock.Eq("KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E"),
			gomock.Eq(int64(0)),
			gomock.Eq(uint64(0))).
		Return(tbModel.TokenBalance{
			Contract: "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
			Account: account.Account{
				Address: "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
				Type:    types.AccountTypeTz,
			},
			TokenID: 0,
			Balance: decimal.RequireFromString("168000"),
		}, nil).
		AnyTimes()

	tests := []struct {
		name          string
		operation     *operation.Operation
		st            *stacktrace.StackTrace
		want          *parsers.TestStore
		wantTransfers []*transfer.Transfer
		wantErr       bool
	}{
		{
			name: "test 1",
			operation: &operation.Operation{
				Tags: types.FA12Tag,
				Entrypoint: types.NullString{
					Str:   consts.TransferEntrypoint,
					Valid: true,
				},
				DeffatedStorage: []byte(`{"int":257}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         257,
						Key:         []byte(`{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}`),
						KeyHash:     "expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf",
						Value:       []byte(`{"int":"1000000"}`),
						OperationID: 1,
						Level:       1269694,
						Contract:    "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
						Timestamp:   time.Date(2020, 12, 22, 19, 19, 49, 0, time.UTC),
						ProtocolID:  1,
					},
				},
			},
			want: parsers.NewTestStore(),
		}, {
			name: "test 2",
			operation: &operation.Operation{
				Tags: types.FA12Tag | types.LedgerTag,
				Entrypoint: types.NullString{
					Str:   consts.TransferEntrypoint,
					Valid: true,
				},
				DeffatedStorage: []byte(`{"int":257}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         257,
						Key:         []byte(`{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"}`),
						KeyHash:     "expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf",
						Value:       []byte(`{"int":"1000000"}`),
						OperationID: 1,
						Level:       1269694,
						Contract:    "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
						Timestamp:   time.Date(2020, 12, 22, 19, 19, 49, 0, time.UTC),
						ProtocolID:  1,
					},
				},
			},
			want: parsers.NewTestStore(),
		}, {
			name: "test 3",
			operation: &operation.Operation{
				Tags: types.FA2Tag | types.LedgerTag,
				Entrypoint: types.NullString{
					Str:   "burn",
					Valid: true,
				},
				Kind: types.OperationKindTransaction,
				Destination: account.Account{
					Address: "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
					Type:    types.AccountTypeContract,
				},
				Hash:            "opNQeUBKfJzBjCNLuo5HkyZynhm5TMe1KEtwioqUrWM1ygmYVDX",
				Level:           1455291,
				DeffatedStorage: []byte(`{"int":2071}`),
				BigMapDiffs: []*bigmapdiff.BigMapDiff{
					{
						Ptr:         2071,
						Key:         []byte(`{"bytes":"00016631ce723071ea19a87bd93d7e2f81dd82c18565"}`),
						KeyHash:     "exprvA4NaRxQEqyJad5LzVz7rohGSQDn9B32KR87KxRQ43kWiaqPS4",
						Value:       []byte(``),
						OperationID: 1,
						Level:       1455291,
						Contract:    "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
						Timestamp:   time.Date(2021, 05, 03, 10, 03, 20, 0, time.UTC),
						ProtocolID:  2,
					},
				},
			},
			want: &parsers.TestStore{
				Contracts:       make([]*contract.Contract, 0),
				BigMapState:     make([]*bigmapdiff.BigMapState, 0),
				Migrations:      make([]*migration.Migration, 0),
				Operations:      make([]*operation.Operation, 0),
				GlobalConstants: make([]*global_constant.GlobalConstant, 0),
				TokenBalances: []*tbModel.TokenBalance{
					{
						Account: account.Account{
							Address: "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
							Type:    types.AccountTypeTz,
						},
						Contract: "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
						Balance:  decimal.RequireFromString("0"),
						TokenID:  0,
						IsLedger: true,
					},
				},
			},
			wantTransfers: []*transfer.Transfer{
				{
					From: account.Account{
						Address: "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
						Type:    types.AccountTypeTz,
					},
					Contract:   "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
					Amount:     decimal.RequireFromString("168000"),
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
				accounts:      accountsRepo,
			}
			store := parsers.NewTestStore()
			if err := ledger.Parse(tt.operation, tt.st, store); (err != nil) != tt.wantErr {
				t.Errorf("Ledger.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, store)

			assert.Len(t, tt.operation.Transfers, len(tt.wantTransfers))
			assert.Equal(t, tt.wantTransfers, tt.operation.Transfers)
		})
	}
}
