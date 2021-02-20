package macros

import (
	"testing"

	"github.com/valyala/fastjson"
)

func Test_fail_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "success",
			tree: `[[{"prim": "UNIT"},{"prim": "FAILWITH"}]]`,
			want: failMacros{},
		}, {
			name: "without UNIT",
			tree: `[[{"prim": "FAILWITH"}]]`,
			want: nil,
		}, {
			name: "without FAILWITH",
			tree: `[[{"prim": "UNIT"}]]`,
			want: nil,
		}, {
			name: "wrong primitive",
			tree: `[[{"prim": "PAIR"},{"prim": "FAILWITH"}]]`,
			want: nil,
		}, {
			name: "wrong primitive2",
			tree: `[[{"prim": "UNIT"},{"prim": "PAIR"}]]`,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := failFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("fail.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("fail.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fail.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
