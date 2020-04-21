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

func TestMaxInt(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Case 1",
			args: args{
				a: 10,
				b: 5,
			},
			want: 10,
		}, {
			name: "Case 2",
			args: args{
				a: 1,
				b: 5,
			},
			want: 5,
		}, {
			name: "Case 3",
			args: args{
				a: 5,
				b: 5,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxInt(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("MaxInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinInt(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Case 1",
			args: args{
				a: 10,
				b: 5,
			},
			want: 5,
		}, {
			name: "Case 2",
			args: args{
				a: 1,
				b: 5,
			},
			want: 1,
		}, {
			name: "Case 3",
			args: args{
				a: 5,
				b: 5,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinInt(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("MinInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
