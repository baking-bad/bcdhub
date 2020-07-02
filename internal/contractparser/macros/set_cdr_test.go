package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_setCdrFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "invalid length",
			tree: `[{"prim":"CDR"},{"prim":"SWAP"},{"prim":"CDR"}]`,
		}, {
			name: "invalid prim 1",
			tree: `[{"prim":"CDR"},{"prim":"PAIR"}]`,
		}, {
			name: "invalid prim 2",
			tree: `[{"prim":"CAR"},{"prim":"CDR"}]`,
		}, {
			name: "set cdr",
			tree: `[{"prim":"CAR"},{"prim":"PAIR"}]`,
			want: setCdrMacros{},
		}, {
			name: "set cdr",
			tree: `[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DROP"},{"prim":"CAR"},{"prim":"PAIR"}]`,
			want: setCdrMacros{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setCdrFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("setCdrFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("setCdrFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setCdrFamily.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
