package operations

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/operation"
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
				Result: operation.Result{
					Status:      "applied",
					ConsumedGas: 10207,
				},
			},
		}, {
			name:     "test 2",
			fileName: "./data/operation_metadata/test2.json",
			want: &Metadata{
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
			got := parseMetadata(data)
			if !reflect.DeepEqual(got.Result, tt.want.Result) {
				t.Errorf("parseMetadata() Result = %v, want %v", got.Result, tt.want.Result)
				return
			}
		})
	}
}
