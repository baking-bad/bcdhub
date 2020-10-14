package operations

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
)

func TestBalanceUpdate_Parse(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		fileName string
		want     []models.BalanceUpdate
	}{
		{
			name:     "test 1",
			root:     "",
			fileName: "./data/balance_update/test1.json",
			want: []models.BalanceUpdate{
				{
					Kind:     "contract",
					Contract: "tz1Yshp26pvsFcjwpDQvoLdHsWEJMrvwuCpo",
					Change:   -2655,
				},
			},
		},
		{
			name:     "test 2",
			root:     "operation_result",
			fileName: "./data/balance_update/test2.json",
			want: []models.BalanceUpdate{
				{
					Kind:     "contract",
					Contract: "tz1XWGzpVmLV4oyGc7aM3GhjXEZQh4Qutgq3",
					Change:   -29075891,
				}, {
					Kind:     "contract",
					Contract: "tz1KvnPAoC4XTeLqtvYMGMt8BYdbkAjkbtvf",
					Change:   29075891,
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
			if got := NewBalanceUpdate(tt.root).Parse(data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BalanceUpdate.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
