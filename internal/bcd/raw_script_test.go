package bcd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawScript_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantErr       bool
		wantCode      []byte
		wantParameter []byte
		wantStorage   []byte
	}{
		{
			name:          "test 1",
			data:          []byte(`[{"prim":"code","args":[{"prim":"code"}]},{"prim":"storage","args":[{"prim":"storage"}]},{"prim":"parameter","args":[{"prim":"parameter"}]}]`),
			wantCode:      []byte(`[{"prim":"code"}]`),
			wantParameter: []byte(`[{"prim":"parameter"}]`),
			wantStorage:   []byte(`[{"prim":"storage"}]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s RawScript
			err := s.UnmarshalJSON(tt.data)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}

			require.ElementsMatch(t, tt.wantCode, s.Code)
			require.ElementsMatch(t, tt.wantParameter, s.Parameter)
			require.ElementsMatch(t, tt.wantStorage, s.Storage)
		})
	}
}
