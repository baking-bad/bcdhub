package operations

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
)

func Test_parseMetadata(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     *Metadata
	}{
		{
			name:     "test 1",
			fileName: "./data/operation_metadata/test1.json",
			want: &Metadata{
				BalanceUpdates: []models.BalanceUpdate{
					{
						Kind:     "contract",
						Contract: "tz1TEZtYnuLiZLdA6c7JysAUJcHMrogu4Cpr",
						Change:   -6410,
					}, {
						Kind:     "contract",
						Contract: "tz1Y8zdtVe2wWe7QdNTnAdwBceqYBCdA3Jj8",
						Change:   6410,
					},
				},
				Result: models.OperationResult{
					Status:      "applied",
					ConsumedGas: 10207,
				},
			},
		}, {
			name:     "test 2",
			fileName: "./data/operation_metadata/test2.json",
			want: &Metadata{
				BalanceUpdates: []models.BalanceUpdate{},
				Result: models.OperationResult{
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
			if got := parseMetadata(data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
