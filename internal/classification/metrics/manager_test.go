package metrics

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func TestManager_Compute(t *testing.T) {
	type args struct {
		a contract.Contract
		b contract.Contract
	}
	tests := []struct {
		name string
		args args
		want Feature
	}{
		{
			name: "Case 1",
			args: args{
				a: contract.Contract{
					Manager: "test",
					Network: 0,
				},
				b: contract.Contract{
					Manager: "test",
					Network: 0,
				},
			},
			want: Feature{
				Name:  "manager",
				Value: 1.0,
			},
		}, {
			name: "Case 2",
			args: args{
				a: contract.Contract{
					Manager: "other",
					Network: 1,
				},
				b: contract.Contract{
					Manager: "test",
					Network: 1,
				},
			},
			want: Feature{
				Name:  "manager",
				Value: 0.0,
			},
		}, {
			name: "Case 3",
			args: args{
				a: contract.Contract{
					Manager: "test",
					Network: 1,
				},
				b: contract.Contract{
					Manager: "test",
					Network: 2,
				},
			},
			want: Feature{
				Name:  "manager",
				Value: 0.0,
			},
		}, {
			name: "Case 4",
			args: args{
				a: contract.Contract{
					Manager: "other",
					Network: 1,
				},
				b: contract.Contract{
					Manager: "test",
					Network: 2,
				},
			},
			want: Feature{
				Name:  "manager",
				Value: 0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			if got := m.Compute(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Manager.Compute() = %v, want %v", got, tt.want)
			}
		})
	}
}
