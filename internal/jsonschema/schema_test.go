package jsonschema

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		binPath  string
		metadata string
		want     Schema
		wantErr  bool
	}{
		{
			name:     "Case 1: string field",
			binPath:  `0`,
			metadata: `{"0":{"prim":"string","type":"string"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "string",
						"prim":    "string",
						"title":   "string",
						"default": "",
					},
				},
			},
		}, {
			name:     "Case 2: integer field",
			binPath:  `0`,
			metadata: `{"0":{"prim":"int","type":"int"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "integer",
						"prim":    "int",
						"title":   "int",
						"default": 0,
					},
				},
			},
		}, {
			name:     "Case 3: integer field (nat)",
			binPath:  `0`,
			metadata: `{"0":{"prim":"nat","type":"nat"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "integer",
						"prim":    "nat",
						"title":   "nat",
						"default": 0,
					},
				},
			},
		}, {
			name:     "Case 4: pair fields",
			binPath:  `0`,
			metadata: `{"0":{"prim":"pair","args":["0/0","0/1"],"type":"tuple"},"0/0":{"prim":"string","type":"string"},"0/1":{"prim":"nat","type":"nat"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0/1": Schema{
						"type":    "integer",
						"prim":    "nat",
						"title":   "nat",
						"default": 0,
					},
					"0/0": Schema{
						"type":    "string",
						"prim":    "string",
						"title":   "string",
						"default": "",
					},
				},
			},
		}, {
			name:     "Case 5: string field (key_hash)",
			binPath:  `0`,
			metadata: `{"0":{"prim":"key_hash","type":"key_hash"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "string",
						"prim":    "key_hash",
						"title":   "key_hash",
						"default": "",
					},
				},
			},
		}, {
			name:     "Case 6: unit",
			binPath:  `0`,
			metadata: `{"0":{"prim":"unit","type":"unit"}}`,
			want:     nil,
		}, {
			name:     "Case 7: boolean field",
			binPath:  `0`,
			metadata: `{"0":{"prim":"bool","type":"bool"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "boolean",
						"prim":    "bool",
						"title":   "bool",
						"default": false,
					},
				},
			},
		}, {
			name:     "Case 8: map field",
			binPath:  `0/1`,
			metadata: `{"0":{"typename":"_entries","prim":"or","args":["0/0","0/1"],"type":"namedunion"},"0/0":{"fieldname":"main","prim":"unit","type":"unit","name":"main"},"0/1":{"fieldname":"debit","prim":"map","type":"map","name":"debit"},"0/1/k":{"prim":"address","type":"address"},"0/1/v":{"prim":"nat","type":"nat"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0/1": Schema{
						"type":        "array",
						"title":       "debit",
						"x-itemTitle": "0/1/k",
						"items": Schema{
							"type":     "object",
							"required": []string{"0/1/k", "0/1/v"},
							"properties": Schema{
								"0/1/k": Schema{
									"type":      "string",
									"prim":      "address",
									"title":     "address",
									"minLength": 36,
									"maxLength": 36,
									"default":   "",
								},
								"0/1/v": Schema{
									"type":    "integer",
									"prim":    "nat",
									"title":   "nat",
									"default": 0,
								},
							},
						},
					},
				},
			},
		}, {
			name:     "Case 9: list field",
			binPath:  "0",
			metadata: `{"0":{"prim":"list","type":"list"},"0/l":{"prim":"int","type":"int"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "array",
						"title":   "list",
						"default": []interface{}{},
						"items": Schema{
							"type":     "object",
							"required": []string{"0/l"},
							"properties": Schema{
								"0/l": Schema{
									"type":    "integer",
									"prim":    "int",
									"title":   "int",
									"default": 0,
								},
							},
						},
					},
				},
			},
		}, {
			name:     "Case 10: timestamp field",
			binPath:  "0",
			metadata: `{"0":{"fieldname":"refund_time","prim":"timestamp","type":"timestamp","name":"refund_time"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0": Schema{
						"type":    "string",
						"prim":    "timestamp",
						"title":   "refund_time",
						"format":  "date-time",
						"default": "",
					},
				},
			},
		}, {
			name:     "Case 11: or field",
			binPath:  "0/1",
			metadata: `{"0/1": {"fieldname": "withdraw","prim": "or","type": "or", "args":["0/1/0", "0/1/1"]},"0/1/0": {"typename": "secret","fieldname": "redeem","prim": "bytes","type": "bytes","name": "redeem"},"0/1/1": {"typename": "hashed_secret", "fieldname": "refund", "prim": "bytes","type": "bytes","name": "refund"}}`,
			want: Schema{
				"type":  "object",
				"title": "withdraw",
				"prim":  "or",
				"oneOf": []Schema{
					{
						"title": "redeem",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "0/1/0",
							},
							"0/1/0": Schema{
								"type":    "string",
								"prim":    "bytes",
								"title":   "redeem",
								"default": "",
							},
						},
					},
					{
						"title": "refund",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "0/1/1",
							},
							"0/1/1": Schema{
								"type":    "string",
								"prim":    "bytes",
								"title":   "refund",
								"default": "",
							},
						},
					},
				},
			},
		}, {
			name:     "Case 12: option field",
			binPath:  "0/1/o",
			metadata: `{"0/1/o":{"fieldname":"Pour","prim":"pair","args":["0/1/o/0","0/1/o/1"],"type":"namedtuple","name":"Pour"},"0/1/o/0":{"fieldname":"pour_auth","prim":"signature","type":"signature","name":"pour_auth"},"0/1/o/1":{"fieldname":"pour_amount","prim":"mutez","type":"mutez","name":"pour_amount"}}`,
			want: Schema{
				"type":  "object",
				"prim":  "option",
				"title": "Pour",
				"default": Schema{
					"schemaKey": Schema{
						"type":  "string",
						"const": "none",
					},
				},
				"oneOf": []Schema{
					{
						"title": "None",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "none",
							},
						},
					},
					{
						"title": "Some",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "some",
							},
							"0/1/o/0": Schema{
								"type":    "string",
								"title":   "pour_auth",
								"prim":    "signature",
								"default": "",
							},
							"0/1/o/1": Schema{
								"type":    "integer",
								"title":   "pour_amount",
								"prim":    "mutez",
								"default": 0,
							},
						},
					},
				},
			},
		}, {
			name:     "Case 13: tzBTC upgrade",
			binPath:  "0/1/1/1/1/0/0/0/1",
			metadata: `{"0":{"prim":"or","args":["0/0/0/0","0/0/0/1","0/0/1/0","0/0/1/1/0","0/0/1/1/1","0/1/0/0","0/1/0/1","0/1/1/0","0/1/1/1/0","0/1/1/1/1/0/0/0/0","0/1/1/1/1/0/0/0/1","0/1/1/1/1/0/0/1/0","0/1/1/1/1/0/0/1/1","0/1/1/1/1/0/1/0/0","0/1/1/1/1/0/1/0/1","0/1/1/1/1/0/1/1/0","0/1/1/1/1/0/1/1/1","0/1/1/1/1/1/0/0/0","0/1/1/1/1/1/0/0/1","0/1/1/1/1/1/0/1/0","0/1/1/1/1/1/0/1/1","0/1/1/1/1/1/1/0/0","0/1/1/1/1/1/1/0/1","0/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/0","0/1/1/1/1/1/1/1/1/1"],"type":"namedunion"},"0/0":{"prim":"or","type":"or"},"0/0/0":{"prim":"or","type":"or"},"0/0/0/0":{"fieldname":"getVersion","prim":"pair","args":["0/0/0/0/0","0/0/0/0/1"],"type":"tuple","name":"getVersion"},"0/0/0/0/0":{"prim":"unit","type":"unit"},"0/0/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/0/1":{"fieldname":"getAllowance","prim":"pair","args":["0/0/0/1/0/0","0/0/0/1/0/1","0/0/0/1/1"],"type":"namedtuple","name":"getAllowance"},"0/0/0/1/0":{"prim":"pair","type":"pair"},"0/0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/0/1/0/1":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/0/0/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1":{"prim":"or","type":"or"},"0/0/1/0":{"fieldname":"getBalance","prim":"pair","args":["0/0/1/0/0","0/0/1/0/1"],"type":"namedtuple","name":"getBalance"},"0/0/1/0/0":{"typename":"owner","prim":"address","type":"address","name":"owner"},"0/0/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1":{"prim":"or","type":"or"},"0/0/1/1/0":{"fieldname":"getTotalSupply","prim":"pair","args":["0/0/1/1/0/0","0/0/1/1/0/1"],"type":"tuple","name":"getTotalSupply"},"0/0/1/1/0/0":{"prim":"unit","type":"unit"},"0/0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/0/1/1/1":{"fieldname":"getTotalMinted","prim":"pair","args":["0/0/1/1/1/0","0/0/1/1/1/1"],"type":"tuple","name":"getTotalMinted"},"0/0/1/1/1/0":{"prim":"unit","type":"unit"},"0/0/1/1/1/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1":{"prim":"or","type":"or"},"0/1/0":{"prim":"or","type":"or"},"0/1/0/0":{"fieldname":"getTotalBurned","prim":"pair","args":["0/1/0/0/0","0/1/0/0/1"],"type":"tuple","name":"getTotalBurned"},"0/1/0/0/0":{"prim":"unit","type":"unit"},"0/1/0/0/1":{"prim":"contract","parameter":"{\"prim\":\"nat\"}","type":"contract"},"0/1/0/1":{"fieldname":"getOwner","prim":"pair","args":["0/1/0/1/0","0/1/0/1/1"],"type":"tuple","name":"getOwner"},"0/1/0/1/0":{"prim":"unit","type":"unit"},"0/1/0/1/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/1":{"prim":"or","type":"or"},"0/1/1/0":{"fieldname":"getRedeemAddress","prim":"pair","args":["0/1/1/0/0","0/1/1/0/1"],"type":"tuple","name":"getRedeemAddress"},"0/1/1/0/0":{"prim":"unit","type":"unit"},"0/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"address\"}","type":"contract"},"0/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/0":{"fieldname":"getTokenMetadata","prim":"pair","args":["0/1/1/1/0/0","0/1/1/1/0/1"],"type":"tuple","name":"getTokenMetadata"},"0/1/1/1/0/0":{"prim":"list","type":"list"},"0/1/1/1/0/0/l":{"prim":"nat","type":"nat"},"0/1/1/1/0/1":{"prim":"contract","parameter":"{\"prim\":\"list\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"nat\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"pair\",\"args\":[{\"prim\":\"nat\"},{\"prim\":\"map\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"string\"}]}]}]}]}]}]}","type":"contract"},"0/1/1/1/1":{"fieldname":"safeEntrypoints","prim":"or","type":"or"},"0/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/0/0":{"fieldname":"run","prim":"pair","args":["0/1/1/1/1/0/0/0/0/0","0/1/1/1/1/0/0/0/0/1"],"type":"tuple","name":"run"},"0/1/1/1/1/0/0/0/0/0":{"prim":"string","type":"string"},"0/1/1/1/1/0/0/0/0/1":{"prim":"bytes","type":"bytes"},"0/1/1/1/1/0/0/0/1":{"fieldname":"upgrade","prim":"pair","args":["0/1/1/1/1/0/0/0/1/0/0","0/1/1/1/1/0/0/0/1/0/1","0/1/1/1/1/0/0/0/1/1/0","0/1/1/1/1/0/0/0/1/1/1/0/o","0/1/1/1/1/0/0/0/1/1/1/1/o"],"type":"tuple","name":"upgrade"},"0/1/1/1/1/0/0/0/1/0":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/0/0":{"typename":"currentVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/0/1":{"typename":"newVersion","prim":"nat","type":"nat"},"0/1/1/1/1/0/0/0/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/0":{"typename":"migrationScript","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/0/0/1/1/1/0":{"typename":"newCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/0/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/0/1/1/1/1":{"typename":"newPermCode","prim":"option","type":"option"},"0/1/1/1/1/0/0/0/1/1/1/1/o":{"prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"unit\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda"},"0/1/1/1/1/0/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/0/1/0":{"fieldname":"epwBeginUpgrade","prim":"pair","args":["0/1/1/1/1/0/0/1/0/0","0/1/1/1/1/0/0/1/0/1"],"type":"namedtuple","name":"epwBeginUpgrade"},"0/1/1/1/1/0/0/1/0/0":{"typename":"current","prim":"nat","type":"nat","name":"current"},"0/1/1/1/1/0/0/1/0/1":{"typename":"new","prim":"nat","type":"nat","name":"new"},"0/1/1/1/1/0/0/1/1":{"typename":"migrationscript","fieldname":"epwApplyMigration","prim":"lambda","parameter":"{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}","type":"lambda","name":"epwApplyMigration"},"0/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/0/0":{"typename":"contractcode","fieldname":"epwSetCode","prim":"lambda","parameter":"{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}","type":"lambda","name":"epwSetCode"},"0/1/1/1/1/0/1/0/1":{"fieldname":"epwFinishUpgrade","prim":"unit","type":"unit","name":"epwFinishUpgrade"},"0/1/1/1/1/0/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/0/1/1/0":{"fieldname":"transfer","prim":"pair","args":["0/1/1/1/1/0/1/1/0/0","0/1/1/1/1/0/1/1/0/1/0","0/1/1/1/1/0/1/1/0/1/1"],"type":"namedtuple","name":"transfer"},"0/1/1/1/1/0/1/1/0/0":{"typename":"from","prim":"address","type":"address","name":"from"},"0/1/1/1/1/0/1/1/0/1":{"prim":"pair","type":"pair"},"0/1/1/1/1/0/1/1/0/1/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/0/1/1/0/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/0/1/1/1":{"fieldname":"approve","prim":"pair","args":["0/1/1/1/1/0/1/1/1/0","0/1/1/1/1/0/1/1/1/1"],"type":"namedtuple","name":"approve"},"0/1/1/1/1/0/1/1/1/0":{"typename":"spender","prim":"address","type":"address","name":"spender"},"0/1/1/1/1/0/1/1/1/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/0/0":{"fieldname":"mint","prim":"pair","args":["0/1/1/1/1/1/0/0/0/0","0/1/1/1/1/1/0/0/0/1"],"type":"namedtuple","name":"mint"},"0/1/1/1/1/1/0/0/0/0":{"typename":"to","prim":"address","type":"address","name":"to"},"0/1/1/1/1/1/0/0/0/1":{"typename":"value","prim":"nat","type":"nat","name":"value"},"0/1/1/1/1/1/0/0/1":{"typename":"value","fieldname":"burn","prim":"nat","type":"nat","name":"burn"},"0/1/1/1/1/1/0/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/0/1/0":{"typename":"operator","fieldname":"addOperator","prim":"address","type":"address","name":"addOperator"},"0/1/1/1/1/1/0/1/1":{"typename":"operator","fieldname":"removeOperator","prim":"address","type":"address","name":"removeOperator"},"0/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/0/0":{"typename":"redeem","fieldname":"setRedeemAddress","prim":"address","type":"address","name":"setRedeemAddress"},"0/1/1/1/1/1/1/0/1":{"fieldname":"pause","prim":"unit","type":"unit","name":"pause"},"0/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/0":{"fieldname":"unpause","prim":"unit","type":"unit","name":"unpause"},"0/1/1/1/1/1/1/1/1":{"prim":"or","type":"or"},"0/1/1/1/1/1/1/1/1/0":{"typename":"newOwner","fieldname":"transferOwnership","prim":"address","type":"address","name":"transferOwnership"},"0/1/1/1/1/1/1/1/1/1":{"fieldname":"acceptOwnership","prim":"unit","type":"unit","name":"acceptOwnership"}}`,
			want: Schema{
				"type": "object",
				"properties": Schema{
					"0/1/1/1/1/0/0/0/1/1/0": Schema{
						"prim":    "lambda",
						"type":    "string",
						"title":   "lambda",
						"default": "",
					},
					"0/1/1/1/1/0/0/0/1/0/0": Schema{
						"prim":    "nat",
						"type":    "integer",
						"title":   "nat",
						"default": 0,
					},
					"0/1/1/1/1/0/0/0/1/0/1": Schema{
						"prim":    "nat",
						"type":    "integer",
						"title":   "nat",
						"default": 0,
					},
					"0/1/1/1/1/0/0/0/1/1/1/0/o": Schema{
						"type":  "object",
						"prim":  "option",
						"title": "",
						"default": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "none",
							},
						},
						"oneOf": []Schema{
							{
								"title": "None",
								"properties": Schema{
									"schemaKey": Schema{
										"type":  "string",
										"const": "none",
									},
								},
							},
							{
								"title": "Some",
								"properties": Schema{
									"schemaKey": Schema{
										"type":  "string",
										"const": "some",
									},
									"0/1/1/1/1/0/0/0/1/1/1/0/o": Schema{
										"type":    "string",
										"prim":    "lambda",
										"title":   "lambda",
										"default": "",
									},
								},
							},
						},
					},
					"0/1/1/1/1/0/0/0/1/1/1/1/o": Schema{
						"type":  "object",
						"prim":  "option",
						"title": "",
						"default": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "none",
							},
						},
						"oneOf": []Schema{
							{
								"title": "None",
								"properties": Schema{
									"schemaKey": Schema{
										"type":  "string",
										"const": "none",
									},
								},
							},
							{
								"title": "Some",
								"properties": Schema{
									"schemaKey": Schema{
										"type":  "string",
										"const": "some",
									},
									"0/1/1/1/1/0/0/0/1/1/1/1/o": Schema{
										"type":    "string",
										"prim":    "lambda",
										"title":   "lambda",
										"default": "",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata meta.Metadata
			if err := json.Unmarshal([]byte(tt.metadata), &metadata); err != nil {
				t.Errorf("Create() %v", err)
				return
			}

			got, err := Create(tt.binPath, metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
