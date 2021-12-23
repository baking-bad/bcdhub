package bcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			if err := s.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("RawScript.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.ElementsMatch(t, tt.wantCode, s.Code) {
				return
			}
			if !assert.ElementsMatch(t, tt.wantParameter, s.Parameter) {
				return
			}
			if !assert.ElementsMatch(t, tt.wantStorage, s.Storage) {
				return
			}
		})
	}
}
