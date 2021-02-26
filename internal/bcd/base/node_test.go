package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_Hash(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    string
		wantErr bool
	}{
		{
			name: "Hello world contract",
			code: `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"Hello World! :)"}]},{"prim":"FAILWITH"}]]}]`,
			want: "006c016c0243686827",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			if err := json.UnmarshalFromString(tt.code, &node); err != nil {
				t.Errorf("Node.Hash() UnmarshalFromString error = %v", err)
				return
			}

			got, err := node.Hash()
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNode_GetAnnotations(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want map[string]struct{}
	}{
		{
			name: "mainnet/KT1R3uoZ6W1ZxEwzqtv75Ro7DhVY6UAcxuK2/parameter",
			tree: `{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"or","args":[{"prim":"mutez"},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"bool"}]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"string"},{"prim":"option","args":[{"prim":"bytes"}]}]}]},{"prim":"operation"}]}]}]}],"annots":["%default"]}]}`,
			want: map[string]struct{}{
				"do":      {},
				"default": {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			if err := json.UnmarshalFromString(tt.tree, &node); err != nil {
				t.Errorf("UnmarshalFromString error=%s", err)
				return
			}
			got := node.GetAnnotations()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNode_Fingerprint(t *testing.T) {
	type args struct {
		script string
		isCode bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple prim",
			args: args{
				script: `{ "prim": "string" }`,
				isCode: false,
			},
			want:    "68",
			wantErr: false,
		}, {
			name: "simple prim 2",
			args: args{
				script: `{ "prim": "UNPAPAPAIR" }`,
				isCode: false,
			},
			want:    "8e",
			wantErr: false,
		}, {
			name: "code",
			args: args{
				script: `{ "prim": "code", "args":[{"prim": "CAST", "args":[{"prim": "string"}]}, { "prim": "string" }] }`,
				isCode: true,
			},
			want:    "0268",
			wantErr: false,
		}, {
			name: "parameter",
			args: args{
				script: `{ "prim": "parameter", "args":[{"prim": "or", "args":[{"prim": "string"}, { "string": "string" }]}]}`,
				isCode: false,
			},
			want:    "006868",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			if err := json.UnmarshalFromString(tt.args.script, &node); err != nil {
				t.Errorf("UnmarshalFromString() error = %v", err)
				return
			}
			got, err := node.Fingerprint(tt.args.isCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Fingerprint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Node.Fingerprint() = %v, want %v", got, tt.want)
			}
		})
	}
}
