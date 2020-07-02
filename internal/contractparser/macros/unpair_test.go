package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_unpairFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "Success case",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
			want: unpairMacros{},
		}, {
			name: "prim 0 invalid",
			tree: `[{"prim":"DIP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
		}, {
			name: "prim 1 invalid",
			tree: `[{"prim":"DUP"},{"prim":"CDR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
		}, {
			name: "prim 2 invalid",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DUP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
		}, {
			name: "args 2 invalid",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[]}]`,
		}, {
			name: "dip args: invalid type",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[{"prim":"CDR","annots":["%b"]}]}]`,
		}, {
			name: "dip args: invalid prim",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CAR","annots":["%b"]}]]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := unpairFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("unpairFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpairFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unpairFamily.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
