package macros

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestCollapse(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    string
		wantErr bool
	}{
		{
			name: "FAIL",
			tree: `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[{"prim":"FAIL"}]}]`,
		}, {
			name: "ASSERT",
			tree: `[{"prim":"parameter","args":[{"prim":"bool"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},[{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"bool"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"ASSERT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_EQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"EQ"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_EQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_NEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"NEQ"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_NEQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_LT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LT"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_LT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_LE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LE"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_LE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_GT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GT"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_GT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_GE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GE"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"ASSERT_GE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPEQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPNEQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPLT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPLE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPGT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_CMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"ASSERT_CMPGE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"EQ"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFEQ","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"NEQ"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFNEQ","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LT"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFLT","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LE"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFLE","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GT"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFGT","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GE"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},{"prim":"IFGE","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPEQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPNEQ"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPLT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPLE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPGT"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "CMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"CMPGE"},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPEQ","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPNEQ","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPLT","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPLE","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPGT","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "IFCMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"IFCMPGE","args":[[],[{"prim":"UNIT"}]]},{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "ASSERT_NONE",
			tree: `[{"prim":"IF_NONE","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: `{"prim":"ASSERT_NONE"}`,
		}, {
			name: "ASSERT_SOME",
			tree: `[{"prim":"IF_NONE","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]],[{"prim":"RENAME","annots":["@test"]}]]}]`,
			want: `{"prim":"ASSERT_SOME","annots":["@test"]}`,
		}, {
			name: "ASSERT_LEFT",
			tree: `[{"prim":"IF_LEFT","args":[[{"prim":"RENAME","annots":["@test"]}],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: `{"prim":"ASSERT_LEFT","annots":["@test"]}`,
		}, {
			name: "ASSERT_RIGHT",
			tree: `[{"prim":"IF_LEFT","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]],[{"prim":"RENAME","annots":["@test"]}]]}]`,
			want: `{"prim":"ASSERT_RIGHT","annots":["@test"]}`,
		}, {
			name: "SET_CAR",
			tree: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"DUP"},{"prim":"CAR","annots":["%s"]},{"prim":"DROP"},{"prim":"CDR","annots":["@%%"]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%s","%@"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},{"prim":"SET_CAR","annots":["%s"]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "SET_CAR without annots",
			tree: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"CDR","annots":["@%%"]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%","%@"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},{"prim":"SET_CAR"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "SET_CDR",
			tree: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"DUP"},{"prim":"CDR","annots":["%n"]},{"prim":"DROP"},{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%n"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},{"prim":"SET_CDR","annots":["%n"]},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "SET_CDR without annots",
			tree: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},{"prim":"SET_CDR"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "MAP_CAR",
			tree: `[{"prim":"DUP"},{"prim":"CDR","annots":["@%%"]},{"prim":"DIP","args":[[{"prim":"CAR","annots":["@ahaha"]},[{"prim":"AND"}]]]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%ahaha","%@"]}]`,
			want: `{"prim":"MAP_CAR","args":[[{"prim":"AND"}]],"annots":["%ahaha"]}`,
		}, {
			name: "MAP_CDR",
			tree: `[{"prim":"DUP"},{"prim":"CDR","annots":["@ahaha"]},[{"prim":"AND"}],{"prim":"SWAP"},{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%ahaha"]}]`,
			want: `{"prim":"MAP_CDR","args":[[{"prim":"AND"}]],"annots":["%ahaha"]}`,
		}, {
			name: "COMPARE LT IF in a raw",
			tree: `[{"prim":"COMPARE"},{"prim":"LT"},{"prim":"IF","args":[[{"prim":"UNIT"},{"prim":"TRANSFER_TOKENS"},{"prim":"DIP","args":[[{"prim":"SWAP"}]]},{"prim":"CONS"}],[{"prim":"DROP"},{"prim":"DROP"},{"prim":"SWAP"}]]}]`,
			want: `[{"prim":"COMPARE"},{"prim":"LT"},{"prim":"IF","args":[[{"prim":"UNIT"},{"prim":"TRANSFER_TOKENS"},{"prim":"DIP","args":[[{"prim":"SWAP"}]]},{"prim":"CONS"}],[{"prim":"DROP"},{"prim":"DROP"},{"prim":"SWAP"}]]}]`,
		}, {
			name: "UNPAIR",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
			want: `{"prim":"UNPAIR","annots":["%a","%b"]}`,
		}, {
			name: "UNPAIR without annots",
			tree: `[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]`,
			want: `{"prim":"UNPAIR"}`,
		}, {
			name: "CADR",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":l"]},{"prim":"nat","annots":[":r"]}]}]},{"prim":"storage","args":[{"prim":"nat"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"CAR","annots":["@test"]}],{"prim":"PAIR"}]}]`,
			want: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":l"]},{"prim":"nat","annots":[":r"]}]}]},{"prim":"storage","args":[{"prim":"nat"}]},{"prim":"code","args":[{"prim":"CAAR","annots":["@test"]},{"prim":"PAIR"}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := gjson.Parse(tt.tree)

			got, err := Collapse(tree, GetAllFamilies())
			if (err != nil) != tt.wantErr {
				t.Errorf("Collapse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("Collapse() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestHasMacros(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want bool
	}{
		{
			name: "FAIL",
			tree: `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: true,
		}, {
			name: "ASSERT",
			tree: `[{"prim":"parameter","args":[{"prim":"bool"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},[{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_EQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"EQ"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_NEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"NEQ"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_LT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LT"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_LE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LE"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_GT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GT"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_GE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GE"},{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_CMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"EQ"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"NEQ"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LT"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"LE"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GT"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},{"prim":"COMPARE"},[{"prim":"GE"},{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "CMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPNEQ",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"NEQ"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPLT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LT"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPLE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"LE"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPGT",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GT"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "IFCMPGE",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]},[[{"prim":"COMPARE"},{"prim":"GE"}],{"prim":"IF","args":[[],[{"prim":"UNIT"}]]}],{"prim":"UNIT"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "ASSERT_NONE",
			tree: `[{"prim":"IF_NONE","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: true,
		}, {
			name: "ASSERT_SOME",
			tree: `[{"prim":"IF_NONE","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]],[{"prim":"RENAME","annots":["@test"]}]]}]`,
			want: true,
		}, {
			name: "ASSERT_LEFT",
			tree: `[{"prim":"IF_LEFT","args":[[{"prim":"RENAME","annots":["@test"]}],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}]`,
			want: true,
		}, {
			name: "ASSERT_RIGHT",
			tree: `[{"prim":"IF_LEFT","args":[[[{"prim":"UNIT"},{"prim":"FAILWITH"}]],[{"prim":"RENAME","annots":["@test"]}]]}]`,
			want: true,
		}, {
			name: "SET_CAR",
			tree: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"DUP"},{"prim":"CAR","annots":["%s"]},{"prim":"DROP"},{"prim":"CDR","annots":["@%%"]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%s","%@"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "SET_CAR without annots",
			tree: `[{"prim":"parameter","args":[{"prim":"string"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"CDR","annots":["@%%"]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%","%@"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "SET_CDR",
			tree: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"DUP"},{"prim":"CDR","annots":["%n"]},{"prim":"DROP"},{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%n"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "SET_CDR without annots",
			tree: `[{"prim":"parameter","args":[{"prim":"nat"}]},{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"string","annots":["%s"]},{"prim":"nat","annots":["%n"]}]}]},{"prim":"code","args":[[{"prim":"DUP"},{"prim":"CDR"},{"prim":"DIP","args":[[{"prim":"CAR"}]]},[{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%"]}],{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
			want: true,
		}, {
			name: "MAP_CAR",
			tree: `[{"prim":"DUP"},{"prim":"CDR","annots":["@%%"]},{"prim":"DIP","args":[[{"prim":"CAR","annots":["@ahaha"]},[{"prim":"AND"}]]]},{"prim":"SWAP"},{"prim":"PAIR","annots":["%ahaha","%@"]}]`,
			want: true,
		}, {
			name: "MAP_CDR",
			tree: `[{"prim":"DUP"},{"prim":"CDR","annots":["@ahaha"]},[{"prim":"AND"}],{"prim":"SWAP"},{"prim":"CAR","annots":["@%%"]},{"prim":"PAIR","annots":["%@","%ahaha"]}]`,
			want: true,
		}, {
			name: "COMPARE LT IF in a raw",
			tree: `[{"prim":"COMPARE"},{"prim":"LT"},{"prim":"IF","args":[[{"prim":"UNIT"},{"prim":"TRANSFER_TOKENS"},{"prim":"DIP","args":[[{"prim":"SWAP"}]]},{"prim":"CONS"}],[{"prim":"DROP"},{"prim":"DROP"},{"prim":"SWAP"}]]}]`,
			want: false,
		}, {
			name: "UNPAIR",
			tree: `[{"prim":"DUP"},{"prim":"CAR","annots":["%a"]},{"prim":"DIP","args":[[{"prim":"CDR","annots":["%b"]}]]}]`,
			want: true,
		}, {
			name: "UNPAIR without annots",
			tree: `[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]`,
			want: true,
		}, {
			name: "CADR",
			tree: `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":l"]},{"prim":"nat","annots":[":r"]}]}]},{"prim":"storage","args":[{"prim":"nat"}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"CAR","annots":["@test"]}],{"prim":"PAIR"}]}]`,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := gjson.Parse(tt.tree)

			got, err := HasMacros(tree, GetAllFamilies())
			if err != nil {
				t.Errorf("HasMacros() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("HasMacros() = %v, want %v", got, tt.want)
			}
		})
	}
}
