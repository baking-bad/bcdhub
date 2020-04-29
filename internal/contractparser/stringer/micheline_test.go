package stringer

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestMicheline(t *testing.T) {
	tests := []struct {
		name    string
		node    string
		want    gjson.Result
		wantErr bool
	}{
		{
			name: "Case 1: key hash",
			node: `{"bytes": "0010fc2282886d9cf8a1eebdc2733e302c7b110f38"}`,
			want: gjson.Parse(`{"string":"tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS"}`),
		}, {
			name: "Case 2: array",
			node: `[{"bytes": "0010fc2282886d9cf8a1eebdc2733e302c7b110f38"},{"bytes": "0010fc2282886d9cf8a1eebdc2733e302c7b110f38"}]`,
			want: gjson.Parse(`[{"string":"tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS"},{"string":"tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS"}]`),
		}, {
			name: "Case 3: prim",
			node: `{"prim": "PAIR", "args":[{"bytes": "0010fc2282886d9cf8a1eebdc2733e302c7b110f38"},{"bytes": "0010fc2282886d9cf8a1eebdc2733e302c7b110f38"}]}`,
			want: gjson.Parse(`{"prim":"PAIR","args":[{"string":"tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS"},{"string":"tz1MBqYpcoGU93c1bePp5A6dmwKYjmHXRopS"}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := gjson.Parse(tt.node)
			got, err := Micheline(node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Micheline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want.String() {
				t.Errorf("Micheline() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}
