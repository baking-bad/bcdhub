package noderpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLazyStorageDiff_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		lsd     LazyStorageDiff
		data    []byte
		wantErr bool
	}{
		{
			name: "test 1",
			lsd: LazyStorageDiff{
				LazyStorageDiffKind: LazyStorageDiffKind{
					Kind: "sapling_state",
					ID:   154,
				},
				Diff: &Diff{
					SaplingState: &LazySaplingStateDiff{
						Action: "update",
						Updates: LazySaplingStateUpdate{
							Nullifiers: []string{},
							CommitmentsAndCiphertexts: []CommitmentsAndCiphertexts{
								{
									Commitment: "0a99a1e4a81f0e67c390fd56d792ace87e19514e0d800a33eab7acdaa42e6159",
									CipherText: CipherText{
										CV:         "2b659bc87151b23cbdfe92e3dd7ced4ea5af541e7b56a44b184b7042f6600985",
										EPK:        "783c9ce539b44a23da6aafd9fa0b2b2588056f0df5f60c9b7b166ccb16f4f524",
										PayloadEnc: "3574edddc500e322cde8a5033de933b4df6eb6c7d3a2075da6521e859debaacecd6e529311eb0dd98374e4b014adfaaa631b9a68e7a1b7a7ba56d69a624485bc02a624ead89e30b7a1f8a3bb442df3",
										NonceEnc:   "d13c4073ed76f607c51aab47afeae66c1c047ff21ea3df85",
										PayloadOut: "a24bb23cac3dcddd208d3f884fda05cb1388389dfa7546ae67ea7ce58de6fbe0b8e4f1653dcca65b95710872f988dfd7d4051861e20473daa0f57c824d8e8ba51cd10dac5920d01963ba1470c79295e9",
										NonceOut:   "43b24c9282d0234a0e802d71e5701625c04403135bb6c7de",
									},
								},
							},
						},
					},
				},
			},
			data: []byte(`{
				"kind": "sapling_state",
				"id": "154",
				"diff": {
					"action": "update",
					"updates": {
						"commitments_and_ciphertexts": [
							[
								"0a99a1e4a81f0e67c390fd56d792ace87e19514e0d800a33eab7acdaa42e6159",
								{
									"cv": "2b659bc87151b23cbdfe92e3dd7ced4ea5af541e7b56a44b184b7042f6600985",
									"epk": "783c9ce539b44a23da6aafd9fa0b2b2588056f0df5f60c9b7b166ccb16f4f524",
									"payload_enc": "3574edddc500e322cde8a5033de933b4df6eb6c7d3a2075da6521e859debaacecd6e529311eb0dd98374e4b014adfaaa631b9a68e7a1b7a7ba56d69a624485bc02a624ead89e30b7a1f8a3bb442df3",
									"nonce_enc": "d13c4073ed76f607c51aab47afeae66c1c047ff21ea3df85",
									"payload_out": "a24bb23cac3dcddd208d3f884fda05cb1388389dfa7546ae67ea7ce58de6fbe0b8e4f1653dcca65b95710872f988dfd7d4051861e20473daa0f57c824d8e8ba51cd10dac5920d01963ba1470c79295e9",
									"nonce_out": "43b24c9282d0234a0e802d71e5701625c04403135bb6c7de"
								}
							]
						],
						"nullifiers": []
					}
				}
			}`),
		}, {
			name: "test 2",
			lsd: LazyStorageDiff{
				LazyStorageDiffKind: LazyStorageDiffKind{
					Kind: "big_map",
					ID:   154,
				},
				Diff: &Diff{
					BigMap: &LazyBigMapDiff{
						Action: "remove",
					},
				},
			},
			data: []byte(`{
				"kind": "big_map",
				"id": "154",
				"diff": {
					"action": "remove"
				}
			}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got LazyStorageDiff
			if err := got.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("LazyStorageDiff.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.lsd.Diff, got.Diff)
			assert.Equal(t, tt.lsd.LazyStorageDiffKind.ID, got.LazyStorageDiffKind.ID)
			assert.Equal(t, tt.lsd.LazyStorageDiffKind.Kind, got.LazyStorageDiffKind.Kind)
		})
	}
}
