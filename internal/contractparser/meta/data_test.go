package meta

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/tidwall/gjson"
)

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name    string
		v       string
		want    string
		wantErr bool
	}{
		{
			name: "Case: tzbtc upgrade",
			v:    `[{"prim": "or", "args": [{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%run"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":currentVersion"]},{"prim":"nat","annots":[":newVersion"]}]},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationScript"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newCode"]},{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newPermCode"]}]}]}],"annots":["%upgrade"]}]}]`,
			want: `{"0":{"prim":"or","type":"namedunion", "args":["0/0", "0/1"]},"0/0":{"fieldname":"run","prim":"pair","args":["0/0/0","0/0/1"],"type":"tuple","name":"run"},"0/0/0":{"prim":"string","type":"string"},"0/0/1":{"prim":"bytes","type":"bytes"},"0/1":{"fieldname":"upgrade","prim":"pair","args":["0/1/0/0","0/1/0/1","0/1/1/0","0/1/1/1/0/o","0/1/1/1/1/o"],"type":"namedtuple","name":"upgrade"},"0/1/0":{"prim":"pair","type":"pair"},"0/1/0/0":{"typename":"currentVersion", "name":"currentVersion","prim":"nat","type":"nat"},"0/1/0/1":{"typename":"newVersion","name":"newVersion","prim":"nat","type":"nat"},"0/1/1":{"prim":"pair","type":"pair"},"0/1/1/0":{"typename":"migrationScript","name":"migrationScript","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda"},"0/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/0":{"typename":"newCode","prim":"option","type":"option"},"0/1/1/1/0/o":{"name":"newCode","prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1":{"typename": "newPermCode","prim": "option","type": "option"},"0/1/1/1/1/o": {"name":"newPermCode","prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"unit\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := gjson.Parse(tt.v)
			var metadata Metadata
			if err := json.Unmarshal([]byte(tt.want), &metadata); err != nil {
				t.Errorf("ParseMetadata() error = %v", err)
				return
			}

			got, err := ParseMetadata(value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			first, err := json.MarshalIndent(got, "", " ")
			if err != nil {
				log.Println(err)
				t.Errorf("ParseMetadata() = %v, want %v", got, metadata)
				return
			}
			second, err := json.MarshalIndent(metadata, "", " ")
			if err != nil {
				log.Println(err)
				t.Errorf("ParseMetadata() = %v, want %v", got, metadata)
				return
			}
			if string(first) != string(second) {
				t.Errorf("ParseMetadata() = %v, want %v", string(first), string(second))
				return
			}
		})
	}
}
