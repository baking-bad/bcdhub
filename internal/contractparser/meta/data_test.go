package meta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
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
			want: `{"0":{"prim":"or","args":["0/0","0/1"],"type":"namedunion"},"0/0":{"fieldname":"run","prim":"pair","args":["0/0/0","0/0/1"],"type":"tuple","name":"run"},"0/0/0":{"prim":"string","type":"string"},"0/0/1":{"prim":"bytes","type":"bytes"},"0/1":{"fieldname":"upgrade","prim":"pair","args":["0/1/0/0","0/1/0/1","0/1/1/0","0/1/1/1/0/o","0/1/1/1/1/o"],"type":"namedtuple","name":"upgrade"},"0/1/0":{"prim":"pair","type":"pair"},"0/1/0/0":{"typename":"currentVersion","prim":"nat","type":"nat","name":"currentVersion"},"0/1/0/1":{"typename":"newVersion","prim":"nat","type":"nat","name":"newVersion"},"0/1/1":{"prim":"pair","type":"pair"},"0/1/1/0":{"typename":"migrationScript","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","return_value":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda","name":"migrationScript"},"0/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/0":{"typename":"newCode","prim":"option","type":"option"},"0/1/1/1/0/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","return_value":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda","name":"newCode"},"0/1/1/1/1":{"typename":"newPermCode","prim":"option","type":"option"},"0/1/1/1/1/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"unit\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","return_value":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda","name":"newPermCode"}}`,
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

func TestContractMetadata_IsUpgradable(t *testing.T) {
	testCases := []struct {
		address string
		result  bool
	}{
		{
			address: "KT1CyJxNgctn3gQKBu9ivKN5RSgqpmEhX5W8",
			result:  true,
		},
		{
			address: "KT1G9SQK1YK8oDTJAWaPjuBmY2fX5QGBnYLj",
			result:  true,
		},
		{
			address: "KT18bwMJoY3xj6vdB94mLyGGasyNZmSgZBuT",
			result:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.address, func(t *testing.T) {
			paramPath := fmt.Sprintf("./testdata/metadata/%s/parameter.json", tt.address)
			storagePath := fmt.Sprintf("./testdata/metadata/%s/storage.json", tt.address)

			paramFile, err := ioutil.ReadFile(paramPath)
			if err != nil {
				t.Errorf("ioutil.ReadFile %v error %v", paramPath, err)
				return
			}

			storageFile, err := ioutil.ReadFile(storagePath)
			if err != nil {
				t.Errorf("ioutil.ReadFile %v error %v", storagePath, err)
				return
			}

			symLink := "test"
			metadata := models.Metadata{
				Parameter: map[string]string{
					symLink: string(paramFile),
				},
				Storage: map[string]string{
					symLink: string(storageFile),
				},
			}

			contractMetadata, err := GetContractMetadataFromModel(metadata)
			if err != nil {
				t.Errorf("GetContractMetadataFromModel error %v", err)
				return
			}
			if contractMetadata.IsUpgradable(symLink) != tt.result {
				t.Errorf("invalid result %v", tt.address)
			}
		})
	}
}
