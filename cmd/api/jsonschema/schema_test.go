package jsonschema

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

func TestCreate(t *testing.T) {
	type args struct {
	}
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
					"string_0": Schema{
						"type":  "string",
						"title": "string",
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
					"int_0": Schema{
						"type":  "integer",
						"title": "int",
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
					"nat_0": Schema{
						"type":  "integer",
						"title": "nat",
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
					"nat_01": Schema{
						"type":  "integer",
						"title": "nat",
					},
					"string_00": Schema{
						"type":  "string",
						"title": "string",
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
					"key_hash_0": Schema{
						"type":  "string",
						"title": "key_hash",
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
					"bool_0": Schema{
						"type":  "boolean",
						"title": "bool",
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
					"map_01": Schema{
						"type":        "array",
						"title":       "debit",
						"x-itemTitle": "address_01k",
						"items": Schema{
							"type":     "object",
							"required": []string{"address_01k", "nat_01v"},
							"properties": Schema{
								"address_01k": Schema{
									"type":  "string",
									"title": "address",
								},
								"nat_01v": Schema{
									"type":  "integer",
									"title": "nat",
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
					"list_0": Schema{
						"type":        "array",
						"title":       "list",
						"x-itemTitle": "int_0l",
						"items": Schema{
							"type":     "object",
							"required": []string{"int_0l"},
							"properties": Schema{
								"int_0l": Schema{
									"type":  "integer",
									"title": "int",
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
					"timestamp_0": Schema{
						"type":   "string",
						"title":  "refund_time",
						"format": "date-time",
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
				"oneOf": []Schema{
					Schema{
						"title": "Left",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "0/1/0",
							},
							"bytes_010": Schema{
								"type":  "string",
								"title": "redeem",
							},
						},
					},
					Schema{
						"title": "Right",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "0/1/1",
							},
							"bytes_011": Schema{
								"type":  "string",
								"title": "refund",
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
				"title": "Pour",
				"oneOf": []Schema{
					Schema{
						"title": "None",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "none",
							},
							"noneField": Schema{
								"type":     "string",
								"title":    "Option",
								"readOnly": true,
							},
						},
					},
					Schema{
						"title": "Some",
						"properties": Schema{
							"schemaKey": Schema{
								"type":  "string",
								"const": "0/1/o",
							},
							"signature_01o0": Schema{
								"type":  "string",
								"title": "pour_auth",
							},
							"mutez_01o1": Schema{
								"type":  "integer",
								"title": "pour_amount",
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
