package contractparser

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_parseFail(t *testing.T) {
	tests := []struct {
		name string
		args string
		want *fail
	}{
		{
			name: "empty",
			args: ``,
			want: nil,
		}, {
			name: "big",
			args: `[ { "prim": "CAR" }, { "prim": "CAR" }, { "prim": "PUSH", "args": [ { "prim": "string" }, { "string": "UparamNoSuchEntryPoint" } ] }, { "prim": "PAIR" }, { "prim": "FAILWITH" } ]`,
			want: &fail{
				With: "UparamNoSuchEntryPoint",
			},
		}, {
			name: "nil big",
			args: `[ { "prim": "CAR" }, { "prim": "CAR" }, { "prim": "PAIR" }, { "prim": "FAILWITH" } ]`,
			want: nil,
		}, {
			name: "nil big 2",
			args: `[ { "prim": "CAR" }, { "prim": "CAR" }, { "prim": "PUSH", "args": [ { "prim": "string" }, { "string": "UparamNoSuchEntryPoint" } ] }, { "prim": "PAIR" } ]`,
			want: nil,
		}, {
			name: "small",
			args: `[ { "prim": "PUSH", "args": [ { "prim": "string" }, { "string": "UparamNoSuchEntryPoint" } ] }, { "prim": "FAILWITH" } ]`,
			want: &fail{
				With: "UparamNoSuchEntryPoint",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.args)
			if got := parseFail(data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFail() = %v, want %v", got, tt.want)
			}
		})
	}
}
