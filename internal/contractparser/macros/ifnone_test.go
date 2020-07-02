package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_ifNoneFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "Invalid type",
			tree: `[[]]`,
			want: nil,
		}, {
			name: "Invalid prim",
			tree: `[{"prim":"string"}]`,
			want: nil,
		}, {
			name: "Invalid args",
			tree: `[{"prim":"IF_NONE", "args":[]}]`,
			want: nil,
		}, {
			name: "Invalid first arg",
			tree: `[{"prim":"IF_NONE", "args":[{"prim":"PAIR"}, {"prim":"FAIL"}]}]`,
			want: nil,
		}, {
			name: "Invalid second arg",
			tree: `[{"prim":"IF_NONE", "args":[{"prim":"FAIL"}, {"prim":"FAIL"}]}]`,
			want: nil,
		}, {
			name: "Assert Some",
			tree: `[{"prim":"IF_NONE", "args":[{"prim":"FAIL"}, [{"prim":"RENAME"}]]}]`,
			want: assertSome{},
		}, {
			name: "Assert None",
			tree: `[{"prim":"IF_NONE", "args":[[], {"prim":"FAIL"}]}]`,
			want: assertNone{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := ifNoneFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("ifNoneFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ifNoneFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ifNoneFamily.Find() = %T, want %T", got, tt.want)
			}
		})
	}
}
