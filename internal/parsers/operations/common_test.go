package operations

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

func Test_parseMetadata(t *testing.T) {
	op := operation.Operation{
		Network:      "test",
		Level:        100,
		Hash:         "hash",
		ContentIndex: 1,
		Nonce:        nil,
	}

	tests := []struct {
		name     string
		fileName string
		want     *Metadata
	}{
		{
			name:     "test 1",
			fileName: "./data/operation_metadata/test1.json",
			want: &Metadata{
				BalanceUpdates: []*balanceupdate.BalanceUpdate{
					{
						Contract:      "KT1PDAELuX7CypUHinUgFgGFskKs7ytwh5Vw",
						Change:        6410,
						Network:       "test",
						Level:         100,
						OperationHash: "hash",
						ContentIndex:  1,
						Nonce:         nil,
					},
				},
				Result: operation.Result{
					Status:      "applied",
					ConsumedGas: 10207,
				},
			},
		}, {
			name:     "test 2",
			fileName: "./data/operation_metadata/test2.json",
			want: &Metadata{
				BalanceUpdates: []*balanceupdate.BalanceUpdate{},
				Result: operation.Result{
					Status:      "backtracked",
					ConsumedGas: 96591,
					StorageSize: 196,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := readJSONFile(tt.fileName)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.fileName, err)
				return
			}
			got := parseMetadata(data, op)
			if !reflect.DeepEqual(got.Result, tt.want.Result) {
				t.Errorf("parseMetadata() Result = %v, want %v", got.Result, tt.want.Result)
				return
			}

			if len(got.BalanceUpdates) != len(tt.want.BalanceUpdates) {
				t.Errorf("parseMetadata() BalanceUpdates = %v, want %v", got.BalanceUpdates, tt.want.BalanceUpdates)
				return
			}

			for i := range got.BalanceUpdates {
				if !compareBalanceUpdates(got.BalanceUpdates[i], tt.want.BalanceUpdates[i]) {
					t.Errorf("parseMetadata() BalanceUpdates = %v, want %v", got.BalanceUpdates[i], tt.want.BalanceUpdates[i])
					return
				}
			}
		})
	}
}
