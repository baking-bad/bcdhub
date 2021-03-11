package tokens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestTokenMetadata_Parse(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		want    *TokenMetadata
	}{
		{
			name:    "test 1",
			value:   `{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":""},{"bytes":"697066733a2f2f516d543633634b35584a6943645047436a586a526162634167646a787875714d4d67553679416d4c6e6178455a35"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				Link:    "ipfs://QmT63cK5XJiCdPGCjXjRabcAgdjxxuqMMgU6yAmLnaxEZ5",
				TokenID: 0,
				Extras: map[string]interface{}{
					"@@empty": "ipfs://QmT63cK5XJiCdPGCjXjRabcAgdjxxuqMMgU6yAmLnaxEZ5",
				},
			},
		}, {
			name:    "test 2",
			value:   `{"prim":"Pair","args":[{"int":"1"},[{"prim":"Elt","args":[{"string":"decimals"},{"bytes":"36"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"4e616d65"}]},{"prim":"Elt","args":[{"string":"symbol"},{"bytes":"534d42"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				TokenID:  1,
				Decimals: getIntPtr(6),
				Name:     "Name",
				Symbol:   "SMB",
				Extras:   make(map[string]interface{}),
			},
		}, {
			name:    "test 3",
			value:   `{"prim":"Pair","args":[{"int":"2"},[{"prim":"Elt","args":[{"string":""},{"bytes":"74657a6f732d73746f726167653a636f6e74656e74"}]},{"prim":"Elt","args":[{"string":"content"},{"bytes":"7b226e616d65223a20224e616d65222c202273796d626f6c223a2022534d42222c2022646563696d616c73223a20367d"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				TokenID: 2,
				Extras: map[string]interface{}{
					"@@empty": "tezos-storage:content",
					"content": "{\"name\": \"Name\", \"symbol\": \"SMB\", \"decimals\": 6}",
				},
				Link: "tezos-storage:content",
			},
		}, {
			name:    "test 4: invalid prim",
			value:   `{"prim":"list","args":[{"int":"2"},[{"prim":"Elt","args":[{"string":""},{"bytes":"74657a6f732d73746f726167653a636f6e74656e74"}]},{"prim":"Elt","args":[{"string":"content"},{"bytes":"7b226e616d65223a20224e616d65222c202273796d626f6c223a2022534d42222c2022646563696d616c73223a20367d"}]}]]}`,
			wantErr: true,
			want:    &TokenMetadata{},
		}, {
			name:    "test 5: invalid token ID",
			value:   `{"prim":"Pair","args":[{"string":"2"},[{"prim":"Elt","args":[{"string":""},{"bytes":"74657a6f732d73746f726167653a636f6e74656e74"}]},{"prim":"Elt","args":[{"string":"content"},{"bytes":"7b226e616d65223a20224e616d65222c202273796d626f6c223a2022534d42222c2022646563696d616c73223a20367d"}]}]]}`,
			wantErr: true,
			want:    &TokenMetadata{},
		}, {
			name:    "test 6: invalid metadata map",
			value:   `{"prim":"Pair","args":[{"int":"2"},{"prim":"Elt","args":[{"string":""},{"bytes":"74657a6f732d73746f726167653a636f6e74656e74"}]}]}`,
			wantErr: true,
			want:    &TokenMetadata{},
		}, {
			name:    "test 7: KT1WfbXNtvNDy7HLzPipn3x8CURTUgGYNSj9",
			value:   `{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":"artifactUri"},{"bytes":"68747470733a2f2f636c6f7564666c6172652d697066732e636f6d2f697066732f516d53395634504b536a516838687a79517a52714b46786b4363535931794c755851594b7837596f54794a595965"}]},{"prim":"Elt","args":[{"string":"booleanAmount"},{"bytes":"74727565"}]},{"prim":"Elt","args":[{"string":"decimals"},{"bytes":"30"}]},{"prim":"Elt","args":[{"string":"displayUri"},{"bytes":"68747470733a2f2f636c6f7564666c6172652d697066732e636f6d2f697066732f516d53395634504b536a516838687a79517a52714b46786b4363535931794c755851594b7837596f54794a595965"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"4361742044726177696e67"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				TokenID:  0,
				Decimals: getIntPtr(0),
				Name:     "Cat Drawing",
				Extras: map[string]interface{}{
					"artifactUri":   "https://cloudflare-ipfs.com/ipfs/QmS9V4PKSjQh8hzyQzRqKFxkCcSY1yLuXQYKx7YoTyJYYe",
					"booleanAmount": "true",
					"displayUri":    "https://cloudflare-ipfs.com/ipfs/QmS9V4PKSjQh8hzyQzRqKFxkCcSY1yLuXQYKx7YoTyJYYe",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := gjson.Parse(tt.value)

			m := &TokenMetadata{}
			if err := m.Parse(value, "", 0); (err != nil) != tt.wantErr {
				t.Errorf("TokenMetadataParser.parseMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, m, tt.want) {
				t.Errorf("assert")
			}
		})
	}
}

func getIntPtr(value int64) *int64 {
	return &value
}

func TestTokenMetadata_Merge(t *testing.T) {
	tests := []struct {
		name   string
		one    *TokenMetadata
		second *TokenMetadata
		want   *TokenMetadata
	}{
		{
			name: "test 1",
			one:  &TokenMetadata{},
			second: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  10,
			},
			want: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
			},
		}, {
			name:   "test 2",
			second: &TokenMetadata{},
			one: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  10,
			},
			want: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  10,
			},
		}, {
			name: "test 2",
			one: &TokenMetadata{
				Symbol:   "symbol old",
				Name:     "name old",
				Decimals: getIntPtr(9),
				TokenID:  11,
			},
			second: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  10,
			},
			want: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  11,
			},
		}, {
			name: "test 2",
			one: &TokenMetadata{
				Symbol:   "symbol old",
				Name:     "name old",
				Decimals: getIntPtr(9),
				TokenID:  11,
				Extras: map[string]interface{}{
					"test": "1234",
					"a":    "234",
				},
			},
			second: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  10,
				Extras: map[string]interface{}{
					"test": "12345",
					"b":    "234",
				},
			},
			want: &TokenMetadata{
				Symbol:   "symbol",
				Name:     "name",
				Decimals: getIntPtr(10),
				TokenID:  11,
				Extras: map[string]interface{}{
					"test": "12345",
					"a":    "234",
					"b":    "234",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.one.Merge(tt.second)
			if !assert.Equal(t, tt.one, tt.want) {
				t.Errorf("assert")
			}
		})
	}
}

