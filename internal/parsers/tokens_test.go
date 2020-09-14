package parsers

import (
	"testing"
)

func Test_normalizeParameter(t *testing.T) {
	tests := []struct {
		name   string
		params string
		want   string
	}{
		{
			name:   "Without left/right",
			params: `{"value": {"string": "test"}}`,
			want:   `{"string": "test"}`,
		}, {
			name:   "Without left/right and value",
			params: `{"string": "test"}`,
			want:   `{"string": "test"}`,
		}, {
			name:   "With left/right and value",
			params: `{"value": {"prim": "Left", "args":[{"prim": "Left", "args":[{"string": "test"}]]}}`,
			want:   `{"string": "test"}`,
		}, {
			name:   "With left/right and without value",
			params: `{"prim": "Left", "args":[{"prim": "Left", "args":[{"string": "test"}]]}`,
			want:   `{"string": "test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeParameter(tt.params); got.String() != tt.want {
				t.Errorf("normalizeParameter() = %v, want %v", got, tt.want)
			}
		})
	}
}
