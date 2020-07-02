package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_setCarFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "invalid length",
			tree: `[{}]`,
		}, {
			name: "invalid prim 3",
			tree: `[{"prim":"CDR"},{"prim":"SWAP"},{"prim":"CDR"}]`,
		}, {
			name: "invalid prim 2",
			tree: `[{"prim":"CDR"},{"prim":"CDR"},{"prim":"PAIR"}]`,
		}, {
			name: "invalid prim 1",
			tree: `[{"prim":"SWAP"},{"prim":"SWAP"},{"prim":"PAIR"}]`,
		}, {
			name: "set car",
			tree: `[{"prim":"CDR"},{"prim":"SWAP"},{"prim":"PAIR"}]`,
			want: setCarMacros{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setCarFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("setCarFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("setCarFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setCarFamily.Find() = %T, want %T", got, tt.want)
			}
		})
	}
}