func TestTokenMetadata_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		tm   TokenMetadata
		data []byte
	}{
		{
			name: "test ipfs",
			tm: TokenMetadata{
				Symbol:   "SIMMAW",
				Name:     "Mystery Map Award",
				Decimals: int64Ptr(0),
				Extras: map[string]interface{}{
					"description":         "A most mysterious map has been discovered. Where it leads is uncertain, but an adventure lies ahead.",
					"nonTransferable":     false,
					"symbolPreference":    false,
					"booleanAmount":       false,
					"displayUri":          "https://gateway.pinata.cloud/ipfs/QmPkJBaRnb2JwqA1S2sUQayTV9xT3x8MBnsmq7ForBWKuU",
					"defaultPresentation": "large",
					"actionLabel":         "Send",
				},
			},
			data: []byte(`{
				"name": "Mystery Map Award",
				"symbol": "SIMMAW",
				"decimals": 0,
				"description": "A most mysterious map has been discovered. Where it leads is uncertain, but an adventure lies ahead.",
				"nonTransferable": false,
				"symbolPreference": false,
				"booleanAmount": false,
				"displayUri": "https://gateway.pinata.cloud/ipfs/QmPkJBaRnb2JwqA1S2sUQayTV9xT3x8MBnsmq7ForBWKuU",
				"defaultPresentation": "large",
				"actionLabel": "Send"
				}`),
		}, {
			name: "test ipfs 2",
			tm: TokenMetadata{
				Symbol:   "TZBKAB",
				Name:     "Klassare Alpha Brain",
				Decimals: int64Ptr(0),
				Extras: map[string]interface{}{
					"description":         "An upgraded unit, the great Klassare reborn.",
					"isNft":               true,
					"nonTransferrable":    false,
					"symbolPrecedence":    false,
					"binaryAmount":        true,
					"imageUri":            "https://gateway.pinata.cloud/ipfs/QmZjeBZT5QykT4sEELYP2cYYEPTtgwx3vQhnyMzCmDKB7Q",
					"defaultPresentation": "small",
					"actionLabel":         "Send",
				},
			},
			data: []byte(`{
				"name": "Klassare Alpha Brain",
				"symbol": "TZBKAB",
				"decimals": "0",
				"description": "An upgraded unit, the great Klassare reborn.",
				"isNft": true,
				"nonTransferrable": false,
				"symbolPrecedence": false,
				"binaryAmount": true,
				"imageUri": "https://gateway.pinata.cloud/ipfs/QmZjeBZT5QykT4sEELYP2cYYEPTtgwx3vQhnyMzCmDKB7Q",
				"defaultPresentation": "small",
				"actionLabel": "Send"
				}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(TokenMetadata)
			if err := m.UnmarshalJSON(tt.data); err != nil {
				t.Errorf("TokenMetadata.UnmarshalJSON() error = %v", err)
				return
			}

			assert.Equal(t, tt.tm, *m)
		})
	}
}

func int64Ptr(val int64) *int64 {
	return &val
}
