package macros

import (
	"testing"

	"github.com/valyala/fastjson"
)

func Test_getPrim(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want string
	}{
		{
			name: "Array",
			tree: `[]`,
			want: "",
		}, {
			name: "Empty object",
			tree: `{}`,
			want: "",
		}, {
			name: "Prim",
			tree: `{"prim": "PAIR"}`,
			want: `"PAIR"`,
		}, {
			name: "nil",
			tree: ``,
			want: "",
		},
	}
	for _, tt := range tests {
		var tree *fastjson.Value
		if tt.tree != "" {
			tree = fastjson.MustParse(tt.tree)
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := getPrim(tree); got != tt.want {
				t.Errorf("getPrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getArgs(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want int
	}{
		{
			name: "Array",
			tree: `[]`,
		}, {
			name: "Empty object",
			tree: `{}`,
		}, {
			name: "Args",
			tree: `{"args":[{}]}`,
			want: 1,
		}, {
			name: "nil",
			tree: ``,
		},
	}
	for _, tt := range tests {
		var tree *fastjson.Value
		if tt.tree != "" {
			tree = fastjson.MustParse(tt.tree)
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := getArgs(tree); len(got) != tt.want {
				t.Errorf("getArgs() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func Test_isEq(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "EQ",
			text: `"EQ"`,
			want: true,
		}, {
			name: "NEQ",
			text: `"NEQ"`,
			want: true,
		}, {
			name: "LT",
			text: `"LT"`,
			want: true,
		}, {
			name: "LE",
			text: `"LE"`,
			want: true,
		}, {
			name: "GT",
			text: `"GT"`,
			want: true,
		}, {
			name: "GE",
			text: `"GE"`,
			want: true,
		}, {
			name: "GEQ",
			text: `"GEQ"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEq(tt.text); got != tt.want {
				t.Errorf("isEq() = %v, want %v", got, tt.want)
			}
		})
	}
}
