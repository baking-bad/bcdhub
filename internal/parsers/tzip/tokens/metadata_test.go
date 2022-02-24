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
				TokenID:     0,
				Decimals:    getIntPtr(0),
				Name:        "Cat Drawing",
				ArtifactURI: "https://cloudflare-ipfs.com/ipfs/QmS9V4PKSjQh8hzyQzRqKFxkCcSY1yLuXQYKx7YoTyJYYe",
				DisplayURI:  "https://cloudflare-ipfs.com/ipfs/QmS9V4PKSjQh8hzyQzRqKFxkCcSY1yLuXQYKx7YoTyJYYe",
				Extras: map[string]interface{}{
					"booleanAmount": "true",
				},
			},
		}, {
			name: "test 8: KT1XRxmUFNcbzGTwQPvNPa5FuuM43uEunp8K",
			value: `{
					"prim": "Pair",
					"args": [
					  {
						"int": "0"
					  },
					  [
						{
						  "prim": "Elt",
						  "args": [
							{
							  "string": "decimals"
							},
							{
							  "bytes": "30"
							}
						  ]
						},
						{
						  "prim": "Elt",
						  "args": [
							{
							  "string": "description"
							},
							{
							  "bytes": "54686973206973207468652054657a6f73205370616e69736820636f6d6d756e69747920746f6b656e2e0a0a4573746520657320656c20746f6b656e206465206c6120636f6d756e696461642065737061f16f6c612054657a6f73205370616e6973682e"
							}
						  ]
						},
						{
						  "prim": "Elt",
						  "args": [
							{
							  "string": "name"
							},
							{
							  "bytes": "54657a6f73205370616e697368"
							}
						  ]
						},
						{
						  "prim": "Elt",
						  "args": [
							{
							  "string": "symbol"
							},
							{
							  "bytes": "545a53"
							}
						  ]
						},
						{
						  "prim": "Elt",
						  "args": [
							{
							  "string": "thumbnailUri"
							},
							{
							  "bytes": "68747470733a2f2f6962622e636f2f7a514c62746851"
							}
						  ]
						}
					  ]
					]
				  }`,
			wantErr: false,
			want: &TokenMetadata{
				TokenID:      0,
				Decimals:     getIntPtr(0),
				Name:         "Tezos Spanish",
				Symbol:       "TZS",
				ThumbnailURI: "https://ibb.co/zQLbthQ",
				Extras:       map[string]interface{}{},
			},
		}, {
			name:    "test 9",
			value:   `{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":"decimal"},{"bytes":"06"}]},{"prim":"Elt","args":[{"string":"icon"},{"bytes":"05010000005a68747470733a2f2f696d616765732e6c61646570656368652e66722f6170692f76312f696d616765732f766965772f3563326539343365336534353436313134363339393066312f6f726967696e616c2f696d6167652e6a7067"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"05010000000d54696e6f206c757620636f696e"}]},{"prim":"Elt","args":[{"string":"symbol"},{"bytes":"050100000005383d3d3d44"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				Symbol: "\x05\x01\\u0000\\u0000\\u0000\x058===D",
				Name:   "\x05\x01\\u0000\\u0000\\u0000\rTino luv coin",
				Extras: map[string]interface{}{
					"decimal": "06",
					"icon":    "{ \"https://images.ladepeche.fr/api/v1/images/view/5c2e943e3e454611463990f1/original/image.jpg\" }",
				},
			},
		}, {
			name:    "test 10: KT1W6fJBgy2AtDZDoEGvSZvEseN7bPSdzNjc",
			value:   `{"prim":"Pair","args":[{"int":"0"},[{"prim":"Elt","args":[{"string":"artifactUri"},{"bytes":"68747470733a2f2f74612e636f2f313833323637342e676c7466"}]},{"prim":"Elt","args":[{"string":"isBooleanAmount"},{"bytes":"74727565"}]},{"prim":"Elt","args":[{"string":"decimals"},{"bytes":"30"}]},{"prim":"Elt","args":[{"string":"displayUri"},{"bytes":"68747470733a2f2f74612e636f2f313833323637342e737667"}]},{"prim":"Elt","args":[{"string":"name"},{"bytes":"4e4654205465737420546f6b656e"}]}]]}`,
			wantErr: false,
			want: &TokenMetadata{
				TokenID:         0,
				Decimals:        getIntPtr(0),
				Name:            "NFT Test Token",
				ArtifactURI:     "https://ta.co/1832674.gltf",
				DisplayURI:      "https://ta.co/1832674.svg",
				IsBooleanAmount: true,
				Extras:          map[string]interface{}{},
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
				Symbol:         "SIMMAW",
				Name:           "Mystery Map Award",
				Decimals:       int64Ptr(0),
				Description:    "A most mysterious map has been discovered. Where it leads is uncertain, but an adventure lies ahead.",
				DisplayURI:     "https://gateway.pinata.cloud/ipfs/QmPkJBaRnb2JwqA1S2sUQayTV9xT3x8MBnsmq7ForBWKuU",
				Creators:       []string{"test", "author"},
				IsTransferable: true,
				Extras: map[string]interface{}{
					"nonTransferable":     false,
					"symbolPreference":    false,
					"booleanAmount":       false,
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
						"actionLabel": "Send",
						"creators": ["test", "author"]
						}`),
		}, {
			name: "test ipfs 2",
			tm: TokenMetadata{
				Symbol:         "TZBKAB",
				Name:           "Klassare Alpha Brain",
				Decimals:       int64Ptr(0),
				Description:    "An upgraded unit, the great Klassare reborn.",
				IsTransferable: true,
				Extras: map[string]interface{}{
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
		}, {
			name: "test ipfs 3",
			tm: TokenMetadata{
				Symbol:         "OBJKT",
				Name:           "$XTZ all the way up! #Tezos",
				Decimals:       int64Ptr(0),
				Description:    "Tezos on Kilimanjaro",
				Tags:           []string{},
				ArtifactURI:    "ipfs://QmV9j9cYtXeB3fyutbTrdycaPHkXigJeEzBXbCKkTVR5Ah",
				ThumbnailURI:   "ipfs://QmNrhZHUaEqxhyLfqoq1mtHSipkWHeT31LNHb1QEbDHgnc",
				IsTransferable: true,
				Creators:       []string{"tz1QpaWdNjarzfDfjDVacXUeadF9kxBchzEQ"},
				Formats:        []byte(`[{"mimeType":"image/jpeg","uri":"ipfs://QmV9j9cYtXeB3fyutbTrdycaPHkXigJeEzBXbCKkTVR5Ah"}]`),
				Extras:         map[string]interface{}{},
			},
			data: []byte(`{"name":"$XTZ all the way up! #Tezos","description":"Tezos on Kilimanjaro","tags":[],"symbol":"OBJKT","artifactUri":"ipfs://QmV9j9cYtXeB3fyutbTrdycaPHkXigJeEzBXbCKkTVR5Ah","creators":["tz1QpaWdNjarzfDfjDVacXUeadF9kxBchzEQ"],"formats":[{"uri":"ipfs://QmV9j9cYtXeB3fyutbTrdycaPHkXigJeEzBXbCKkTVR5Ah","mimeType":"image/jpeg"}],"thumbnailUri":"ipfs://QmNrhZHUaEqxhyLfqoq1mtHSipkWHeT31LNHb1QEbDHgnc","decimals":0,"isBooleanAmount":false,"shouldPreferSymbol":false}`),
		}, {
			name: "test ipfs 4",
			tm: TokenMetadata{
				Symbol:         "OBJKT",
				Name:           "8-5",
				Decimals:       int64Ptr(0),
				Description:    "dat\\u0000\\u0000\\u0000\\u0000\\u0000™dy!\fTm¸4f\x1a\bB\x05y´∫÷\x1aæj\x10A!%ÕJV33cñâ¿ F*)R‰\x02IV\x0f\x13\x02\x1cf@ÁìMX\x12W/€*dZÜ$\x01‘!¥\x03ê¡(&\v≡///╱━━━━━------_____🔫",
				Tags:           []string{"Glitch", "gun", "lines", "🔫", "9983", "2021"},
				ArtifactURI:    "ipfs://QmQ1KzkhbkPvnRrWczXoekG64bCZ5T2NT7dwuFS6qVCi69",
				ThumbnailURI:   "ipfs://QmNrhZHUaEqxhyLfqoq1mtHSipkWHeT31LNHb1QEbDHgnc",
				IsTransferable: true,
				Creators:       []string{"tz1dAVKwbGe1PVnPBkRZYmYFsecDtTLHjHLK"},
				Formats:        []byte(`[{"mimeType":"video/mp4","uri":"ipfs://QmQ1KzkhbkPvnRrWczXoekG64bCZ5T2NT7dwuFS6qVCi69"}]`),
				Extras:         map[string]interface{}{},
			},
			data: []byte(`{"name":"8-5","description":"dat\u0000\u0000\u0000\u0000\u0000™dy!\fTm¸4f\u001a\bB\u0005y´∫÷\u001aæj\u0010A!%ÕJV33cñâ¿ F*)R‰\u0002IV\u000f\u0013\u0002\u001cf@ÁìMX\u0012W/€*dZÜ$\u0001‘!¥\u0003ê¡(&\u000b≡///╱━━━━━------_____🔫","tags":["Glitch","gun","lines","🔫","9983","2021"],"symbol":"OBJKT","artifactUri":"ipfs://QmQ1KzkhbkPvnRrWczXoekG64bCZ5T2NT7dwuFS6qVCi69","displayUri":"","thumbnailUri":"ipfs://QmNrhZHUaEqxhyLfqoq1mtHSipkWHeT31LNHb1QEbDHgnc","creators":["tz1dAVKwbGe1PVnPBkRZYmYFsecDtTLHjHLK"],"formats":[{"uri":"ipfs://QmQ1KzkhbkPvnRrWczXoekG64bCZ5T2NT7dwuFS6qVCi69","mimeType":"video/mp4"}],"decimals":0,"isBooleanAmount":false,"shouldPreferSymbol":false}`),
		}, {
			name: "test ipfs 5",
			data: []byte(`{"name":"ASCII Squigl #6","iterationHash":"opSHCwaLLUsTV2r15pwY6Pijr1NftAh7LdqCLYMMjkAw1pjx2hy","description":"fx({asciisquigl:'Squigls flowing through the Tezos Blockchain.'})\n\nASCII Squigls have 9 attributes and they are drawn to a 1000 x 1000 px canvas.\n\nSave image with 'S'\nStop animation with mouseclick","tags":["squiggle","squigl","hicetsquigl","ascii"],"generatorUri":"ipfs://QmXjGCuZPLH1W8YAupKqobNknhihDY38k5d8L4h7bM3Wxy","artifactUri":"ipfs://QmXjGCuZPLH1W8YAupKqobNknhihDY38k5d8L4h7bM3Wxy?fxhash=opSHCwaLLUsTV2r15pwY6Pijr1NftAh7LdqCLYMMjkAw1pjx2hy","displayUri":"ipfs://QmZD2gxEfK2JXGVoJqHurUxU9CgmLqKW6BCPtF5Sidp4mE","thumbnailUri":"ipfs://Qmehj9m1dnx5AqdbNEu7tRexjDhHYDGcF58NU4NcvCUqxn","authenticityHash":"b5784430923afe0ea3a124f3f0c6080d1585e11d83bd8948a82aacaf061f5df0","attributes":[{"name":"step","value":20},{"name":"theme","value":"Grayscale"},{"name":"color","value":"Grayscale"},{"name":"squigl_chars","value":23},{"name":"squigl_string","value":"£µ_\b9\u0019\u0003\u0000#88u;Dg»E+7*¥»"},{"name":"background_chars","value":12},{"name":"background_string","value":"f¥>\u0002\u0011cE3@["},{"name":"animation","value":"None"},{"name":"animation_speed","value":"DISABLED"}],"decimals":0,"symbol":"GENTK","version":"0.2"}`),
			tm: TokenMetadata{
				TokenID:         0,
				Decimals:        getIntPtr(0),
				Name:            "ASCII Squigl #6",
				Symbol:          "GENTK",
				Tags:            []string{"squiggle", "squigl", "hicetsquigl", "ascii"},
				Description:     "fx({asciisquigl:'Squigls flowing through the Tezos Blockchain.'})\n\nASCII Squigls have 9 attributes and they are drawn to a 1000 x 1000 px canvas.\n\nSave image with 'S'\nStop animation with mouseclick",
				IsBooleanAmount: false,
				IsTransferable:  true,
				ArtifactURI:     "ipfs://QmXjGCuZPLH1W8YAupKqobNknhihDY38k5d8L4h7bM3Wxy?fxhash=opSHCwaLLUsTV2r15pwY6Pijr1NftAh7LdqCLYMMjkAw1pjx2hy",
				DisplayURI:      "ipfs://QmZD2gxEfK2JXGVoJqHurUxU9CgmLqKW6BCPtF5Sidp4mE",
				ThumbnailURI:    "ipfs://Qmehj9m1dnx5AqdbNEu7tRexjDhHYDGcF58NU4NcvCUqxn",
				Extras: map[string]interface{}{
					"authenticityHash": "b5784430923afe0ea3a124f3f0c6080d1585e11d83bd8948a82aacaf061f5df0",
					"generatorUri":     "ipfs://QmXjGCuZPLH1W8YAupKqobNknhihDY38k5d8L4h7bM3Wxy",
					"iterationHash":    "opSHCwaLLUsTV2r15pwY6Pijr1NftAh7LdqCLYMMjkAw1pjx2hy",
					"version":          "0.2",
					"attributes": []interface{}{
						map[string]interface{}{
							"name":  "step",
							"value": 20.,
						},
						map[string]interface{}{
							"name":  "theme",
							"value": "Grayscale",
						},
						map[string]interface{}{
							"name":  "color",
							"value": "Grayscale",
						},
						map[string]interface{}{
							"name":  "squigl_chars",
							"value": 23.,
						},
						map[string]interface{}{
							"name":  "squigl_string",
							"value": "£µ_\b9\x19\x03\x00#88u;Dg»E+7*¥»",
						},
						map[string]interface{}{
							"name":  "background_chars",
							"value": 12.,
						},
						map[string]interface{}{
							"name":  "background_string",
							"value": "f¥\u0080>\x02\x11cE3@[\u009a",
						},
						map[string]interface{}{
							"name":  "animation",
							"value": "None",
						},
						map[string]interface{}{
							"name":  "animation_speed",
							"value": "DISABLED",
						},
					},
				},
			},
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
