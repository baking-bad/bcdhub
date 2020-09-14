package database

import (
	"testing"

	"github.com/tidwall/gjson"
)

func Test_compareTypes(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "without token ID",
			args: args{
				a: `{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "nat"} ] }`,
				b: `{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "nat"} ] }`,
			},
			want: true,
		}, {
			name: "with token ID",
			args: args{
				a: `{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "nat" } ] }`,
				b: `{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "nat" } ] }`,
			},
			want: true,
		}, {
			name: "false",
			args: args{
				a: `{ "prim": "map", "args": [ { "prim": "address"}, {"prim": "nat"} ] }`,
				b: `{ "prim": "map", "args": [ { "prim": "pair", "args": [{ "prim": "address"}, {"prim": "nat"}] }, { "prim" : "nat" } ] }`,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := gjson.Parse(tt.args.a)
			b := gjson.Parse(tt.args.b)
			if got := compareTypes(a, b); got != tt.want {
				t.Errorf("compareTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}
