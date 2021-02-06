package base

import "testing"

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
			if got != tt.want {
				t.Errorf("Node.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
