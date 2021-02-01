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
			name: "Empty string",
			code: "",
			want: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		}, {
			name: "Test string",
			code: "test",
			want: "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff",
		}, {
			name: "Hello world contract",
			code: `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"unit"}]},{"prim":"code","args":[[{"prim":"PUSH","args":[{"prim":"string"},{"string":"Hello World! :)"}]},{"prim":"FAILWITH"}]]}]`,
			want: "892c8b82a76c19c5986909e83ca79560f6f9a5fe87952ccf8969cc7d508635413c42565a2ed922bf14d397652314f97b63607307ae9e4692c816578ceceed210",
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
