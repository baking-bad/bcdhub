package operations

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
)

func TestBalanceUpdate_Parse(t *testing.T) {
	operation := models.Operation{
		Network:      "test",
		Level:        100,
		Hash:         "hash",
		ContentIndex: 1,
		Nonce:        nil,
	}
	tests := []struct {
		name     string
		root     string
		fileName string
		want     []*models.BalanceUpdate
	}{
		{
			name:     "test 1",
			root:     "",
			fileName: "./data/balance_update/test1.json",
			want: []*models.BalanceUpdate{
				{
					Contract:      "KT1A946hDgLGfFudWU7hzfnTdZK8TZyLRHeT",
					Change:        -2655,
					Network:       "test",
					Level:         100,
					OperationHash: "hash",
					ContentIndex:  1,
					Nonce:         nil,
				},
			},
		},
		{
			name:     "test 2",
			root:     "operation_result",
			fileName: "./data/balance_update/test2.json",
			want: []*models.BalanceUpdate{
				{
					Contract:      "KT1A946hDgLGfFudWU7hzfnTdZK8TZyLRHeT",
					Change:        -29075891,
					Network:       "test",
					Level:         100,
					OperationHash: "hash",
					ContentIndex:  1,
					Nonce:         nil,
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
			got := NewBalanceUpdate(tt.root, operation).Parse(data)
			if len(got) != len(tt.want) {
				t.Errorf("BalanceUpdate.Parse() = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if !compareBalanceUpdates(got[i], tt.want[i]) {
					t.Errorf("BalanceUpdate.Parse() = %v, want %v", got[i], tt.want[i])
					return
				}
			}
		})
	}
}
