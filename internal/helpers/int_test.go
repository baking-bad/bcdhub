package helpers

import (
	"reflect"
	"testing"
)

func TestMerge2Arrays(t *testing.T) {
	type args struct {
		a []int64
		b []int64
	}
	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "a is empty",
			args: args{
				a: nil,
				b: []int64{1, 2},
			},
			want: []int64{1, 2},
		}, {
			name: "b is empty",
			args: args{
				b: nil,
				a: []int64{1, 2},
			},
			want: []int64{1, 2},
		}, {
			name: "both are empty",
			args: args{
				b: nil,
				a: nil,
			},
			want: []int64{},
		}, {
			name: "a larger than b",
			args: args{
				a: []int64{1, 3, 4},
				b: []int64{1, 2},
			},
			want: []int64{1, 2, 3, 4},
		}, {
			name: "b larger than a",
			args: args{
				b: []int64{1, 3, 4},
				a: []int64{1, 2},
			},
			want: []int64{1, 2, 3, 4},
		}, {
			name: "b larger than a (2)",
			args: args{
				b: []int64{1, 3, 4, 6, 9},
				a: []int64{1, 2, 5, 8},
			},
			want: []int64{1, 2, 3, 4, 5, 6, 8, 9},
		}, {
			name: "b larger than a (3)",
			args: args{
				b: []int64{1, 3, 4, 6, 9},
				a: []int64{5},
			},
			want: []int64{1, 3, 4, 5, 6, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Merge2ArraysInt64(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge2Arrays() = %v, want %v", got, tt.want)
			}
		})
	}
}
