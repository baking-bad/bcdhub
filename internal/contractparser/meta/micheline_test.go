package meta

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func Test_unitBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
		builder  *strings.Builder
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
				builder:  &strings.Builder{},
			},
			want: `{"prim": "Unit"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("unitBuilder() %v", err)
				return
			}

			if err := unitBuilder(metadata, tt.args.path, tt.args.data, tt.args.builder); (err != nil) != tt.wantErr {
				t.Errorf("unitBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.builder.String() != tt.want {
				t.Errorf("unitBuilder() %v, want %v", tt.args.builder.String(), tt.want)
			}
		})
	}
}

func Test_defaultBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
		builder  *strings.Builder
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
				builder: &strings.Builder{},
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
				builder: &strings.Builder{},
			},
			want: `{"int": 12}`,
		}, {
			name: "'bytes' test",
			args: args{
				metadata: `{"0":{"prim":"bytes","type":"bytes"}}`,
				path:     "0",
				data: map[string]interface{}{
					"0": "0001",
				},
				builder: &strings.Builder{},
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
				builder: &strings.Builder{},
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
				builder: &strings.Builder{},
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
			if err := defaultBuilder(metadata, tt.args.path, tt.args.data, tt.args.builder); (err != nil) != tt.wantErr {
				t.Errorf("defaultBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.builder.String() != tt.want {
				t.Errorf("defaultBuilder() %v, want %v", tt.args.builder.String(), tt.want)
			}
		})
	}
}

func Test_pairBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
		builder  *strings.Builder
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
				builder: &strings.Builder{},
			},
			want: `{"prim": "Pair", "args":[{"string": "test string"},{"int": 12}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("pairBuilder() %v", err)
				return
			}

			if err := pairBuilder(metadata, tt.args.path, tt.args.data, tt.args.builder); (err != nil) != tt.wantErr {
				t.Errorf("pairBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.builder.String() != tt.want {
				t.Errorf("pairBuilder() %v, want %v", tt.args.builder.String(), tt.want)
			}
		})
	}
}

