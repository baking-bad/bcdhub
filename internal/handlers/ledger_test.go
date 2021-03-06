package handlers

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
)

func TestLedger_getResultModels(t *testing.T) {
	tests := []struct {
		name       string
		bmd        string
		bigMapType string
		want       []models.Model
		wantErr    bool
	}{
		{
			name:       "test 1",
			bmd:        `{"ptr":257,"key":{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"},"key_hash":"expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf","key_strings":["tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"],"value":{"int":"1000000"},"value_strings":[],"operation_id":"4784c35cc6444b8ca0eb9b7b4698e6cb","level":1269694,"address":"KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH","network":"mainnet","indexed_time":1612996343064065,"timestamp":"2020-12-22T19:19:49Z","protocol":"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"}`,
			bigMapType: `{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`,
			want: []models.Model{
				&tbModel.TokenBalance{
					Address:  "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					Contract: "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
					Network:  "mainnet",
					Value:    big.NewInt(1000000),
					TokenID:  0,
				},
			},
		}, {
			name:       "test 2",
			bmd:        `{"ptr":257,"key":{"bytes":"0000c67788ea8ada32b2426e1b02b9ebebdc2dc51007"},"key_hash":"expruCQuxuWpbLgZ5a4AhQ9nmdLVssrFZXmzTe8jFB5LMKvX6XPXVf","key_strings":["tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1"],"value":"","value_strings":[],"operation_id":"4784c35cc6444b8ca0eb9b7b4698e6cb","level":1269694,"address":"KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH","network":"mainnet","indexed_time":1612996343064065,"timestamp":"2020-12-22T19:19:49Z","protocol":"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo"}`,
			bigMapType: `{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}`,
			want: []models.Model{
				&tbModel.TokenBalance{
					Address:  "tz1djRgXXWWJiY1rpMECCxr5d9ZBqWewuiU1",
					Contract: "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH",
					Network:  "mainnet",
					Value:    big.NewInt(0),
					TokenID:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledger := &Ledger{}
			typ := []byte(tt.bigMapType)

			var bmd bigmapdiff.BigMapDiff
			if err := json.UnmarshalFromString(tt.bmd, &bmd); err != nil {
				t.Errorf("UnmarshalFromString error=%s", err)
				return
			}
			got, err := ledger.getResultModels(&bmd, typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ledger.getTokenBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
