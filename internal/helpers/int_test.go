package helpers

import (
	"testing"
)

func TestMax(t *testing.T) {
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
			if got := Max(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("MaxInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
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
			if got := Min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("MinInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
