package metrics

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func TestBinMask_Compute(t *testing.T) {
	type args struct {
		a contract.Contract
		b contract.Contract
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test 1",
			args: args{
				a: contract.Contract{Tags: 192},
				b: contract.Contract{Tags: 192},
			},
			want: 1,
		}, {
			name: "test 2",
			args: args{
				a: contract.Contract{Tags: 3},
				b: contract.Contract{Tags: 0},
			},
			want: 0.133333,
		}, {
			name: "test 3",
			args: args{
				a: contract.Contract{Tags: 7},
				b: contract.Contract{Tags: 3},
			},
			want: 0.066666,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BinMask{
				Field: "Tags",
			}
			if got := m.Compute(tt.args.a, tt.args.b); got.Value != tt.want {
				t.Errorf("BinMask.Compute() = %v, want %v", got.Value, tt.want)
			}
		})
	}
}
