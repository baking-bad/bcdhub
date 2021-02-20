package storage

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/stretchr/testify/assert"

	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	"github.com/golang/mock/gomock"
	"github.com/tidwall/gjson"
)

func newTestBabylon(ctrl *gomock.Controller) (*Babylon, *mock_bmd.MockRepository) {
	repo := mock_bmd.NewMockRepository(ctrl)
	return &Babylon{
		repo: repo,

		ptrMap:            make(map[int64]int64),
		temporaryPointers: make(map[int64]*ast.BigMap),
	}, repo
}

func TestBabylon_ParseTransaction(t *testing.T) {
	type args struct {
		content   string
		operation operation.Operation
	}
	tests := []struct {
		name    string
		args    args
		want    RichStorage
		wantErr bool
	}{
		{
			name: "delphinet/KT1HHsW85jrLrHdAy9DwScqiM1RERkTT9Q6e",
			args: args{
				content: `{"kind":"transaction","source":"tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb","fee":"10135","counter":"704669","gas_limit":"97777","storage_limit":"31656","amount":"1000000","destination":"KT1HHsW85jrLrHdAy9DwScqiM1RERkTT9Q6e","parameters":{"entrypoint":"launchExchange","value":{"prim":"Pair","args":[{"string":"KT1KVJ4S53zE6E8oo8L8TyMgAh1ACpf9HweA"},{"int":"1000000"}]}},"metadata":{"balance_updates":[{"kind":"contract","contract":"tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb","change":"-10135"},{"kind":"freezer","category":"fees","delegate":"tz1PirboHQVqkYqLSWfHUHEy3AdhYUNJpvGy","cycle":91,"change":"10135"}],"operation_result":{"status":"applied","storage":{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"10416"},{"int":"10417"}]},{"prim":"Pair","args":[[{"bytes":"0177a057dafeab5f829e044dc0e52047e01283d6d500"}],{"int":"10418"}]}]},"big_map_diff":[{"action":"copy","source_big_map":"10417","destination_big_map":"-8"},{"action":"alloc","big_map":"-7","key_type":{"prim":"key_hash"},"value_type":{"prim":"nat"}},{"action":"alloc","big_map":"-6","key_type":{"prim":"address"},"value_type":{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}]},{"prim":"timestamp"}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}]}},{"action":"alloc","big_map":"-5","key_type":{"prim":"key_hash"},"value_type":{"prim":"timestamp"}},{"action":"alloc","big_map":"-4","key_type":{"prim":"address"},"value_type":{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}]},{"prim":"timestamp"}]}},{"action":"alloc","big_map":"-3","key_type":{"prim":"address"},"value_type":{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"nat"}]}},{"action":"update","big_map":"-3","key_hash":"exprtr3iA2ZhFDtnJZDS1nVxJYeXGWw2AWziVAD7DZf7kxsHmNLZBB","key":{"bytes":"00006b82198cb179e8306c1bedd08f12dc863f328886"},"value":{"prim":"Pair","args":[{"prim":"Pair","args":[[],{"int":"1000"}]},{"int":"0"}]}},{"action":"alloc","big_map":"-2","key_type":{"prim":"string"},"value_type":{"prim":"bytes"}},{"action":"update","big_map":"-2","key_hash":"expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo","key":{"string":""},"value":{"bytes":"4b54313964473234535066486464564d4d50456131675342706f4c5a6374786e556b48542f6d65746164617461"}},{"action":"copy","source_big_map":"10416","destination_big_map":"-1"},{"action":"update","big_map":"10418","key_hash":"exprvDFsAkF12eo7cP1EtDk52Ef72CzDhxuJmwXCqbqSWq6CrJ3ziX","key":{"bytes":"0177a057dafeab5f829e044dc0e52047e01283d6d500"},"value":{"bytes":"0117f1f0e206ba4c32f1f43de336b0ef2785f4014500"}}],"balance_updates":[{"kind":"contract","contract":"tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb","change":"-29750"},{"kind":"contract","contract":"tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb","change":"-1000000"},{"kind":"contract","contract":"KT1HHsW85jrLrHdAy9DwScqiM1RERkTT9Q6e","change":"1000000"}],"consumed_gas":"56838","consumed_milligas":"56837252","storage_size":"39563","paid_storage_size_diff":"119"}}}`,
				operation: operation.Operation{
					Level:    186900,
					Network:  "delphinet",
					Protocol: "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Script:   gjson.Parse(`{"code":[{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%min_tez"]},{"prim":"nat","annots":["%min_tokens"]}]},{"prim":"nat","annots":["%shares"]}],"annots":["%divestLiquidity"]},{"prim":"nat","annots":["%initializeExchange"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%investLiquidity"]},{"prim":"pair","args":[{"prim":"nat","annots":["%amount"]},{"prim":"address","annots":["%receiver"]}],"annots":["%tezToTokenPayment"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%amount"]},{"prim":"nat","annots":["%min_out"]}]},{"prim":"address","annots":["%receiver"]}],"annots":["%tokenToTezPayment"]},{"prim":"pair","args":[{"prim":"nat","annots":["%value"]},{"prim":"address","annots":["%voter"]}],"annots":["%veto"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"key_hash","annots":["%candidate"]},{"prim":"nat","annots":["%value"]}]},{"prim":"address","annots":["%voter"]}],"annots":["%vote"]},{"prim":"address","annots":["%withdrawProfit"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_candidate"]},{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_delegated"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_veto"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%allowances"]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%last_loyalty_per_share"]},{"prim":"timestamp","annots":["%last_period_finish"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_update_time"]},{"prim":"nat","annots":["%loyalty_per_share"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_per_token"]},{"prim":"nat","annots":["%total_accomulated_loyalty"]}]}]}],"annots":["%reward_info"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%tez_pool"]},{"prim":"address","annots":["%token_address"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%total_votes"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%loyalty"]},{"prim":"nat","annots":["%loyalty_paid"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"timestamp","annots":["%update_time"]}]}],"annots":["%user_rewards"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"timestamp"}],"annots":["%vetos"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%candidate"]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}],"annots":["%voters"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"nat"}],"annots":["%votes"]}]}]}]}]}]},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_candidate"]},{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_delegated"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_veto"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%allowances"]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%last_loyalty_per_share"]},{"prim":"timestamp","annots":["%last_period_finish"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_update_time"]},{"prim":"nat","annots":["%loyalty_per_share"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_per_token"]},{"prim":"nat","annots":["%total_accomulated_loyalty"]}]}]}],"annots":["%reward_info"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%tez_pool"]},{"prim":"address","annots":["%token_address"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%total_votes"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%loyalty"]},{"prim":"nat","annots":["%loyalty_paid"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"timestamp","annots":["%update_time"]}]}],"annots":["%user_rewards"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"timestamp"}],"annots":["%vetos"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%candidate"]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}],"annots":["%voters"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"nat"}],"annots":["%votes"]}]}]}]}]}]}]}],"annots":["%dex_lambdas"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%spender"]},{"prim":"nat","annots":["%value"]}],"annots":["%iApprove"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%spender"]}]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%iGetAllowance"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%iGetBalance"]},{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%iGetTotalSupply"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%from"]},{"prim":"pair","args":[{"prim":"address","annots":["%to"]},{"prim":"nat","annots":["%value"]}]}],"annots":["%iTransfer"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_candidate"]},{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_delegated"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_veto"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%allowances"]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%last_loyalty_per_share"]},{"prim":"timestamp","annots":["%last_period_finish"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_update_time"]},{"prim":"nat","annots":["%loyalty_per_share"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_per_token"]},{"prim":"nat","annots":["%total_accomulated_loyalty"]}]}]}],"annots":["%reward_info"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%tez_pool"]},{"prim":"address","annots":["%token_address"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%total_votes"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%loyalty"]},{"prim":"nat","annots":["%loyalty_paid"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"timestamp","annots":["%update_time"]}]}],"annots":["%user_rewards"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"timestamp"}],"annots":["%vetos"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%candidate"]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}],"annots":["%voters"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"nat"}],"annots":["%votes"]}]}]}]}]}]},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_candidate"]},{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%current_delegated"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_veto"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%allowances"]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%last_loyalty_per_share"]},{"prim":"timestamp","annots":["%last_period_finish"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_update_time"]},{"prim":"nat","annots":["%loyalty_per_share"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward_per_token"]},{"prim":"nat","annots":["%total_accomulated_loyalty"]}]}]}],"annots":["%reward_info"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%tez_pool"]},{"prim":"address","annots":["%token_address"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%total_votes"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%loyalty"]},{"prim":"nat","annots":["%loyalty_paid"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"timestamp","annots":["%update_time"]}]}],"annots":["%user_rewards"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"timestamp"}],"annots":["%vetos"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}],"annots":["%candidate"]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}],"annots":["%voters"]},{"prim":"big_map","args":[{"prim":"key_hash"},{"prim":"nat"}],"annots":["%votes"]}]}]}]}]}]}]}],"annots":["%token_lambdas"]}]},{"prim":"pair","args":[{"prim":"set","args":[{"prim":"address"}],"annots":["%token_list"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"address"}],"annots":["%token_to_exchange"]}]}]}]}]}`),
				},
			},
			want: RichStorage{
				DeffatedStorage: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"10416"},{"int":"10417"}]},{"prim":"Pair","args":[[{"bytes":"0177a057dafeab5f829e044dc0e52047e01283d6d500"}],{"int":"10418"}]}]}`,
				Models: []models.Model{
					&bigmapdiff.BigMapDiff{
						Ptr:      10418,
						Key:      []byte(`{"bytes":"0177a057dafeab5f829e044dc0e52047e01283d6d500"}`),
						KeyHash:  "exprvDFsAkF12eo7cP1EtDk52Ef72CzDhxuJmwXCqbqSWq6CrJ3ziX",
						Value:    []byte(`{"bytes":"0117f1f0e206ba4c32f1f43de336b0ef2785f4014500"}`),
						Level:    186900,
						Address:  "KT1HHsW85jrLrHdAy9DwScqiM1RERkTT9Q6e",
						Network:  "delphinet",
						Protocol: "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			b, repo := newTestBabylon(ctrl)

			repo.
				EXPECT().
				GetByPtr(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]bigmapdiff.BigMapDiff{}, nil).
				AnyTimes()

			content := gjson.Parse(tt.args.content)

			got, err := b.ParseTransaction(content, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("Babylon.ParseTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.DeffatedStorage, got.DeffatedStorage)
			assert.Equal(t, false, got.Empty)
			assert.Len(t, got.Models, len(tt.want.Models))

			for i := range tt.want.Models {
				bmd := got.Models[i].(*bigmapdiff.BigMapDiff)
				newBmd := tt.want.Models[i].(*bigmapdiff.BigMapDiff)
				newBmd.ID = bmd.ID
				newBmd.IndexedTime = bmd.IndexedTime
			}
			assert.Equal(t, tt.want.Models, got.Models)
		})
	}
}
