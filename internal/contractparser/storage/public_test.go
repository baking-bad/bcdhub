package storage

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

func Test_findPtrJSONPath(t *testing.T) {
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
			got, err := findPtrJSONPath(tt.args.ptr, tt.args.path, data)
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

func TestEnrichEmptyPointers(t *testing.T) {
	type args struct {
		metadata string
		storage  string
	}
	tests := []struct {
		name    string
		args    args
		want    gjson.Result
		wantErr bool
	}{
		{
			name: "Empty big map",
			args: args{
				metadata: `{"0": {"fieldname":"set_metadata","prim":"big_map","type":"big_map","name":"set_metadata"}}`,
				storage:  `{"int": 1000}`,
			},
			want: emptyBigMap,
		},
		{
			name: "KT1XnAdcer9EK3qWg4GTYzZM7i3x1gu1k837",
			args: args{
				metadata: `{"0":{"prim":"pair","args":["0/0","0/1"],"type":"namedtuple"},"0/0":{"fieldname":"admin","prim":"pair","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1/o"],"type":"namedtuple","name":"admin"},"0/0/0":{"prim":"pair","type":"pair"},"0/0/0/0":{"fieldname":"admin","prim":"address","type":"address","name":"admin"},"0/0/0/1":{"fieldname":"metadata","prim":"big_map","type":"big_map","name":"metadata"},"0/0/0/1/k":{"prim":"string","type":"string"},"0/0/0/1/v":{"prim":"bytes","type":"bytes"},"0/0/1":{"prim":"pair","type":"pair"},"0/0/1/0":{"fieldname":"paused","prim":"bool","type":"bool","name":"paused"},"0/0/1/1":{"fieldname":"pending_admin","prim":"option","type":"option"},"0/0/1/1/o":{"prim":"address","type":"address","name":"pending_admin"},"0/1":{"fieldname":"assets","prim":"pair","args":["0/1/0/0/0","0/1/0/0/1","0/1/0/1/0","0/1/0/1/1","0/1/1"],"type":"namedtuple","name":"assets"},"0/1/0":{"prim":"pair","type":"pair"},"0/1/0/0":{"prim":"pair","type":"pair"},"0/1/0/0/0":{"fieldname":"ledger","prim":"big_map","type":"big_map","name":"ledger"},"0/1/0/0/0/k":{"prim":"nat","type":"nat"},"0/1/0/0/0/v":{"prim":"address","type":"address"},"0/1/0/0/1":{"fieldname":"next_token_id","prim":"nat","type":"nat","name":"next_token_id"},"0/1/0/1":{"prim":"pair","type":"pair"},"0/1/0/1/0":{"fieldname":"operators","prim":"big_map","type":"big_map","name":"operators"},"0/1/0/1/0/k":{"prim":"pair","args":["0/1/0/1/0/k/0","0/1/0/1/0/k/1/0","0/1/0/1/0/k/1/1"],"type":"tuple"},"0/1/0/1/0/k/0":{"prim":"address","type":"address"},"0/1/0/1/0/k/1":{"prim":"pair","type":"pair"},"0/1/0/1/0/k/1/0":{"prim":"address","type":"address"},"0/1/0/1/0/k/1/1":{"prim":"nat","type":"nat"},"0/1/0/1/0/v":{"prim":"unit","type":"unit"},"0/1/0/1/1":{"fieldname":"permissions_descriptor","prim":"pair","args":["0/1/0/1/1/0","0/1/0/1/1/1/0","0/1/0/1/1/1/1/0","0/1/0/1/1/1/1/1/o"],"type":"namedtuple","name":"permissions_descriptor"},"0/1/0/1/1/0":{"fieldname":"operator","prim":"or","args":["0/1/0/1/1/0/0","0/1/0/1/1/0/1/0","0/1/0/1/1/0/1/1"],"type":"namedenum","name":"operator"},"0/1/0/1/1/0/0":{"fieldname":"no_transfer","prim":"unit","type":"unit","name":"no_transfer"},"0/1/0/1/1/0/1":{"prim":"or","type":"or"},"0/1/0/1/1/0/1/0":{"fieldname":"owner_transfer","prim":"unit","type":"unit","name":"owner_transfer"},"0/1/0/1/1/0/1/1":{"fieldname":"owner_or_operator_transfer","prim":"unit","type":"unit","name":"owner_or_operator_transfer"},"0/1/0/1/1/1":{"prim":"pair","type":"pair"},"0/1/0/1/1/1/0":{"fieldname":"receiver","prim":"or","args":["0/1/0/1/1/1/0/0","0/1/0/1/1/1/0/1/0","0/1/0/1/1/1/0/1/1"],"type":"namedenum","name":"receiver"},"0/1/0/1/1/1/0/0":{"fieldname":"owner_no_hook","prim":"unit","type":"unit","name":"owner_no_hook"},"0/1/0/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/0/1/1/1/0/1/0":{"fieldname":"optional_owner_hook","prim":"unit","type":"unit","name":"optional_owner_hook"},"0/1/0/1/1/1/0/1/1":{"fieldname":"required_owner_hook","prim":"unit","type":"unit","name":"required_owner_hook"},"0/1/0/1/1/1/1":{"prim":"pair","type":"pair"},"0/1/0/1/1/1/1/0":{"fieldname":"sender","prim":"or","args":["0/1/0/1/1/1/1/0/0","0/1/0/1/1/1/1/0/1/0","0/1/0/1/1/1/1/0/1/1"],"type":"namedenum","name":"sender"},"0/1/0/1/1/1/1/0/0":{"fieldname":"owner_no_hook","prim":"unit","type":"unit","name":"owner_no_hook"},"0/1/0/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/0/1/1/1/1/0/1/0":{"fieldname":"optional_owner_hook","prim":"unit","type":"unit","name":"optional_owner_hook"},"0/1/0/1/1/1/1/0/1/1":{"fieldname":"required_owner_hook","prim":"unit","type":"unit","name":"required_owner_hook"},"0/1/0/1/1/1/1/1":{"fieldname":"custom","prim":"option","type":"option"},"0/1/0/1/1/1/1/1/o":{"prim":"pair","args":["0/1/0/1/1/1/1/1/o/0","0/1/0/1/1/1/1/1/o/1/o"],"type":"namedtuple","name":"custom"},"0/1/0/1/1/1/1/1/o/0":{"fieldname":"tag","prim":"string","type":"string","name":"tag"},"0/1/0/1/1/1/1/1/o/1":{"fieldname":"config_api","prim":"option","type":"option"},"0/1/0/1/1/1/1/1/o/1/o":{"prim":"address","type":"address","name":"config_api"},"0/1/1":{"fieldname":"token_metadata","prim":"big_map","type":"big_map","name":"token_metadata"},"0/1/1/k":{"prim":"nat","type":"nat"},"0/1/1/v":{"prim":"pair","args":["0/1/1/v/0","0/1/1/v/1/0","0/1/1/v/1/1/0","0/1/1/v/1/1/1/0","0/1/1/v/1/1/1/1"],"type":"namedtuple"},"0/1/1/v/0":{"fieldname":"token_id","prim":"nat","type":"nat","name":"token_id"},"0/1/1/v/1":{"prim":"pair","type":"pair"},"0/1/1/v/1/0":{"fieldname":"symbol","prim":"string","type":"string","name":"symbol"},"0/1/1/v/1/1":{"prim":"pair","type":"pair"},"0/1/1/v/1/1/0":{"fieldname":"name","prim":"string","type":"string","name":"name"},"0/1/1/v/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/v/1/1/1/0":{"fieldname":"decimals","prim":"nat","type":"nat","name":"decimals"},"0/1/1/v/1/1/1/1":{"fieldname":"extras","prim":"map","type":"map","name":"extras"},"0/1/1/v/1/1/1/1/k":{"prim":"string","type":"string"},"0/1/1/v/1/1/1/1/v":{"prim":"bytes","type":"bytes"}}`,
				storage:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"000014d1bd5d123190df1a298048895ff7d213ecca39"},{"int":"31948"}]},{"prim":"Pair","args":[{"prim":"False"},{"prim":"None"}]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[[{"args":[{"int":"0"},{"string":"tz1bZ9gRxi8fF26MyJ3pfipzbMt5aquWomP3"}],"prim":"Elt"}],{"int":"1"}]},{"prim":"Pair","args":[{"int":"31950"},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]},{"prim":"None"}]}]}]}]}]},[{"args":[{"int":"0"},{"args":[{"int":"0"},{"args":[{"string":"TBSTS"},{"args":[{"string":"Star Token"},{"args":[{"int":"0"},[{"args":[{"string":"uri"},{"string":"ipfs://QmQsx6PauxbAfC7uo97Mg8BBuLdfsB6auojtxLr4yGSuyY"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]]}]}`,
			},
			want: gjson.Parse(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"000014d1bd5d123190df1a298048895ff7d213ecca39"},[]]},{"prim":"Pair","args":[{"prim":"False"},{"prim":"None"}]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[[{"args":[{"int":"0"},{"string":"tz1bZ9gRxi8fF26MyJ3pfipzbMt5aquWomP3"}],"prim":"Elt"}],{"int":"1"}]},{"prim":"Pair","args":[[],{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]},{"prim":"None"}]}]}]}]}]},[{"args":[{"int":"0"},{"args":[{"int":"0"},{"args":[{"string":"TBSTS"},{"args":[{"string":"Star Token"},{"args":[{"int":"0"},[{"args":[{"string":"uri"},{"string":"ipfs://QmQsx6PauxbAfC7uo97Mg8BBuLdfsB6auojtxLr4yGSuyY"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata meta.Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("json.Unmarshal() metadata error = %v", err)
				return
			}

			storage := gjson.Parse(tt.args.storage)
			got, err := EnrichEmptyPointers(metadata, storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnrichEmptyPointers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnrichEmptyPointers() = %v, want %v", got, tt.want)
			}
		})
	}
}
