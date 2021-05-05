package handlers

import (
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	mock_token_balance "github.com/baking-bad/bcdhub/internal/models/mock/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

func newBigInt(val string) *big.Int {
	i, _ := new(big.Int).SetString(val, 10)
	return i
}

func TestLedger_getResultModels(t *testing.T) {
	ctrlTokenBalanceRepo := gomock.NewController(t)
	defer ctrlTokenBalanceRepo.Finish()
	tbRepo := mock_token_balance.NewMockRepository(ctrlTokenBalanceRepo)

	tbRepo.
		EXPECT().
		Get(
			gomock.Eq("mainnet"),
			gomock.Eq("KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E"),
			gomock.Eq("tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo"),
			gomock.Eq(uint64(0))).
		Return(tbModel.TokenBalance{
			Contract:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
			Network:       "mainnet",
			Address:       "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
			TokenID:       0,
			Balance:       168000,
			BalanceString: "168000",
			Value:         newBigInt("168000"),
		}, nil).
		AnyTimes()

	tests := []struct {
		name       string
		bmd        string
		bigMapType string
		operation  *operation.Operation
		want       []models.Model
		wantErr    bool
	}{
		{
			name:       "test 1",
			bmd:        `{"ptr":257,"key":{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"},"key_hash":"expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf","key_strings":["tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"],"value":{"int":"1000000"},"value_strings":[],"operation_id":"4784c35cc6444b8ca0eb9b7b4698e6cb","level":1269694,"contract":"KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH","network":"mainnet","indexed_time":1612996343064065,"timestamp":"2020-12-22T19:19:49Z","protocol":"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"}`,
			bigMapType: `{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`,
			operation: &operation.Operation{
				Tags:       []string{consts.FA12Tag},
				Entrypoint: consts.TransferEntrypoint,
			},
			want: []models.Model{
				&tbModel.TokenBalance{
					Address:  "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					Contract: "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
					Network:  "mainnet",
					Value:    newBigInt("1000000"),
					TokenID:  0,
					IsLedger: true,
				},
			},
		}, {
			name:       "test 2",
			bmd:        `{"ptr":257,"key":{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"},"key_hash":"expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf","key_strings":["tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"],"value":"","value_strings":[],"operation_id":"4784c35cc6444b8ca0eb9b7b4698e6cb","level":1269694,"contract":"KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH","network":"mainnet","indexed_time":1612996343064065,"timestamp":"2020-12-22T19:19:49Z","protocol":"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"}`,
			bigMapType: `{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`,
			operation: &operation.Operation{
				Tags:       []string{consts.FA12Tag},
				Entrypoint: consts.TransferEntrypoint,
			},
			want: []models.Model{
				&tbModel.TokenBalance{
					Address:  "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					Contract: "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
					Network:  "mainnet",
					Value:    newBigInt("0"),
					TokenID:  0,
					IsLedger: true,
				},
			},
		}, {
			name:       "test 3",
			bmd:        `{"ptr":2071,"key":{"bytes":"00016631ce723071ea19a87bd93d7e2f81dd82c18565"},"key_hash":"exprvA4NaRxQEqyJad5LzVz7rohGSQDn9B32KR87KxRQ43kWiaqPS4","key_strings":["tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo"],"value":"","value_strings":[],"level":1455291,"contract":"KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E","network":"mainnet","timestamp":"2021-05-03T10:03:20Z","protocol":"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA"}`,
			bigMapType: `{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`,
			operation: &operation.Operation{
				Tags:        []string{consts.FA2Tag, consts.LedgerTag},
				Entrypoint:  "burn",
				Kind:        consts.Transaction,
				Network:     "mainnet",
				Destination: "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
				Hash:        "opNQeUBKfJzBjCNLuo5HkyZynhm5TMe1KEtwioqUrWM1ygmYVDX",
				Level:       1455291,
			},
			want: []models.Model{
				&tbModel.TokenBalance{
					Address:  "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
					Contract: "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
					Network:  "mainnet",
					Value:    newBigInt("0"),
					TokenID:  0,
					IsLedger: true,
				}, &transfer.Transfer{
					From:       "tz2HdbFWnzRZ7B9fM2xZCYdZv8rM5frGKDCo",
					Contract:   "KT1981tPmXh4KrUQKZpQKb55kREX7QGJcF3E",
					Network:    "mainnet",
					Value:      newBigInt("168000"),
					TokenID:    0,
					Entrypoint: "burn",
					Hash:       "opNQeUBKfJzBjCNLuo5HkyZynhm5TMe1KEtwioqUrWM1ygmYVDX",
					Level:      1455291,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledger := &Ledger{
				tokenBalances: tbRepo,
			}

			var bmd bigmapdiff.BigMapDiff
			if err := json.UnmarshalFromString(tt.bmd, &bmd); err != nil {
				t.Errorf("UnmarshalFromString error=%s", err)
				return
			}

			typ, err := ast.NewTypedAstFromString(tt.bigMapType)
			if err != nil {
				t.Errorf("NewTypedAstFromString error=%s", err)
				return
			}

			got, err := ledger.getResultModels(&bmd, typ.Nodes[0].(*ast.BigMap), tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ledger.getTokenBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
