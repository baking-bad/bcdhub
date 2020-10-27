package utils

import (
	"testing"

	"github.com/tidwall/gjson"
)

//nolint
func TestInt64Pointer(t *testing.T) {
	type args struct {
		hit gjson.Result
		tag string
	}
	tests := []struct {
		name string
		args args
		want *int64
	}{
		{
			name: "a = 1",
			args: args{
				hit: gjson.Parse(`{"a" : 1}`),
				tag: "a",
			},
			want: getInt64Ptr(1),
		}, {
			name: "a = nil",
			args: args{
				hit: gjson.Parse(`{"b": 1}`),
				tag: "a",
			},
			want: nil,
		}, {
			name: "a = 0",
			args: args{
				hit: gjson.Parse(`{"a": 0}`),
				tag: "a",
			},
			want: getInt64Ptr(0),
		}, {
			name: "a = null",
			args: args{
				hit: gjson.Parse(`{"a": null}`),
				tag: "a",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int64Pointer(tt.args.hit, tt.args.tag)
			if (tt.want == nil && got != nil) || (got == nil && tt.want != nil) {
				t.Errorf("Int64Pointer() = %v, want %v", got, tt.want)
				return
			} else if got == nil && tt.want == nil {
				return
			} else if *got != *tt.want {
				t.Errorf("Int64Pointer() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func getInt64Ptr(i int64) *int64 {
	return &i
}
