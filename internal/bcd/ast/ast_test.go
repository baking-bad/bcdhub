package ast

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypedAst_ToJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *JSONSchema
		wantErr bool
	}{
		{
			name: "Case 1: string field",
			data: `{ "prim": "string" }`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@string_1": {
						Type:    "string",
						Prim:    "string",
						Title:   "@string_1",
						Default: "",
					},
				},
			},
		}, {
			name: "Case 2: integer field",
			data: `{ "prim": "int" }`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@int_1": {
						Type:    JSONSchemaTypeInt,
						Prim:    "int",
						Title:   "@int_1",
						Default: 0,
					},
				},
			},
		}, {
			name: "Case 3: integer field (nat)",
			data: `{ "prim": "nat" }`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@nat_1": {
						Type:    JSONSchemaTypeInt,
						Prim:    "nat",
						Title:   "@nat_1",
						Default: 0,
					},
				},
			},
		}, {
			name: "Case 4: pair fields",
			data: `{"prim":"pair", "args":[{"prim": "string"}, {"prim":"nat"}]}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@nat_3": {
						Type:    JSONSchemaTypeInt,
						Prim:    "nat",
						Title:   "@nat_3",
						Default: 0,
					},
					"@string_2": {
						Type:    "string",
						Prim:    "string",
						Title:   "@string_2",
						Default: "",
					},
				},
			},
		}, {
			name: "Case 5: string field (key_hash)",
			data: `{"prim":"key_hash"}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@key_hash_1": {
						Type:    JSONSchemaTypeString,
						Prim:    "key_hash",
						Title:   "@key_hash_1",
						Default: "",
					},
				},
			},
		}, {
			name: "Case 6: unit",
			data: `{"prim":"unit"}`,
			want: nil,
		}, {
			name: "Case 7: boolean field",
			data: `{"prim":"bool"}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@bool_1": {
						Type:    JSONSchemaTypeBool,
						Prim:    "bool",
						Title:   "@bool_1",
						Default: false,
					},
				},
			},
		}, {
			name: "Case 8: map field",
			data: `{"prim": "map", "annots": [":debit"], "args": [{"prim":"address"}, {"prim":"nat"}]}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"debit": {
						Type:       JSONSchemaTypeArray,
						Title:      "debit",
						XItemTitle: "@address_2",
						Default:    make([]interface{}, 0),
						Items: &SchemaKey{
							Type:     JSONSchemaTypeObject,
							Required: []string{"@address_2", "@nat_3"},
							Properties: map[string]*JSONSchema{
								"@address_2": {
									Type:      JSONSchemaTypeString,
									Prim:      "address",
									Title:     "@address_2",
									MinLength: 36,
									MaxLength: 36,
									Default:   "",
								},
								"@nat_3": {
									Type:    JSONSchemaTypeInt,
									Prim:    "nat",
									Title:   "@nat_3",
									Default: 0,
								},
							},
						},
					},
				},
			},
		}, {
			name: "Case 9: list field",
			data: `{"prim":"list", "args":[{"prim":"int"}]}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@list_1": {
						Type:    JSONSchemaTypeArray,
						Prim:    "list",
						Title:   "@list_1",
						Default: []interface{}{},
						Items: &SchemaKey{
							Type:     JSONSchemaTypeObject,
							Required: []string{"@int_2"},
							Properties: map[string]*JSONSchema{
								"@int_2": {
									Type:    JSONSchemaTypeInt,
									Prim:    "int",
									Title:   "@int_2",
									Default: 0,
								},
							},
						},
					},
				},
			},
		}, {
			name: "Case 10: timestamp field",
			data: `{"annots":[":refund_time"],"prim":"timestamp"}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"refund_time": {
						Type:    JSONSchemaTypeString,
						Prim:    "timestamp",
						Title:   "refund_time",
						Format:  "date-time",
						Default: time.Now().UTC().Format(time.RFC3339),
					},
				},
			},
		}, {
			name: "Case 11: or field",
			data: `{"prim": "or","args": [{"prim": "bytes","annots": [":secret","%redeem"]},{"prim": "bytes","annots": [":hashed_secret","%refund"]}],"annots": ["%withdraw"]}`,
			want: &JSONSchema{
				Type:  JSONSchemaTypeObject,
				Title: "withdraw",
				Prim:  "or",
				OneOf: []*JSONSchema{
					{
						Title: "redeem",
						Properties: map[string]*JSONSchema{
							"schemaKey": {
								Type:  JSONSchemaTypeString,
								Const: "redeem",
							},
							"redeem": {
								Type:    JSONSchemaTypeString,
								Prim:    "bytes",
								Title:   "redeem",
								Default: "",
							},
						},
					},
					{
						Title: "refund",
						Properties: map[string]*JSONSchema{
							"schemaKey": {
								Type:  JSONSchemaTypeString,
								Const: "refund",
							},
							"refund": {
								Type:    JSONSchemaTypeString,
								Prim:    "bytes",
								Title:   "refund",
								Default: "",
							},
						},
					},
				},
			},
		}, {
			name: "Case 12: option field",
			data: `{"prim": "option","args": [{"prim": "pair","args": [{"prim": "signature","annots": ["%pour_auth"]},{"prim": "mutez","annots": ["%pour_amount"]}],"annots": ["%Pour"]}]}`,
			want: &JSONSchema{
				Type:  JSONSchemaTypeObject,
				Prim:  "option",
				Title: "@option_1",
				Default: &JSONSchema{
					SchemaKey: &SchemaKey{
						Type:  JSONSchemaTypeString,
						Const: "none",
					},
				},
				OneOf: []*JSONSchema{
					{
						Title: "None",
						Properties: map[string]*JSONSchema{
							"schemaKey": {
								Type:  JSONSchemaTypeString,
								Const: "none",
							},
						},
					},
					{
						Title: "Some",
						Properties: map[string]*JSONSchema{
							"schemaKey": {
								Type:  JSONSchemaTypeString,
								Const: "some",
							},
							"pour_auth": {
								Type:    JSONSchemaTypeString,
								Title:   "pour_auth",
								Prim:    "signature",
								Default: "",
							},
							"pour_amount": {
								Type:    JSONSchemaTypeInt,
								Title:   "pour_amount",
								Prim:    "mutez",
								Default: 0,
							},
						},
					},
				},
			},
		}, {
			name: "Case 13: tzBTC upgrade",
			data: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":currentVersion"]},{"prim":"nat","annots":[":newVersion"]}]},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationScript"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newCode"]},{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newPermCode"]}]}]}],"annots":["%upgrade"]}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"migrationScript": {
						Prim:    "lambda",
						Type:    JSONSchemaTypeString,
						Title:   "migrationScript",
						Default: "",
					},
					"currentVersion": {
						Prim:    "nat",
						Type:    JSONSchemaTypeInt,
						Title:   "currentVersion",
						Default: 0,
					},
					"newVersion": {
						Prim:    "nat",
						Type:    JSONSchemaTypeInt,
						Title:   "newVersion",
						Default: 0,
					},
					"newPermCode": {
						Type:  JSONSchemaTypeObject,
						Prim:  "option",
						Title: "newPermCode",
						Default: &JSONSchema{
							SchemaKey: &SchemaKey{
								Type:  "string",
								Const: "none",
							},
						},
						OneOf: []*JSONSchema{
							{
								Title: "None",
								Properties: map[string]*JSONSchema{
									"schemaKey": {
										Type:  JSONSchemaTypeString,
										Const: "none",
									},
								},
							},
							{
								Title: "Some",
								Properties: map[string]*JSONSchema{
									"schemaKey": {
										Type:  JSONSchemaTypeString,
										Const: "some",
									},
									"@lambda_21": {
										Type:    JSONSchemaTypeString,
										Prim:    "lambda",
										Title:   "@lambda_21",
										Default: "",
									},
								},
							},
						},
					},
					"newCode": {
						Type:  "object",
						Prim:  "option",
						Title: "newCode",
						Default: &JSONSchema{
							SchemaKey: &SchemaKey{
								Type:  JSONSchemaTypeString,
								Const: "none",
							},
						},
						OneOf: []*JSONSchema{
							{
								Title: "None",
								Properties: map[string]*JSONSchema{
									"schemaKey": {
										Type:  JSONSchemaTypeString,
										Const: "none",
									},
								},
							},
							{
								Title: "Some",
								Properties: map[string]*JSONSchema{
									"schemaKey": {
										Type:  JSONSchemaTypeString,
										Const: "some",
									},
									"@lambda_12": {
										Type:    JSONSchemaTypeString,
										Prim:    "lambda",
										Title:   "@lambda_12",
										Default: "",
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
			var a UntypedAST
			if err := json.UnmarshalFromString(tt.data, &a); err != nil {
				t.Errorf("TypedAst.ToJSONSchema() UnmarshalFromString error = %v", err)
				return
			}
			ta, err := a.ToTypedAST()
			if err != nil {
				t.Errorf("TypedAst.ToJSONSchema() ToTypedAST error = %v", err)
				return
			}

			got, err := ta.ToJSONSchema()
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.ToJSONSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
