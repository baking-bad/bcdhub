package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_cadrFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "success",
			tree: `[{"prim":"CAR"},{"prim":"CAR"},{"prim":"CAR"},{"prim":"CDR"},{"prim":"CDR"},{"prim":"CDR"},{"prim":"CAR"},{"prim":"CDR","annots":["%ahaha"]}]`,
			want: cadrMacros{
				name:   "CAAADDDADR",
				length: 8,
			},
		}, {
			name: "success 2",
			tree: `[{"prim":"CAR"},{"prim":"CAR"},{"prim":"CAR"},{"prim":"CDR"},{"prim":"CDR"},{"prim":"CAR"},{"prim":"CDR","annots":["%ahaha"]}]`,
			want: cadrMacros{
				name:   "CAAADDADR",
				length: 7,
			},
		}, {
			name: "invalid prim",
			tree: `[{"prim":"CAR"},{"prim":"CAR"},{"prim":"CAR"},{"prim":"CDR"},{"prim":"CDR"},{"prim":"CAR"},{"prim":"DIP","annots":["%ahaha"]}]`,
		}, {
			name: "alone CAR",
			tree: `[{"prim":"CAR"}]`,
		}, {
			name: "invalid type",
			tree: `[{"prim":"CAR"},[]]`,
		}, {
			name: "empty array",
			tree: `[]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := cadrFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("fail.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("cadrFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cadrFamily.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