func Test_listBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
		builder  *strings.Builder
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
				builder: &strings.Builder{},
			},
			want: `[{"int": 1},{"int": 2},{"int": 3}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("listBuilder() %v", err)
				return
			}

			if err := listBuilder(metadata, tt.args.path, tt.args.data, tt.args.builder); (err != nil) != tt.wantErr {
				t.Errorf("listBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.builder.String() != tt.want {
				t.Errorf("listBuilder() %v, want %v", tt.args.builder.String(), tt.want)
			}
		})
	}
}

func Test_optionBuilder(t *testing.T) {
	type args struct {
		metadata string
		path     string
		data     map[string]interface{}
		builder  *strings.Builder
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
					"0/1/o": map[string]interface{}{
						"schemaKey": "0/1/o",
						"0/1/o/0":   "qwe",
						"0/1/o/1":   123123,
					},
					"0/1/o/0":   "qwe",
					"0/1/o/1":   123123,
					"schemaKey": "0/1/o",
				},
				builder: &strings.Builder{},
			},
			want: `{"prim": "Some", "args":[{"prim": "Pair", "args":[{"string": "qwe"},{"int": 123123}]}]}`,
		}, {
			name: "option None",
			args: args{
				metadata: `{"0":{"prim":"or","args":["0/0","0/1/o"],"type":"namedunion"},"0/0":{"fieldname":"Action","prim":"pair","args":["0/0/0","0/0/1"],"type":"namedtuple","name":"Action"},"0/0/0":{"fieldname":"action_input","prim":"or","args":["0/0/0/0/0","0/0/0/0/1/o","0/0/0/1/0","0/0/0/1/1/o"],"type":"namedunion","name":"action_input"},"0/0/0/0":{"prim":"or","type":"or"},"0/0/0/0/0":{"fieldname":"Transfer","prim":"pair","args":["0/0/0/0/0/0","0/0/0/0/0/1"],"type":"namedtuple","name":"Transfer"},"0/0/0/0/0/0":{"fieldname":"dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"dest"},"0/0/0/0/0/1":{"fieldname":"transfer_amount","prim":"mutez","type":"mutez","name":"transfer_amount"},"0/0/0/0/1":{"fieldname":"Set_pour","prim":"option","type":"option"},"0/0/0/0/1/o":{"prim":"pair","args":["0/0/0/0/1/o/0","0/0/0/0/1/o/1"],"type":"namedtuple","name":"Set_pour"},"0/0/0/0/1/o/0":{"fieldname":"pour_dest","prim":"contract","parameter":"{\"prim\":\"unit\"}","type":"contract","name":"pour_dest"},"0/0/0/0/1/o/1":{"fieldname":"pour_authorizer","prim":"key","type":"key","name":"pour_authorizer"},"0/0/0/1":{"prim":"or","type":"or"},"0/0/0/1/0":{"fieldname":"Set_keys","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1"],"type":"namedtuple","name":"Set_keys"},"0/0/0/1/0/0":{"fieldname":"key_groups","prim":"list","type":"list","name":"key_groups"},"0/0/0/1/0/0/l":{"prim":"pair","args":["0/0/0/1/0/0/l/0","0/0/0/1/0/0/l/1"],"type":"namedtuple"},"0/0/0/1/0/0/l/0":{"fieldname":"signatories","prim":"list","type":"list","name":"signatories"},"0/0/0/1/0/0/l/0/l":{"prim":"key","type":"key"},"0/0/0/1/0/0/l/1":{"fieldname":"group_threshold","prim":"nat","type":"nat","name":"group_threshold"},"0/0/0/1/0/1":{"fieldname":"overall_threshold","prim":"nat","type":"nat","name":"overall_threshold"},"0/0/0/1/1":{"fieldname":"Set_delegate","prim":"option","type":"option"},"0/0/0/1/1/o":{"prim":"key_hash","type":"key_hash","name":"Set_delegate"},"0/0/1":{"fieldname":"signatures","prim":"list","type":"list","name":"signatures"},"0/0/1/l":{"prim":"list","type":"list"},"0/0/1/l/l":{"prim":"option","type":"option"},"0/0/1/l/l/o":{"prim":"signature","type":"signature"},"0/1":{"prim":"option","type":"option"},"0/1/o":{"prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
				path:     "0/1",
				data: map[string]interface{}{
					"0/1/o": map[string]interface{}{
						"schemaKey": "none",
					},
				},
				builder: &strings.Builder{},
			},
			want: `{"prim": "None"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.args.metadata), &metadata); err != nil {
				t.Errorf("optionBuilder() %v", err)
				return
			}

			if err := optionBuilder(metadata, tt.args.path, tt.args.data, tt.args.builder); (err != nil) != tt.wantErr {
				t.Errorf("optionBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.args.builder.String() != tt.want {
				t.Errorf("optionBuilder() %v, want %v", tt.args.builder.String(), tt.want)
			}
		})
	}
}

func Test_preprocessing(t *testing.T) {
	tests := []struct {
		name    string
		binPath string
		data    map[string]interface{}
		want    map[string]interface{}
	}{
		{
			name:    "Test 1",
			binPath: "0/1/o",
			data: map[string]interface{}{
				"schemaKey": "0/1/o",
				"0/1/o/0":   "qwe",
				"0/1/o/1":   123123,
			},
			want: map[string]interface{}{
				"0/1/o": map[string]interface{}{
					"schemaKey": "0/1/o",
					"0/1/o/0":   "qwe",
					"0/1/o/1":   123123,
				},
				"0/1/o/0":   "qwe",
				"0/1/o/1":   123123,
				"schemaKey": "0/1/o",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preprocessing(tt.binPath, tt.data)
			if !reflect.DeepEqual(tt.data, tt.want) {
				t.Errorf("preprocessing() = %v, want %v", tt.data, tt.want)
			}
		})
	}
}
