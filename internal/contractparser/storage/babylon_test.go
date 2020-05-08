package storage

import (
	"log"
	"testing"

	"github.com/tidwall/gjson"
)

func TestBabylon_findPtrJSONPath(t *testing.T) {
	type args struct {
		ptr  int64
		path string
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Big map in map value: 0",
			args: args{
				ptr:  1354,
				path: "args.0.#.args.1.args.0",
				data: `{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"123"},{"prim":"Pair","args":[{"int":"1354"},{"prim":"False"}]}]}],{"prim":"Unit"}]}`,
			},
			want: "args.0.0.args.1.args.0",
		}, {
			name: "Big map in map value: 1",
			args: args{
				ptr:  1355,
				path: "args.0.#.args.1.args.0",
				data: `{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"123"},{"prim":"Pair","args":[{"int":"1354"},{"prim":"False"}]}]},{"prim":"Elt","args":[{"int":"1234"},{"prim":"Pair","args":[{"int":"1355"},{"prim":"False"}]}]}],{"prim":"Unit"}]}`,
			},
			want: "args.0.1.args.1.args.0",
		}, {
			name: "Big map in list",
			args: args{
				ptr:  3,
				path: "args.0.#",
				data: `{"prim":"Pair","args":[[{"int":"1"},{"int":"2"},{"int":"3"}],{"prim":"Unit"}]}`,
			},
			want: "args.0.2",
		}, {
			name: "Big map in list of list",
			args: args{
				ptr:  3,
				path: "args.0.#.#",
				data: `{"prim":"Pair","args":[[[{"int":"1"},{"int":"2"}],[{"int":"3"}]],{"prim":"Unit"}]}`,
			},
			want: "args.0.1.0",
		}, {
			name: "Big map in pair",
			args: args{
				ptr:  1,
				path: "args.0",
				data: `{"prim":"Pair","args":[{"int":"1"},{"prim":"Unit"}]}`,
			},
			want: "args.0",
		}, {
			name: "Big map in list of map",
			args: args{
				ptr:  1355,
				path: "args.0.#.#.args.1.args.0",
				data: `{"prim":"Pair","args":[[[{"prim":"Elt","args":[{"int":"123"},{"prim":"Pair","args":[{"int":"1354"},{"prim":"False"}]}]}],[{"prim":"Elt","args":[{"int":"123"},{"prim":"Pair","args":[{"int":"1355"},{"prim":"False"}]}]}]],{"prim":"Unit"}]}`,
			},
			want: "args.0.1.0.args.1.args.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.args.data)
			b := &Babylon{}
			got, err := b.findPtrJSONPath(tt.args.ptr, tt.args.path, data)
			log.Println("--------------------")
			if (err != nil) != tt.wantErr {
				t.Errorf("Babylon.findPtrJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Babylon.findPtrJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
