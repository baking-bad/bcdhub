package handlers

import (
	"reflect"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/golang/mock/gomock"
)

func TestTezosDomain_Do(t *testing.T) {
	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ts := time.Now()

	type args struct {
		bmd     *domains.BigMapDiff
		storage string
	}
	tests := []struct {
		name    string
		args    args
		want1   []models.Model
		want    bool
		wantErr bool
	}{
		{
			name: "test 1: record",
			args: args{
				bmd: &domains.BigMapDiff{
					BigMapDiff: &bigmapdiff.BigMapDiff{
						ID:          10160561,
						Ptr:         1264,
						KeyHash:     "exprvG95A3YxzqRnRUNkeJH6sXL3at8TEPNzsRLpjuL3aoGk4MuWEk",
						Level:       1529158,
						Contract:    "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
						Network:     types.Mainnet,
						Timestamp:   ts,
						ProtocolID:  11,
						OperationID: 11136755,
						KeyStrings:  []string{"sodagenjo.tez"},
						ValueStrings: []string{
							"tz2Xd3x8mbRMpmLeRGMfuRAHNYgWDim53BQh",
							"sodagenjo.tez",
							"tz2Xd3x8mbRMpmLeRGMfuRAHNYgWDim53BQh",
						},
						Key:   []byte(`{"bytes":"736f646167656e6a6f2e74657a"}`),
						Value: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0001ffa97094ca718ef3c9b42416bd7cd8013b4176bf"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"736f646167656e6a6f2e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0001ffa97094ca718ef3c9b42416bd7cd8013b4176bf"}]},{"prim":"Some","args":[{"int":"10149"}]}]}]}`),
					},
					Operation: &operation.Operation{
						DeffatedStorage: []byte(`{"prim":"Pair","args":[[{"int":"1260"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1261"},{"int":"1262"}]},{"prim":"Pair","args":[{"int":"1263"},{"int":"10150"}]}]},{"prim":"Pair","args":[{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"},{"int":"1264"}]},{"int":"1265"},{"int":"1266"}],[{"bytes":"014796e76af90e6327adfab057bbbe0375cd2c8c1000"},{"bytes":"015c6799f783b8d118b704267f634c5d24d19e9a9f00"},{"bytes":"0168e9b7d86646e312c76dfbedcbcdb24320875a3600"},{"bytes":"019178a76f3c41a9541d2291cad37dd5fb96a6850500"},{"bytes":"01ac3638385caa4ad8126ea84e061f4f49baa44d3c00"},{"bytes":"01d2a0974172cf6fc8b1eefdebd5bea681616f7c6f00"}]]}`),
					},
				},
				storage: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes"},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}]}]}]}],"annots":["%actions"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}],"annots":["%store"]}]},{"prim":"set","args":[{"prim":"address"}],"annots":["%trusted_senders"]}]}`,
			},
			want: true,
			want1: []models.Model{
				&tezosdomain.TezosDomain{
					Name:      "sodagenjo.tez",
					Network:   types.Mainnet,
					Address:   "tz2Xd3x8mbRMpmLeRGMfuRAHNYgWDim53BQh",
					Level:     1529158,
					Timestamp: ts,
				},
			},
			wantErr: false,
		}, {
			name: "test 2: expiry map",
			args: args{
				bmd: &domains.BigMapDiff{
					BigMapDiff: &bigmapdiff.BigMapDiff{
						ID:           10160562,
						Ptr:          1262,
						KeyHash:      "exprvG95A3YxzqRnRUNkeJH6sXL3at8TEPNzsRLpjuL3aoGk4MuWEk",
						Level:        1529158,
						Contract:     "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
						Network:      types.Mainnet,
						Timestamp:    ts,
						ProtocolID:   11,
						OperationID:  11136755,
						KeyStrings:   []string{"sodagenjo.tez"},
						ValueStrings: []string{},
						Key:          []byte(`{"bytes":"736f646167656e6a6f2e74657a"}`),
						Value:        []byte(`{"int":"1656071422"}`),
					},
					Operation: &operation.Operation{
						DeffatedStorage: []byte(`{"prim":"Pair","args":[[{"int":"1260"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1261"},{"int":"1262"}]},{"prim":"Pair","args":[{"int":"1263"},{"int":"10150"}]}]},{"prim":"Pair","args":[{"bytes":"01ebb657570e494e8a7bd43ac3bf7cfd0267a32a9f00"},{"int":"1264"}]},{"int":"1265"},{"int":"1266"}],[{"bytes":"014796e76af90e6327adfab057bbbe0375cd2c8c1000"},{"bytes":"015c6799f783b8d118b704267f634c5d24d19e9a9f00"},{"bytes":"0168e9b7d86646e312c76dfbedcbcdb24320875a3600"},{"bytes":"019178a76f3c41a9541d2291cad37dd5fb96a6850500"},{"bytes":"01ac3638385caa4ad8126ea84e061f4f49baa44d3c00"},{"bytes":"01d2a0974172cf6fc8b1eefdebd5bea681616f7c6f00"}]]}`),
					},
				},
				storage: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes"},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}]}]}]}],"annots":["%actions"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}],"annots":["%data"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}],"annots":["%expiry_map"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"bytes"}],"annots":["%expiry_key"]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%tzip12_token_id"]}]}]}],"annots":["%records"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%internal_data"]},{"prim":"option","args":[{"prim":"bytes"}],"annots":["%name"]}]},{"prim":"address","annots":["%owner"]}]}],"annots":["%reverse_records"]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}],"annots":["%tzip12_tokens"]}]}]}],"annots":["%store"]}]},{"prim":"set","args":[{"prim":"address"}],"annots":["%trusted_senders"]}]}`,
			},
			want: true,
			want1: []models.Model{
				&tezosdomain.TezosDomain{
					Name:       "sodagenjo.tez",
					Expiration: time.Unix(1656071422, 0).UTC(),
					Network:    types.Mainnet,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := &TezosDomain{
				storage: generalRepo,
				contracts: map[contract.Address]struct{}{
					{
						Address: "KT1GBZmSxmnKJXGMdMLbugPfLyUPmuLSMwKS",
						Network: types.Mainnet,
					}: {},
				},
				ptrs: make(map[contract.Address]ptrs),
			}

			var storageType ast.TypedAst
			if err := json.UnmarshalFromString(tt.args.storage, &storageType); err != nil {
				t.Errorf("UnmarshalFromString() error = %v", err)
				return
			}

			got, got1, err := td.Do(tt.args.bmd, &storageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("TezosDomain.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TezosDomain.Do() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("TezosDomain.Do() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
