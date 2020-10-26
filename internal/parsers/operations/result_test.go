package operations

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
)

func TestResult_Parse(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		fileName string
		want     models.OperationResult
	}{
		{
			name:     "test 1",
			root:     "",
			fileName: "./data/result/test1.json",
			want: models.OperationResult{
				Status:      "applied",
				ConsumedGas: 10207,
			},
		}, {
			name:     "test 2",
			root:     "operation_result",
			fileName: "./data/result/test2.json",
			want: models.OperationResult{
				Status:      "applied",
				ConsumedGas: 10207,
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

			if got := NewResult(tt.root).Parse(data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Result.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
