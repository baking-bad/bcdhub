package macros

import (
	"reflect"
	"testing"

	"github.com/valyala/fastjson"
)

func Test_mapFamily_Find(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		want    Macros
		wantErr bool
	}{
		{
			name: "invalid length",
			tree: `[{},{}]`,
		}, {
			name: "invalid 1 prim",
			tree: `[{"prim": "DIP"},{"prim": "CDR"},{"prim": "PAIR"},{"prim": "SWAP"},{"prim": "CAR"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid 2 prim",
			tree: `[{"prim": "DUP"},{"prim": "CAR"},{"prim": "PAIR"},{"prim": "SWAP"},{"prim": "CAR"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid 4 prim: MAP_CDR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "PAIR"},{"prim": "SWUP"},{"prim": "CAR"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid 5 prim: MAP_CDR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "PAIR"},{"prim": "SWAP"},{"prim": "CDR"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid 6 prim: MAP_CDR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "PAIR"},{"prim": "SWAP"},{"prim": "CAR"},{"prim": "PIR"}]`,
		}, {
			name: "MAP_CDR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},[{"prim": "PAIR"}],{"prim": "SWAP"},{"prim": "CAR"},{"prim": "PAIR"}]`,
			want: mapCdrMacros{},
		}, {
			name: "invalid 4 prim: MAP_CAR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "DIP"},{"prim": "SWUP"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid 5 prim: MAP_CAR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "DIP"},{"prim": "SWAP"},{"prim": "PIR"}]`,
		}, {
			name: "invalid dip args length: MAP_CAR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "DIP"},{"prim": "SWAP"},{"prim": "PAIR"}]`,
		}, {
			name: "invalid dip arg 0: MAP_CAR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "DIP", "args":[{"prim": "CDR"},{"prim": "PAIR"}]},{"prim": "SWAP"},{"prim": "PAIR"}]`,
		}, {
			name: "MAP_CAR",
			tree: `[{"prim": "DUP"},{"prim": "CDR"},{"prim": "DIP", "args":[[{"prim": "CAR"},[{"prim": "PAIR"}]]]},{"prim": "SWAP"},{"prim": "PAIR"}]`,
			want: mapCarMacros{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := mapFamily{}
			tree, err := fastjson.MustParse(tt.tree).Array()
			if err != nil {
				t.Errorf("mapFamily.Find() error = %v", err)
				return
			}
			got, err := f.Find(tree...)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapFamily.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapFamily.Find() = %T, want %T", got, tt.want)
			}
		})
	}
}
