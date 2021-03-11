package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMichelsonStorageView_GetCode(t *testing.T) {
	type fields struct {
		Parameter  []byte
		Code       []byte
		ReturnType []byte
		Name       string
	}
	tests := []struct {
		name        string
		fields      fields
		storageType []byte
		want        string
		wantErr     bool
	}{
		{
			name: "test 1",
			fields: fields{
				Parameter:  []byte(`{"prim":"unit"}`),
				Code:       []byte(`{"prim":"unit"}`),
				ReturnType: []byte(`{"prim":"unit"}`),
			},
			storageType: []byte(`{"prim":"unit"}`),
			want:        `[{"prim":"parameter","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"unit"}]}]},{"prim":"storage","args":[{"prim":"option","args":[{"prim":"unit"}]}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"unit"},{"prim":"SOME"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		}, {
			name: "test 2: parameter is nil",
			fields: fields{
				Code:       []byte(`{"prim":"unit"}`),
				ReturnType: []byte(`{"prim":"unit"}`),
			},
			storageType: []byte(`{"prim":"unit"}`),
			want:        `[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[{"prim":"option","args":[{"prim":"unit"}]}]},{"prim":"code","args":[[{"prim":"CAR"},{"prim":"unit"},{"prim":"SOME"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msv := &MichelsonStorageView{
				Parameter:  tt.fields.Parameter,
				Code:       tt.fields.Code,
				ReturnType: tt.fields.ReturnType,
				Name:       tt.fields.Name,
			}
			got, err := msv.GetCode(tt.storageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("MichelsonStorageView.GetCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, string(got), tt.want) {
				t.Errorf("Invalid result")
				return
			}
		})
	}
}
