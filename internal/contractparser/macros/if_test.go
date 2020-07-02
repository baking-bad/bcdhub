package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_ifFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "ASSERT",
			tree: `[{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertMacros{},
		}, {
			name: "ASSERT_EQ: EQ",
			tree: `[{"prim": "EQ"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_EQ: NEQ",
			tree: `[{"prim": "NEQ"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_EQ: LT",
			tree: `[{"prim": "LT"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_EQ: LE",
			tree: `[{"prim": "LE"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_EQ: GT",
			tree: `[{"prim": "GT"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_EQ: GE",
			tree: `[{"prim": "GE"},{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: EQ",
			tree: `[[{"prim": "COMPARE"},{"prim": "EQ"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: NEQ",
			tree: `[[{"prim": "COMPARE"},{"prim": "NEQ"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: LT",
			tree: `[[{"prim": "COMPARE"},{"prim": "LT"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: LE",
			tree: `[[{"prim": "COMPARE"},{"prim": "LE"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: GT",
			tree: `[[{"prim": "COMPARE"},{"prim": "GT"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "ASSERT_CMPEQ: GE",
			tree: `[[{"prim": "COMPARE"},{"prim": "GE"}],{"prim": "IF", "args":[[],{"prim": "FAIL"}]}]`,
			want: assertCmpEqMacros{},
		}, {
			name: "CMPEQ: EQ",
			tree: `[{"prim": "COMPARE"},{"prim": "EQ"}]`,
			want: cmpEqMacros{},
		}, {
			name: "CMPEQ: NEQ",
			tree: `[{"prim": "COMPARE"},{"prim": "NEQ"}]`,
			want: cmpEqMacros{},
		}, {
			name: "CMPEQ: LT",
			tree: `[{"prim": "COMPARE"},{"prim": "LT"}]`,
			want: cmpEqMacros{},
		}, {
			name: "CMPEQ: LE",
			tree: `[{"prim": "COMPARE"},{"prim": "LE"}]`,
			want: cmpEqMacros{},
		}, {
			name: "CMPEQ: GT",
			tree: `[{"prim": "COMPARE"},{"prim": "GT"}]`,
			want: cmpEqMacros{},
		}, {
			name: "CMPEQ: GE",
			tree: `[{"prim": "COMPARE"},{"prim": "GE"}]`,
			want: cmpEqMacros{},
		}, {
			name: "IFEQ: EQ",
			tree: `[{"prim": "EQ"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFEQ: NEQ",
			tree: `[{"prim": "NEQ"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFEQ: LT",
			tree: `[{"prim": "LT"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFEQ: LE",
			tree: `[{"prim": "LE"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFEQ: GT",
			tree: `[{"prim": "GT"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFEQ: GE",
			tree: `[{"prim": "GE"},{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifEqMacros{},
		}, {
			name: "IFCMPEQ: EQ",
			tree: `[[{"prim": "COMPARE"},{"prim": "EQ"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		}, {
			name: "IFCMPEQ: NEQ",
			tree: `[[{"prim": "COMPARE"},{"prim": "NEQ"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		}, {
			name: "IFCMPEQ: LT",
			tree: `[[{"prim": "COMPARE"},{"prim": "LT"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		}, {
			name: "IFCMPEQ: LE",
			tree: `[[{"prim": "COMPARE"},{"prim": "LE"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		}, {
			name: "IFCMPEQ: GT",
			tree: `[[{"prim": "COMPARE"},{"prim": "GT"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		}, {
			name: "IFCMPEQ: GE",
			tree: `[[{"prim": "COMPARE"},{"prim": "GE"}],{"prim": "IF", "args":[[],{"prim": "PAIR"}]}]`,
			want: ifCmpEqMacros{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := ifFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("ifFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ifFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ifFamily.Find() = %T, want %T", got, tt.want)
			}
		})
	}
}
