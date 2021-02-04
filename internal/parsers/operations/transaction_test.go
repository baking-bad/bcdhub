package operations

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_findByFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		data      string
		want      string
		wantErr   bool
	}{
		{
			name:      "default",
			fieldName: "default",
			data:      `{"prim": "unit"}`,
			want:      `{"prim": "unit"}`,
			wantErr:   true,
		}, {
			name:      "not found",
			fieldName: "test",
			data:      `{"prim": "unit"}`,
			want:      `{"prim": "unit"}`,
			wantErr:   true,
		}, {
			name:      "found main",
			fieldName: "main",
			data:      `{"prim":"or","args":[{"prim":"unit","annots":["%default"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}]}`,
			want:      `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}`,
		}, {
			name:      "found default",
			fieldName: "default",
			data:      `{"prim":"or","args":[{"prim":"unit","annots":["%default"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%operation"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}],"annots":[":action"]}],"annots":[":payload"]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%main"]}]}`,
			want:      `{"prim":"unit","annots":["%default"]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.data)
			got, err := findByFieldName(tt.fieldName, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("findByFieldName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := gjson.Parse(tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("findByFieldName() = %v, want %v", got, want)
			}
		})
	}
}
