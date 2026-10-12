package contract

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSections_Empty(t *testing.T) {
	tests := []struct {
		name     string
		sections Sections
		want     bool
	}{
		{
			name: "zero value",
			want: true,
		}, {
			name: "null sections",
			sections: Sections{
				Parameter:  []byte("null"),
				ReturnType: []byte("null"),
				Code:       []byte("null"),
			},
			want: true,
		}, {
			name: "mixed null and absent sections",
			sections: Sections{
				Parameter: []byte("null"),
				Code:      []byte{},
			},
			want: true,
		}, {
			name: "only code",
			sections: Sections{
				Code: []byte(`[{"prim":"CAR"}]`),
			},
			want: false,
		}, {
			name: "only parameter",
			sections: Sections{
				Parameter: []byte(`{"prim":"nat"}`),
			},
			want: false,
		}, {
			name: "only return type",
			sections: Sections{
				ReturnType: []byte(`{"prim":"nat"}`),
			},
			want: false,
		}, {
			name: "all sections",
			sections: Sections{
				Parameter:  []byte(`{"prim":"nat"}`),
				ReturnType: []byte(`{"prim":"nat"}`),
				Code:       []byte(`[{"prim":"CAR"}]`),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.sections.Empty())
		})
	}
}
