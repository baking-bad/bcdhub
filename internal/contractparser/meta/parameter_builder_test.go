package meta

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

func Test_unitBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "'Unit' test",
			args: args{
				metadata: `{"0":{"prim":"unit","type":"unit"}}`,
				path:     "0",
				data:     nil,
			},
			want: `{"prim": "Unit"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("unitParameterBuilder() %v", err)
				return
			}

			pb := unitParameterBuilder{}
			res, err := pb.Build(metadata[tt.args.path], tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("unitParameterBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if res != tt.want {
				t.Errorf("unitParameterBuilder() %v, want %v", res, tt.want)
			}
		})
	}
}

func Test_defaultBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "'string' test",
			args: args{
				metadata: `{"0":{"prim":"string","type":"string"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": "test string",
				},
			},
			want: `{"string": "test string"}`,
		}, {
			name: "'int' test",
			args: args{
				metadata: `{"0":{"prim":"int","type":"int"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": 12,
				},
			},
			want: `{"int": "12"}`,
		}, {
			name: "'bytes' test",
			args: args{
				metadata: `{"0":{"prim":"bytes","type":"bytes"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": "0001",
				},
			},
			want: `{"bytes": "0001"}`,
		}, {
			name: "'bool' true test",
			args: args{
				metadata: `{"0":{"prim":"bool","type":"bool"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": true,
				},
			},
			want: `{"prim": "True"}`,
		}, {
			name: "'bool' false test",
			args: args{
				metadata: `{"0":{"prim":"bool","type":"bool"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": false,
				},
			},
			want: `{"prim": "False"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("defaultBuilder() %v", err)
				return
			}

			pb := defaultParameterBuilder{
				validate: false,
			}
			res, err := pb.Build(metadata[tt.args.path], tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultParameterBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if res != tt.want {
				t.Errorf("defaultParameterBuilder() %v, want %v", res, tt.want)
			}
		})
	}
}

func Test_pairBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple pair",
			args: args{
				metadata: `{"0":{"prim":"pair","args":["0/0","0/1"],"type":"tuple"},"0/0":{"prim":"string","type":"string"},"0/1":{"prim":"nat","type":"nat"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0/0": "test string",
					"0/1": 12,
				},
			},
			want: `{"prim": "Pair", "args":[{"string": "test string"}, {"int": "12"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("pairParameterBuilder() %v", err)
				return
			}

			builder := NewParameterBuilder(metadata, false)
			res, err := builder.parameterBuilders[consts.PAIR].Build(metadata[tt.args.path], tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("pairParameterBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if res != tt.want {
				t.Errorf("pairParameterBuilder() %v, want %v", res, tt.want)
			}
		})
	}
}

func Test_listBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "list of int",
			args: args{
				metadata: `{"0":{"prim":"list","type":"list"},"0/l":{"prim":"int","type":"int"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": []int{1, 2, 3},
				},
			},
			want: `[{"int": "1"},{"int": "2"},{"int": "3"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("listParameterBuilder() %v", err)
				return
			}

			builder := NewParameterBuilder(metadata, false)
			res, err := builder.parameterBuilders[consts.LIST].Build(metadata[tt.args.path], tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("listParameterBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if res != tt.want {
				t.Errorf("listParameterBuilder() %v, want %v", res, tt.want)
			}
		})
	}
}

func Test_optionBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "option Some",
			args: args{
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
				path:     "0/1",
				data: map[string]interface{}{
					"0/1": map[string]interface{}{
						"schemaKey": "0/1/o",
						"0/1/o/0":   "qwe",
						"0/1/o/1":   123123,
					},
					"0/1/o/0":   "qwe",
					"0/1/o/1":   123123,
					"schemaKey": "0/1/o",
				},
			},
			want: `{"prim": "Some", "args":[{"prim": "Pair", "args":[{"string": "qwe"}, {"int": "123123"}]}]}`,
		}, {
			name: "option None",
			args: args{
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
				path:     "0/1",
				data: map[string]interface{}{
					"0/1": map[string]interface{}{
						"schemaKey": "none",
					},
				},
			},
			want: `{"prim": "None"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("optionParameterBuilder() %v", err)
				return
			}
			builder := NewParameterBuilder(metadata, false)
			res, err := builder.parameterBuilders[consts.OPTION].Build(metadata[tt.args.path], tt.args.path, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("optionParameterBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if res != tt.want {
				t.Errorf("optionParameterBuilder() %v, want %v", res, tt.want)
			}
		})
	}
}

func TestMetadata_BuildEntrypointMicheline(t *testing.T) {
	type args struct {
		binaryPath string
		data       map[string]interface{}
	}
	tests := []struct {
		name     string
		metadata string
		args     args
		want     string
		wantErr  bool
	}{
		{
			name:     "Vested contract: Action Set_pour Some",
			metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			args: args{
				binaryPath: "0/0",
				data: map[string]interface{}{
					"0/0/0": map[string]interface{}{
						"0/0/0/0/1": map[string]interface{}{
							"0/0/0/0/1/o/0": "safaf",
							"0/0/0/0/1/o/1": "qw34gg",
							"schemaKey":     "some",
						},
						"schemaKey": "0/0/0/0/1",
					},
					"0/0/1": []string{},
				},
			},
			want: `{"entrypoint": "Action", "value": {"prim": "Pair", "args":[{"prim": "Left", "args":[{"prim": "Right", "args":[{"prim": "Some", "args":[{"prim":"Pair", "args":[{"string": "safaf"},{"string": "qw34gg"}]}]}]}]},[]]}}`,
		}, {
			name:     "Vested contract: Action Set_pour None",
			metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			args: args{
				binaryPath: "0/0",
				data: map[string]interface{}{
					"0/0/0": map[string]interface{}{
						"0/0/0/0/1": map[string]interface{}{
							"schemaKey": "none",
						},
						"schemaKey": "0/0/0/0/1",
					},
					"0/0/1": []string{},
				},
			},
			want: `{"entrypoint": "Action", "value": {"prim": "Pair", "args":[{"prim": "Left", "args":[{"prim": "Right", "args":[{"prim": "None"}]}]},[]]}}`,
		}, {
			name:     "Vested contract: default 0/1 None",
			metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			args: args{
				binaryPath: "0/1/o",
				data: map[string]interface{}{
					"schemaKey": "none",
				},
			},
			want: `{"entrypoint": "default", "value": {"prim": "Right", "args":[{"prim": "None"}]}}`,
		}, {
			name:     "Vested contract: default 0/1 Some",
			metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			args: args{
				binaryPath: "0/1/o",
				data: map[string]interface{}{
					"0/1/o/0":   "asdasd",
					"0/1/o/1":   123123,
					"schemaKey": "some",
				},
			},
			want: `{"entrypoint": "default", "value": {"prim": "Right", "args":[{"prim": "Some", "args":[{"prim": "Pair", "args":[{"string": "asdasd"}, {"int": "123123"}]}]}]}}`,
		}, {
			name:     "KT1FQXNH13MAE2WFKyQ5wLC8nVGsaJ55rJjZ: addVoters map",
			metadata: `{"0":{"prim":"or","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1","0/1/0","0/1/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"fieldname":"addAdmin","prim":"address","type":"address","name":"addAdmin"},"0/0/0/1":{"fieldname":"addVoters","prim":"map","type":"map","name":"addVoters"},"0/0/0/1/k":{"prim":"address","type":"address"},"0/0/0/1/v":{"prim":"nat","type":"nat"},"0/0/1":{"prim":"or","type":"or"},"0/0/1/0":{"fieldname":"init","prim":"string","type":"string","name":"init"},"0/0/1/1":{"fieldname":"removeAdmin","prim":"address","type":"address","name":"removeAdmin"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"fieldname":"startVote","prim":"int","type":"int","name":"startVote"},"0/1/1":{"fieldname":"vote","prim":"nat","type":"nat","name":"vote"}}`,
			args: args{
				binaryPath: "0/0/0/1",
				data: map[string]interface{}{
					"0/0/0/1": []map[string]interface{}{
						{
							"0/0/0/1/k": "adasd",
							"0/0/0/1/v": 123,
						},
					},
				},
			},
			want: `{"entrypoint": "addVoters", "value": [{"prim": "Elt", "args":[{"string": "adasd"}, {"int": "123"}]}]}`,
		}, {
			name:     "KT1XPtZhCr2zp8dzqq6hVjNk9jWvAXoMocf8: default",
			metadata: `{"0":{"prim":"nat","type":"nat"}}`,
			args: args{
				binaryPath: "0",
				data: map[string]interface{}{
					"0": 123,
				},
			},
			want: `{"entrypoint": "default", "value": {"int": "123"}}`,
		}, {
			name:     "KT1EXw1mE1Nqtq2oxLQexSiug2ZbH9LkjVJR: default",
			metadata: `{"0":{"prim":"pair","args":["0/0","0/1/0","0/1/1"],"type":"tuple"},"0/0":{"prim":"or","args":["0/0/0","0/0/1"],"type":"union"},"0/0/0":{"prim":"lambda","parameter":"{\"prim\":\"unit\"}","type":"lambda"},"0/0/1":{"prim":"set","type":"set"},"0/0/1/s":{"prim":"key_hash","type":"key_hash"},"0/1":{"prim":"pair","type":"pair"},"0/1/0":{"prim":"nat","type":"nat"},"0/1/1":{"prim":"list","type":"list"},"0/1/1/l":{"prim":"pair","args":["0/1/1/l/0","0/1/1/l/1"],"type":"tuple"},"0/1/1/l/0":{"prim":"key","type":"key"},"0/1/1/l/1":{"prim":"signature","type":"signature"}}`,
			args: args{
				binaryPath: "0",
				data: map[string]interface{}{
					"0/0": map[string]interface{}{
						"schemaKey": "0/0/0",
						"0/0/0":     "{ PUSH nat 42 ; DUP ; DROP }",
					},
					"0/1/1": []map[string]interface{}{
						{
							"0/1/1/l/0": "sdsdg",
							"0/1/1/l/1": "asdgsdfg",
						},
					},
					"0/1/0": 234,
				},
			},
			want: `{"entrypoint": "default", "value": {"prim": "Pair", "args":[{"prim": "Left", "args":[[[{"args":[{"prim": "nat"}, {"int": "42"}], "prim": "PUSH"}, {"prim": "DUP"}, {"prim": "DROP"}]]]}, {"prim": "Pair", "args":[{"int": "234"}, [{"prim": "Pair", "args":[{"string": "sdsdg"}, {"string": "asdgsdfg"}]}]]}]}}`,
		}, {
			name:     "KT1UXMkUNEsSaugsL19SEgiWWPGgDLB4gPjd: updateCountryRestrictions",
			metadata: `{"0":{"prim":"or","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1","0/1/0/0","0/1/0/1","0/1/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"fieldname":"addToken","prim":"address","type":"address","name":"addToken"},"0/0/0/1":{"fieldname":"checkTransfer","prim":"pair","args":["0/0/0/1/0","0/0/0/1/1"],"type":"namedtuple","name":"checkTransfer"},"0/0/0/1/0":{"fieldname":"0","prim":"address","type":"address","name":"0"},"0/0/0/1/1":{"fieldname":"1","prim":"address","type":"address","name":"1"},"0/0/1":{"prim":"or","type":"or"},"0/0/1/0":{"fieldname":"setAccount","prim":"pair","args":["0/0/1/0/0","0/0/1/0/1"],"type":"namedtuple","name":"setAccount"},"0/0/1/0/0":{"fieldname":"0","prim":"address","type":"address","name":"0"},"0/0/1/0/1":{"fieldname":"1","prim":"pair","args":["0/0/1/0/1/0","0/0/1/0/1/1"],"type":"namedtuple","name":"1"},"0/0/1/0/1/0":{"fieldname":"registrar","prim":"address","type":"address","name":"registrar"},"0/0/1/0/1/1":{"fieldname":"restricted","prim":"bool","type":"bool","name":"restricted"},"0/0/1/1":{"fieldname":"setRegistrar","prim":"pair","args":["0/0/1/1/0","0/0/1/1/1"],"type":"namedtuple","name":"setRegistrar"},"0/0/1/1/0":{"fieldname":"0","prim":"address","type":"address","name":"0"},"0/0/1/1/1":{"fieldname":"1","prim":"bool","type":"bool","name":"1"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"prim":"or","type":"or"},"0/1/0/0":{"fieldname":"setToken","prim":"pair","args":["0/1/0/0/0","0/1/0/0/1"],"type":"namedtuple","name":"setToken"},"0/1/0/0/0":{"fieldname":"0","prim":"address","type":"address","name":"0"},"0/1/0/0/1":{"fieldname":"1","prim":"bool","type":"bool","name":"1"},"0/1/0/1":{"fieldname":"updateCountryRestrictions","prim":"list","type":"list","name":"updateCountryRestrictions"},"0/1/0/1/l":{"prim":"pair","args":["0/1/0/1/l/0","0/1/0/1/l/1"],"type":"namedtuple"},"0/1/0/1/l/0":{"fieldname":"0","prim":"nat","type":"nat","name":"0"},"0/1/0/1/l/1":{"fieldname":"1","prim":"pair","args":["0/1/0/1/l/1/0/0","0/1/0/1/l/1/0/1","0/1/0/1/l/1/1/0","0/1/0/1/l/1/1/1"],"type":"namedtuple","name":"1"},"0/1/0/1/l/1/0":{"prim":"pair","type":"pair"},"0/1/0/1/l/1/0/0":{"fieldname":"country_invest_limit","prim":"nat","type":"nat","name":"country_invest_limit"},"0/1/0/1/l/1/0/1":{"fieldname":"min_rating","prim":"nat","type":"nat","name":"min_rating"},"0/1/0/1/l/1/1":{"prim":"pair","type":"pair"},"0/1/0/1/l/1/1/0":{"fieldname":"rating_restrictions","prim":"map","type":"map","name":"rating_restrictions"},"0/1/0/1/l/1/1/0/k":{"prim":"nat","type":"nat"},"0/1/0/1/l/1/1/0/v":{"prim":"nat","type":"nat"},"0/1/0/1/l/1/1/1":{"fieldname":"vesting","prim":"timestamp","type":"timestamp","name":"vesting"},"0/1/1":{"fieldname":"updateGlobalLimit","prim":"nat","type":"nat","name":"updateGlobalLimit"}}`,
			args: args{
				binaryPath: "0/1/0/1",
				data: map[string]interface{}{
					"0/1/0/1": []map[string]interface{}{
						{
							"0/1/0/1/l/1/1/0": []string{},
							"0/1/0/1/l/0":     234,
							"0/1/0/1/l/1/0/0": 1234,
							"0/1/0/1/l/1/0/1": 14312,
							"0/1/0/1/l/1/1/1": "2020-04-01T00:01:00+03:00",
						},
					},
				},
			},
			want: `{"entrypoint": "updateCountryRestrictions", "value": [{"prim": "Pair", "args":[{"int": "234"}, {"prim": "Pair", "args":[{"prim": "Pair", "args":[{"int": "1234"}, {"int": "14312"}]}, {"prim": "Pair", "args":[[], {"int": "1585688460"}]}]}]}]}`,
		}, {
			name:     "tzBTC: upgrade",
			metadata: `{"0":{"prim":"or","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1/0","0/0/1/1/1","0/1/0/0","0/1/0/1","0/1/1/0","0/1/1/1/0","0/1/1/1/1/0/0/0/0","0/1/1/1/1/0/0/0/1","0/1/1/1/1/0/0/1/0","0/1/1/1/1/0/0/1/1","0/1/1/1/1/0/1/0/0","0/1/1/1/1/0/1/0/1","0/1/1/1/1/0/1/1/0","0/1/1/1/1/0/1/1/1","0/1/1/1/1/1/0/0/0","0/1/1/1/1/1/0/0/1","0/1/1/1/1/1/0/1/0","0/1/1/1/1/1/0/1/1","0/1/1/1/1/1/1/0/0","0/1/1/1/1/1/1/0/1","0/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"fieldname":"getVersion","prim":"pair","args":["0/0/0/0/0","0/0/0/0/1"],"type":"tuple","name":"getVersion"},"0/0/0/0/0":{"prim":"unit","type":"unit"},"0/0/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/0/1":{"fieldname":"getAllowance","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1","0/0/0/1/1"],"type":"namedtuple","name":"getAllowance"},"0/0/0/1/0":{"prim":"pair","type":"pair"},"0/0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/0/1/0/1":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/0/0/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1":{"prim":"or","type":"or"},"0/0/1/0":{"fieldname":"getBalance","prim":"pair","args":["0/0/1/0/0","0/0/1/0/1"],"type":"namedtuple","name":"getBalance"},"0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1":{"prim":"or","type":"or"},"0/0/1/1/0":{"fieldname":"getTotalSupply","prim":"pair","args":["0/0/1/1/0/0","0/0/1/1/0/1"],"type":"tuple","name":"getTotalSupply"},"0/0/1/1/0/0":{"prim":"unit","type":"unit"},"0/0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1/1":{"fieldname":"getTotalMinted","prim":"pair","args":["0/0/1/1/1/0","0/0/1/1/1/1"],"type":"tuple","name":"getTotalMinted"},"0/0/1/1/1/0":{"prim":"unit","type":"unit"},"0/0/1/1/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"prim":"or","type":"or"},"0/1/0/0":{"fieldname":"getTotalBurned","prim":"pair","args":["0/1/0/0/0","0/1/0/0/1"],"type":"tuple","name":"getTotalBurned"},"0/1/0/0/0":{"prim":"unit","type":"unit"},"0/1/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1/0/1":{"fieldname":"getOwner","prim":"pair","args":["0/1/0/1/0","0/1/0/1/1"],"type":"tuple","name":"getOwner"},"0/1/0/1/0":{"prim":"unit","type":"unit"},"0/1/0/1/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/1":{"prim":"or","type":"or"},"0/1/1/0":{"fieldname":"getRedeemAddress","prim":"pair","args":["0/1/1/0/0","0/1/1/0/1"],"type":"tuple","name":"getRedeemAddress"},"0/1/1/0/0":{"prim":"unit","type":"unit"},"0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/0":{"fieldname":"getTokenMetadata","prim":"pair","args":["0/1/1/1/0/0","0/1/1/1/0/1"],"type":"tuple","name":"getTokenMetadata"},"0/1/1/1/0/0":{"prim":"list","type":"list"},"0/1/1/1/0/0/l":{"prim":"nat","type":"nat"},"0/1/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"list\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"nat\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"nat\"},{\"prim\":\"map\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"string\"}]}]}]}]}]}]}","type":"contract"},"0/1/1/1/1":{"fieldname":"safeEntrypoints","prim":"or","type":"or"},"0/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0/0":{"fieldname":"run","prim":"pair","args":["0/1/1/1/1/0/0/0/0/0","0/1/1/1/1/0/0/0/0/1"],"type":"tuple","name":"run"},"0/1/1/1/1/0/0/0/0/0":{"prim":"string","type":"string"},"0/1/1/1/1/0/0/0/0/1":{"prim":"bytes","type":"bytes"},"0/1/1/1/1/0/0/0/1":{"fieldname":"upgrade","prim":"pair","args":["0/1/1/1/1/0/0/0/1/0/0","0/1/1/1/1/0/0/0/1/0/1","0/1/1/1/1/0/0/0/1/1/0","0/1/1/1/1/0/0/0/1/1/1/0/o","0/1/1/1/1/0/0/0/1/1/1/1/o"],"type":"tuple","name":"upgrade"},"0/1/1/1/1/0/0/0/1/0":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/0/0":{"typename":"currentVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/0/1":{"typename":"newVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/0":{"typename":"migrationScript","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/1/0":{"typename":"newCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/0/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1/1":{"typename":"newPermCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/1/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"unit\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/1/0":{"fieldname":"epwBeginUpgrade","prim":"pair","args":["0/1/1/1/1/0/0/1/0/0","0/1/1/1/1/0/0/1/0/1"],"type":"namedtuple","name":"epwBeginUpgrade"},"0/1/1/1/1/0/0/1/0/0":{"typename":"current","prim":"nat","type":"nat","name":"current"},"0/1/1/1/1/0/0/1/0/1":{"typename":"new","prim":"nat","type":"nat","name":"new"},"0/1/1/1/1/0/0/1/1":{"typename":"migrationscript","fieldname":"epwApplyMigration","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda","name":"epwApplyMigration"},"0/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0/0":{"typename":"contractcode","fieldname":"epwSetCode","prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda","name":"epwSetCode"},"0/1/1/1/1/0/1/0/1":{"fieldname":"epwFinishUpgrade","prim":"unit","type":"unit","name":"epwFinishUpgrade"},"0/1/1/1/1/0/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/1/0":{"fieldname":"transfer","prim":"pair","args":["0/1/1/1/1/0/1/1/0/0","0/1/1/1/1/0/1/1/0/1/0","0/1/1/1/1/0/1/1/0/1/1"],"type":"namedtuple","name":"transfer"},"0/1/1/1/1/0/1/1/0/0":{"typename":"from","prim":"address","type":"address","name":"from"},"0/1/1/1/1/0/1/1/0/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/1/1/0/1/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/0/1/1/0/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/0/1/1/1":{"fieldname":"approve","prim":"pair","args":["0/1/1/1/1/0/1/1/1/0","0/1/1/1/1/0/1/1/1/1"],"type":"namedtuple","name":"approve"},"0/1/1/1/1/0/1/1/1/0":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/1/1/1/1/0/1/1/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0/0":{"fieldname":"mint","prim":"pair","args":["0/1/1/1/1/1/0/0/0/0","0/1/1/1/1/1/0/0/0/1"],"type":"namedtuple","name":"mint"},"0/1/1/1/1/1/0/0/0/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/1/0/0/0/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1/0/0/1":{"typename":"value","fieldname":"burn","prim":"nat","type":"nat","name":"burn"},"0/1/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/1/0":{"typename":"operator","fieldname":"addOperator","prim":"address","type":"address","name":"addOperator"},"0/1/1/1/1/1/0/1/1":{"typename":"operator","fieldname":"removeOperator","prim":"address","type":"address","name":"removeOperator"},"0/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0/0":{"typename":"redeem","fieldname":"setRedeemAddress","prim":"address","type":"address","name":"setRedeemAddress"},"0/1/1/1/1/1/1/0/1":{"fieldname":"pause","prim":"unit","type":"unit","name":"pause"},"0/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/0":{"fieldname":"unpause","prim":"unit","type":"unit","name":"unpause"},"0/1/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/1/0":{"typename":"newOwner","fieldname":"transferOwnership","prim":"address","type":"address","name":"transferOwnership"},"0/1/1/1/1/1/1/1/1/1":{"fieldname":"acceptOwnership","prim":"unit","type":"unit","name":"acceptOwnership"}}`,
			args: args{
				binaryPath: "0/1/1/1/1/0/0/0/1",
				data: map[string]interface{}{
					"0/1/1/1/1/0/0/0/1/0/0":     1,
					"0/1/1/1/1/0/0/0/1/0/1":     1,
					"0/1/1/1/1/0/0/0/1/1/0":     "{ PUSH nat 42 ; DUP ; DROP }",
					"0/1/1/1/1/0/0/0/1/1/1/0/o": map[string]interface{}{"schemaKey": "none"},
					"0/1/1/1/1/0/0/0/1/1/1/1/o": map[string]interface{}{"schemaKey": "none"},
				},
			},
			want: `{"entrypoint": "upgrade", "value": {"prim": "Pair", "args":[{"prim": "Pair", "args":[{"int": "1"},{"int": "1"}]},{"prim": "Pair", "args":[[[{"args":[{"prim": "nat"}, {"int": "42"}], "prim": "PUSH"}, {"prim": "DUP"}, {"prim": "DROP"}]], {"prim": "Pair", "args":[{"prim": "None"},{"prim": "None"}]}]}]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.metadata), &metadata); err != nil {
				t.Errorf("BuildEntrypointMicheline() %v", err)
				return
			}
			var want map[string]interface{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Errorf("BuildEntrypointMicheline() %v", err)
				return
			}
			got, err := metadata.BuildEntrypointMicheline(tt.args.binaryPath, tt.args.data, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metadata.BuildEntrypointMicheline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Value(), want) {
				t.Errorf("Metadata.BuildEntrypointMicheline() = %v, want %v", got.Value(), want)
			}
		})
	}
}
