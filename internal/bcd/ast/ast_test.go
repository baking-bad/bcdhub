package ast

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/types"
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
								Const: "L",
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
								Const: "R",
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
									"@lambda_30": {
										Type:    JSONSchemaTypeString,
										Prim:    "lambda",
										Title:   "@lambda_30",
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
									"@lambda_15": {
										Type:    JSONSchemaTypeString,
										Prim:    "lambda",
										Title:   "@lambda_15",
										Default: "",
									},
								},
							},
						},
					},
				},
			},
		}, {
			name: "Case 14: big map field",
			data: `{"prim": "big_map", "annots": [":debit"], "args": [{"prim":"address"}, {"prim":"nat"}]}`,
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
			name: "Case 15: set field",
			data: `{"prim":"set", "args":[{"prim":"int"}]}`,
			want: &JSONSchema{
				Type: JSONSchemaTypeObject,
				Properties: map[string]*JSONSchema{
					"@set_1": {
						Type:    JSONSchemaTypeArray,
						Prim:    "set",
						Title:   "@set_1",
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
			name: "Case 15: contract with tag field",
			data: `{"prim":"contract","args":[{"prim":"nat"}],"annots":["%get_countdown_milliseconds"]}`,
			want: &JSONSchema{
				Type:    JSONSchemaTypeString,
				Prim:    "contract",
				Title:   "get_countdown_milliseconds",
				Default: "",
				Tag:     ContractTagViewNat,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta, err := NewTypedAstFromString(tt.data)
			if err != nil {
				t.Errorf("NewTypedAstFromString error = %v", err)
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

func TestTypedAst_Docs(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		entrypoint string
		want       string
		wantErr    bool
	}{
		{
			name: "mainnet/KT1VsSxSXUkgw6zkBGgUuDXXuJs9ToPqkrCg/VestedFunds4",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"contract","args":[{"prim":"unit"}],"annots":["%dest"]},{"prim":"mutez","annots":["%transfer_amount"]}],"annots":["%Transfer"]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"contract","args":[{"prim":"unit"}],"annots":["%pour_dest"]},{"prim":"key","annots":["%pour_authorizer"]}]}],"annots":["%Set_pour"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"key"}],"annots":["%signatories"]},{"prim":"nat","annots":["%group_threshold"]}]}],"annots":["%key_groups"]},{"prim":"nat","annots":["%overall_threshold"]}],"annots":["%Set_keys"]},{"prim":"option","args":[{"prim":"key_hash","annots":["%new_delegate"]}],"annots":["%Set_delegate"]}]}],"annots":["%action_input"]},{"prim":"list","args":[{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%signatures"]}],"annots":["%Action"]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"signature","annots":["%pour_auth"]},{"prim":"mutez","annots":["%pour_amount"]}],"annots":["%Pour"]}]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"Action","value":"$Action"},{"key":"@option_28","value":"option($Pour)"}]},{"name":"Action","type":"pair","args":[{"key":"action_input","value":"$action_input"},{"key":"signatures","value":"list(list(option(signature)))"}]},{"name":"action_input","type":"or","args":[{"key":"Transfer","value":"$Transfer"},{"key":"Set_pour","value":"option($Set_pour)"},{"key":"Set_keys","value":"$Set_keys"},{"key":"Set_delegate","value":"option(key_hash)"}]},{"name":"Transfer","type":"pair","args":[{"key":"dest","value":"contract(unit)"},{"key":"transfer_amount","value":"mutez"}]},{"name":"Set_pour","type":"pair","args":[{"key":"pour_dest","value":"contract(unit)"},{"key":"pour_authorizer","value":"key"}]},{"name":"Set_keys","type":"pair","args":[{"key":"key_groups","value":"list($key_groups_item)"},{"key":"overall_threshold","value":"nat"}]},{"name":"key_groups_item","type":"pair","args":[{"key":"signatories","value":"list(key)"},{"key":"group_threshold","value":"nat"}]},{"name":"signatures_item","type":"list(option(signature))"},{"name":"Pour","type":"pair","args":[{"key":"pour_auth","value":"signature"},{"key":"pour_amount","value":"mutez"}]}]`,
		}, {
			name: "mainnet/KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn/tzBTC",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getVersion"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":spender"]}]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getAllowance"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getBalance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalSupply"]},{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalMinted"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalBurned"]},{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address"}]}],"annots":["%getOwner"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address"}]}],"annots":["%getRedeemAddress"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"nat"}]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"string"},{"prim":"pair","args":[{"prim":"string"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"string"},{"prim":"string"}]}]}]}]}]}]}]}],"annots":["%getTokenMetadata"]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%run"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":currentVersion"]},{"prim":"nat","annots":[":newVersion"]}]},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationScript"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newCode"]},{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newPermCode"]}]}]}],"annots":["%upgrade"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":current"]},{"prim":"nat","annots":[":new"]}],"annots":["%epwBeginUpgrade"]},{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationscript","%epwApplyMigration"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}],"annots":[":contractcode","%epwSetCode"]},{"prim":"unit","annots":["%epwFinishUpgrade"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]},{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":["%approve"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}],"annots":["%mint"]},{"prim":"nat","annots":[":value","%burn"]}]},{"prim":"or","args":[{"prim":"address","annots":[":operator","%addOperator"]},{"prim":"address","annots":[":operator","%removeOperator"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":[":redeem","%setRedeemAddress"]},{"prim":"unit","annots":["%pause"]}]},{"prim":"or","args":[{"prim":"unit","annots":["%unpause"]},{"prim":"or","args":[{"prim":"address","annots":[":newOwner","%transferOwnership"]},{"prim":"unit","annots":["%acceptOwnership"]}]}]}]}]}],"annots":["%safeEntrypoints"]}]}]}]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"getVersion","value":"$getVersion"},{"key":"getAllowance","value":"$getAllowance"},{"key":"getBalance","value":"$getBalance"},{"key":"getTotalSupply","value":"$getTotalSupply"},{"key":"getTotalMinted","value":"$getTotalMinted"},{"key":"getTotalBurned","value":"$getTotalBurned"},{"key":"getOwner","value":"$getOwner"},{"key":"getRedeemAddress","value":"$getRedeemAddress"},{"key":"getTokenMetadata","value":"$getTokenMetadata"},{"key":"run","value":"$run"},{"key":"upgrade","value":"$upgrade"},{"key":"epwBeginUpgrade","value":"$epwBeginUpgrade"},{"key":"epwApplyMigration","value":"$epwApplyMigration"},{"key":"epwSetCode","value":"$epwSetCode"},{"key":"epwFinishUpgrade","value":"unit"},{"key":"transfer","value":"$transfer"},{"key":"approve","value":"$approve"},{"key":"mint","value":"$mint"},{"key":"burn","value":"nat"},{"key":"addOperator","value":"address"},{"key":"removeOperator","value":"address"},{"key":"setRedeemAddress","value":"address"},{"key":"pause","value":"unit"},{"key":"unpause","value":"unit"},{"key":"transferOwnership","value":"address"},{"key":"acceptOwnership","value":"unit"}]},{"name":"getVersion","type":"pair","args":[{"key":"@unit_5","value":"unit"},{"key":"@contract_6","value":"contract(nat)"}]},{"name":"getAllowance","type":"pair","args":[{"key":"owner","value":"address"},{"key":"spender","value":"address"},{"key":"@contract_12","value":"contract(nat)"}]},{"name":"getBalance","type":"pair","args":[{"key":"owner","value":"address"},{"key":"@contract_17","value":"contract(nat)"}]},{"name":"getTotalSupply","type":"pair","args":[{"key":"@unit_21","value":"unit"},{"key":"@contract_22","value":"contract(nat)"}]},{"name":"getTotalMinted","type":"pair","args":[{"key":"@unit_25","value":"unit"},{"key":"@contract_26","value":"contract(nat)"}]},{"name":"getTotalBurned","type":"pair","args":[{"key":"@unit_31","value":"unit"},{"key":"@contract_32","value":"contract(nat)"}]},{"name":"getOwner","type":"pair","args":[{"key":"@unit_35","value":"unit"},{"key":"@contract_36","value":"contract(address)"}]},{"name":"getRedeemAddress","type":"pair","args":[{"key":"@unit_40","value":"unit"},{"key":"@contract_41","value":"contract(address)"}]},{"name":"getTokenMetadata","type":"pair","args":[{"key":"@list_45","value":"list(nat)"},{"key":"@contract_47","value":"contract($contract_47_param)"}]},{"name":"@contract_47_param","type":"list (pair nat (pair string (pair string (pair nat (map string string)))))"},{"name":"run","type":"pair","args":[{"key":"@string_65","value":"string"},{"key":"@bytes_66","value":"bytes"}]},{"name":"upgrade","type":"pair","args":[{"key":"currentVersion","value":"nat"},{"key":"newVersion","value":"nat"},{"key":"migrationScript","value":"$migrationScript"},{"key":"newCode","value":"option($newCode)"},{"key":"newPermCode","value":"option($newPermCode)"}]},{"name":"migrationScript","type":"lambda","args":[{"key":"input","value":"big_map bytes bytes"},{"key":"return","value":"big_map bytes bytes"}]},{"name":"newCode","type":"lambda","args":[{"key":"input","value":"pair (pair string bytes) (big_map bytes bytes)"},{"key":"return","value":"pair (list operation) (big_map bytes bytes)"}]},{"name":"newPermCode","type":"lambda","args":[{"key":"input","value":"pair unit (big_map bytes bytes)"},{"key":"return","value":"pair (list operation) (big_map bytes bytes)"}]},{"name":"epwBeginUpgrade","type":"pair","args":[{"key":"current","value":"nat"},{"key":"new","value":"nat"}]},{"name":"epwApplyMigration","type":"lambda","args":[{"key":"input","value":"big_map bytes bytes"},{"key":"return","value":"big_map bytes bytes"}]},{"name":"epwSetCode","type":"lambda","args":[{"key":"input","value":"pair (pair string bytes) (big_map bytes bytes)"},{"key":"return","value":"pair (list operation) (big_map bytes bytes)"}]},{"name":"transfer","type":"pair","args":[{"key":"from","value":"address"},{"key":"to","value":"address"},{"key":"value","value":"nat"}]},{"name":"approve","type":"pair","args":[{"key":"spender","value":"address"},{"key":"value","value":"nat"}]},{"name":"mint","type":"pair","args":[{"key":"to","value":"address"},{"key":"value","value":"nat"}]}]`,
		}, {
			name: "mainnet/KT1U1JZaXoG4u1EPnhHL4R4otzkWc1L34q3c/Equisafe-KYC-registrar",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}],"annots":["%1"]}]}],"annots":["%addMembers"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country_invest_limit"]},{"prim":"nat","annots":["%min_rating"]}]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%rating_restrictions"]},{"prim":"timestamp","annots":["%vesting"]}]}]}],"annots":["%1"]}]},{"prim":"bool","annots":["%2"]}],"annots":["%checkMember"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}]},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%1"]}],"annots":["%getMember"]},{"prim":"list","args":[{"prim":"address"}],"annots":["%removeMembers"]}]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"addMembers","value":"list($addMembers_item)"},{"key":"checkMember","value":"$checkMember"},{"key":"getMember","value":"$getMember"},{"key":"removeMembers","value":"list(address)"}]},{"name":"addMembers_item","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"$1"}]},{"name":"1","type":"pair","args":[{"key":"country","value":"nat"},{"key":"expires","value":"timestamp"},{"key":"rating","value":"nat"},{"key":"region","value":"nat"},{"key":"restricted","value":"bool"}]},{"name":"checkMember","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"map(nat, $1_value)"},{"key":"2","value":"bool"}]},{"name":"1_value","type":"pair","args":[{"key":"country_invest_limit","value":"nat"},{"key":"min_rating","value":"nat"},{"key":"rating_restrictions","value":"map(nat, nat)"},{"key":"vesting","value":"timestamp"}]},{"name":"getMember","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"$1"}]},{"name":"1","type":"lambda","args":[{"key":"input","value":"pair (pair (pair (nat %country) (timestamp %expires)) (pair (nat %rating) (nat %region))) (bool %restricted)"},{"key":"return","value":"list operation"}]}]`,
		}, {
			name:       "mainnet/KT1U1JZaXoG4u1EPnhHL4R4otzkWc1L34q3c/Equisafe-KYC-registrar/addMembers",
			data:       `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}],"annots":["%1"]}]}],"annots":["%addMembers"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country_invest_limit"]},{"prim":"nat","annots":["%min_rating"]}]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%rating_restrictions"]},{"prim":"timestamp","annots":["%vesting"]}]}]}],"annots":["%1"]}]},{"prim":"bool","annots":["%2"]}],"annots":["%checkMember"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}]},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%1"]}],"annots":["%getMember"]},{"prim":"list","args":[{"prim":"address"}],"annots":["%removeMembers"]}]}]}]}]`,
			entrypoint: "addMembers",
			want:       `[{"name":"addMembers","type":"list($addMembers_item)"},{"name":"addMembers_item","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"$1"}]},{"name":"1","type":"pair","args":[{"key":"country","value":"nat"},{"key":"expires","value":"timestamp"},{"key":"rating","value":"nat"},{"key":"region","value":"nat"},{"key":"restricted","value":"bool"}]}]`,
		}, {
			name: "mainnet/KT1ChNsEFxwyCbJyWGSL3KdjeXE28AY1Kaog/BakersRegistry",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"key_hash","annots":["%delegate"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%bakerName"]},{"prim":"bool","annots":["%openForDelegation"]}]},{"prim":"bytes","annots":["%bakerOffchainRegistryUrl"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%split"]},{"prim":"list","args":[{"prim":"address"}],"annots":["%bakerPaysFromAccounts"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%minDelegation"]},{"prim":"bool","annots":["%subtractPayoutsLessThanMin"]}]},{"prim":"pair","args":[{"prim":"int","annots":["%payoutDelay"]},{"prim":"pair","args":[{"prim":"nat","annots":["%payoutFrequency"]},{"prim":"int","annots":["%minPayout"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bool","annots":["%bakerChargesTransactionFee"]},{"prim":"nat","annots":["%paymentConfigMask"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%overDelegationThreshold"]},{"prim":"bool","annots":["%subtractRewardsFromUninvitedDelegation"]}]}]}]}]}]}],"annots":["%data"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%reporterAccount"]}]}],"annots":["%set_data"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%signup_fee"]},{"prim":"mutez","annots":["%update_fee"]}],"annots":["%set_fees"]},{"prim":"contract","args":[{"prim":"unit"}],"annots":["%withdraw"]}]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"set_data","value":"$set_data"},{"key":"set_fees","value":"$set_fees"},{"key":"withdraw","value":"contract(unit)"}]},{"name":"set_data","type":"pair","args":[{"key":"delegate","value":"key_hash"},{"key":"data","value":"option($data)"},{"key":"reporterAccount","value":"option(address)"}]},{"name":"data","type":"pair","args":[{"key":"bakerName","value":"bytes"},{"key":"openForDelegation","value":"bool"},{"key":"bakerOffchainRegistryUrl","value":"bytes"},{"key":"split","value":"nat"},{"key":"bakerPaysFromAccounts","value":"list(address)"},{"key":"minDelegation","value":"nat"},{"key":"subtractPayoutsLessThanMin","value":"bool"},{"key":"payoutDelay","value":"int"},{"key":"payoutFrequency","value":"nat"},{"key":"minPayout","value":"int"},{"key":"bakerChargesTransactionFee","value":"bool"},{"key":"paymentConfigMask","value":"nat"},{"key":"overDelegationThreshold","value":"nat"},{"key":"subtractRewardsFromUninvitedDelegation","value":"bool"}]},{"name":"set_fees","type":"pair","args":[{"key":"signup_fee","value":"mutez"},{"key":"update_fee","value":"mutez"}]}]`,
		}, {
			name: "edonet/KT1D7MfG9CEBav7TXsa4xbPL3QZgR5eEgx7g/ticket",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"ticket","args":[{"prim":"address"}]},{"prim":"nat"}],"annots":["%sendConvertedBalance"]},{"prim":"mutez","annots":["%setConversionRate"]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"sendConvertedBalance","value":"$sendConvertedBalance"},{"key":"setConversionRate","value":"mutez"}]},{"name":"sendConvertedBalance","type":"pair","args":[{"key":"@ticket_3","value":"ticket(address)"},{"key":"@nat_9","value":"nat"}]}]`,
		}, {
			name: "edonet/KT1MaW1LQ77YpZwtmrb4aHBUteqPN91AruWB/sapling_state",
			data: `[{"prim":"parameter","args":[{"prim":"sapling_state","args":[{"int":"8"}]}]}]`,
			want: `[{"name":"@sapling_state_1","type":"sapling_state(8)"}]`,
		}, {
			name: "edonet/KT1PbFKg3mgJAadojxPjh3EQSLNsuAkYjNnQ/sapling_transaction",
			data: `[{"prim":"parameter","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}]}]`,
			want: `[{"name":"@list_1","type":"list($list_1_item)"},{"name":"@list_1_item","type":"pair","args":[{"key":"@sapling_transaction_3","value":"sapling_transaction(8)"},{"key":"@option_4","value":"option(key_hash)"}]}]`,
		}, {
			name: "unknown contract 1",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"nat"}],"annots":["%setList"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%setMap"]}]},{"prim":"set","args":[{"prim":"nat"}],"annots":["%setSet"]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"setList","value":"list(nat)"},{"key":"setMap","value":"map(nat, mutez)"},{"key":"setSet","value":"set(nat)"}]}]`,
		}, {
			name: "delphinet/KT1Po9Xr5wgj4aeuXhYWx4wRQQyJgoV26KWp/storage",
			data: `[{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%arbitraryValues"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%lambdas"]}]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"address"}],"annots":["%ovens"]}]}]}]`,
			want: `[{"name":"@pair_1","type":"pair","args":[{"key":"arbitraryValues","value":"big_map(string, bytes)"},{"key":"lambdas","value":"big_map(string, bytes)"},{"key":"ovens","value":"big_map(address, address)"}]}]`,
		}, {
			name: "list in bigmap",
			data: `[{"prim":"storage","args":[{"prim": "big_map","args":[{"prim":"int"},{"prim":"list","args":[{"prim":"int"}]}]}]}]`,
			want: `[{"name":"@big_map_1","type":"big_map(int, list(int))"}]`,
		}, {
			name: "map in bigmap",
			data: `[{"prim":"storage","args":[{"prim": "big_map","args":[{"prim":"int"},{"prim":"map","args":[{"prim":"int"},{"prim":"int"}]}]}]}]`,
			want: `[{"name":"@big_map_1","type":"big_map(int, map(int, int))"}]`,
		}, {
			name: "option in bigmap",
			data: `[{"prim":"storage","args":[{"prim": "big_map","args":[{"prim":"int"},{"prim":"option", "args":[{"prim":"map","args":[{"prim":"int"},{"prim":"int"}]}]}]}]}]`,
			want: `[{"name":"@big_map_1","type":"big_map(int, option(map(int, int)))"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			untyped, err := NewScript([]byte(tt.data))
			if err != nil {
				t.Errorf("NewScript() error = %v", err)
				return
			}
			var a *TypedAst
			switch {
			case len(untyped.Parameter) > 0:
				a, err = untyped.Parameter.ToTypedAST()
			case len(untyped.Storage) > 0:
				a, err = untyped.Storage.ToTypedAST()
			default:
				t.Errorf("Need to set parameter or storage")
				return
			}
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			got, err := a.Docs(tt.entrypoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Docs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotStr, err := json.MarshalToString(got)
			if err != nil {
				t.Errorf("MarshalToString() error = %v", err)
				return
			}
			if gotStr != tt.want {
				t.Errorf("TypedAst.Docs() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

func TestTypedAst_Compare(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		a       string
		b       string
		want    int
		wantErr bool
	}{
		{
			name: "simple true",
			typ:  `[{"prim": "string"}]`,
			a:    `{"string": "test"}`,
			b:    `{"string": "test"}`,
			want: 0,
		}, {
			name: "simple false",
			typ:  `[{"prim": "string"}]`,
			a:    `{"string": "test"}`,
			b:    `{"string": "another"}`,
			want: 1,
		}, {
			name: "pair with option None true",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			want: 0,
		}, {
			name: "pair with option Some true",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"bytes": "0000cd1a410ffd5315ded34337f5f76edff48a13999a"}]}]}`,
			want: 0,
		}, {
			name: "pair with option false",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			want: 1,
		}, {
			name: "pair with option Some false",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A"}]}]}`,
			want: 1,
		}, {
			name:    "pair with option uncomparable false",
			typ:     `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "set", "args":[{"prim": "unit"}]}]}]}]`,
			a:       `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[[]]}]}`,
			b:       `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[[]]}]}`,
			want:    0,
			wantErr: true,
		}, {
			name: "",
			typ:  `{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"key"},{"prim":"pair","args":[{"prim":"key_hash"},{"prim":"pair","args":[{"prim":"signature"},{"prim":"pair","args":[{"prim":"chain_id"},{"prim":"timestamp"}]}]}]}]}]}`,
			a:    `{"prim":"Pair","args":[{"bytes":"01ec634b3c5e0ecfe4899310e4e5e4a0a87b8f117500"},{"prim":"Pair","args":[{"bytes":"00419491b1796b13d756d394ed925c10727bca06e97353c5ca09402a9b6b07abcc"},{"prim":"Pair","args":[{"bytes":"00ccf564a5a0bdb15c3dbdf84d68dacac3e1f968a3"},{"prim":"Pair","args":[{"bytes":"cabde71255b1f1674182cb7f8000903909dbe6dbb6a76afd3376c08a4f827b2bc938ed447f3e592766e89aea89fecfd1e8c8e82c71f60271cd08ac012262d603"},{"prim":"Pair","args":[{"bytes":"7a06a770"},{"int":"1607457231"}]}]}]}]}]}`,
			b:    `{"prim":"Pair","args":[{"bytes":"01ec634b3c5e0ecfe4899310e4e5e4a0a87b8f117500"},{"prim":"Pair","args":[{"bytes":"00419491b1796b13d756d394ed925c10727bca06e97353c5ca09402a9b6b07abcc"},{"prim":"Pair","args":[{"bytes":"00ccf564a5a0bdb15c3dbdf84d68dacac3e1f968a3"},{"prim":"Pair","args":[{"bytes":"cabde71255b1f1674182cb7f8000903909dbe6dbb6a76afd3376c08a4f827b2bc938ed447f3e592766e89aea89fecfd1e8c8e82c71f60271cd08ac012262d603"},{"prim":"Pair","args":[{"string":"NetXdQprcVkpaWU"},{"int":"1607457231"}]}]}]}]}]}`,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typ UntypedAST
			if err := json.Unmarshal([]byte(tt.typ), &typ); err != nil {
				t.Errorf("Unmarshal(typA) error = %v", err)
				return

			}
			typA, err := typ.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			typB, err := typ.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			var aTree UntypedAST
			if err := json.Unmarshal([]byte(tt.a), &aTree); err != nil {
				t.Errorf("Unmarshal(a) error = %v", err)
				return

			}
			var bTree UntypedAST
			if err := json.Unmarshal([]byte(tt.b), &bTree); err != nil {
				t.Errorf("Unmarshal(b) error = %v", err)
				return

			}
			if err := typA.Settle(aTree); err != nil {
				t.Errorf("typA.Settle error = %v", err)
				return

			}
			if err := typB.Settle(bTree); err != nil {
				t.Errorf("typA.Settle error = %v", err)
				return

			}

			got, err := typA.Compare(typB)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Compare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TypedAst.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypedAst_Settle(t *testing.T) {
	tests := []struct {
		name       string
		tree       string
		data       string
		entrypoint string
		wantErr    bool
		want       string
	}{
		{
			name: "string",
			tree: `[{"prim":"string"}]`,
			data: `{"string":"test"}`,
			want: `{"string":"test"}`,
		}, {
			name:       "atomex",
			tree:       `[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}]`,
			data:       `{"prim":"Right","args":[{"prim":"Left","args":[{"bytes":"8d1c54042ee5a32d3eb5732d86e32efae058d409f32dbb4859142a1589cc4a3e"}]}]}`,
			want:       `{"bytes":"8d1c54042ee5a32d3eb5732d86e32efae058d409f32dbb4859142a1589cc4a3e"}`,
			entrypoint: "redeem",
		}, {
			name: "tzbtc transfer",
			tree: `[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]}]`,
			data: `{"prim":"Pair","args":[{"string":"tz1QnR36rKBcMLPiTP8TTXD8HmGbXBo2HEZH"},{"prim":"Pair","args":[{"int":"2038026"},{"string":"2021-02-09T18:22:14Z"}]}]}`,
			want: `{"prim":"Pair","args":[{"string":"tz1QnR36rKBcMLPiTP8TTXD8HmGbXBo2HEZH"},{"prim":"Pair","args":[{"int":"2038026"},{"string":"2021-02-09T18:22:14Z"}]}]}`,
		}, {
			name: "setList",
			tree: `[{"prim":"list","args":[{"prim":"nat"}],"annots":["%setList"]}]`,
			data: `[{"int":"2"}]`,
			want: `[{"int":"2"}]`,
		}, {
			name: "setMap",
			tree: `{"prim":"map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%setMap"]}`,
			data: `[{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]`,
			want: `[{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]`,
		}, {
			name: "setBigMap",
			tree: `{"prim":"big_map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%setMap"]}`,
			data: `[{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]`,
			want: `[{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]`,
		}, {
			name: "setSet",
			tree: `{"prim":"set","args":[{"prim":"nat"}],"annots":["%setSet"]}`,
			data: `[{"int":"2"}]`,
			want: `[{"int":"2"}]`,
		}, {
			name: "option",
			tree: `{"prim":"pair","args":[{"prim":"string","annots":["%params"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key"}],"annots":["%public_key"]},{"prim":"option","args":[{"prim":"signature"}],"annots":["%sig"]}]}]}`,
			data: `{"prim":"Pair","args":[{"string":"frfe"},{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"edpktv7KGuCdHVG9Ys1uJ8my3b1HuWKzaW2A2vmJ5uSPfwjwnh81Ly"}]},{"prim":"Some","args":[{"string":"sigrTtiiUxV51dF15yhiPr36XFybypu7EUu8Lkq2qKGUDj9HxhCRRZukHGg1QEAopBvnqMjdtiejPbECm6RM8TqK8kffhtZ3"}]}]}]}`,
			want: `{"prim":"Pair","args":[{"string":"frfe"},{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"edpktv7KGuCdHVG9Ys1uJ8my3b1HuWKzaW2A2vmJ5uSPfwjwnh81Ly"}]},{"prim":"Some","args":[{"string":"sigrTtiiUxV51dF15yhiPr36XFybypu7EUu8Lkq2qKGUDj9HxhCRRZukHGg1QEAopBvnqMjdtiejPbECm6RM8TqK8kffhtZ3"}]}]}]}`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default",
			tree: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			data: `[{"prim":"Pair","args":[{"bytes":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"Some","args":[{"bytes":"00f1c4ee52908e89b832d47ef72be0d29bf326d245"}]}]}]`,
			want: `[{"prim":"Pair","args":[{"bytes":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"Some","args":[{"bytes":"00f1c4ee52908e89b832d47ef72be0d29bf326d245"}]}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := NewSettledTypedAst(tt.tree, tt.data)
			if err != nil {
				t.Errorf("NewSettledTypedAstFromString() error = %v", err)
				return
			}
			if !typ.IsSettled() {
				t.Errorf("tree is not settled")
				return
			}

			b, err := typ.ToParameters(tt.entrypoint)
			if err != nil {
				t.Errorf("ToParameters() error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestTypedAst_GetEntrypoints(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want []string
	}{
		{
			name: "simple",
			tree: `{"prim": "string"}`,
			want: []string{"default"},
		}, {
			name: "mainnet/KT1RHkGHmMTvi4wimZYEbuV1gfY9MGm8meWg",
			tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}]}]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}]},{"prim":"address"}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]},{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]},{"prim":"unit"}]}]}]}`,
			want: []string{
				"entrypoint_0",
				"entrypoint_1",
				"entrypoint_2",
				"entrypoint_3",
				"entrypoint_4",
				"entrypoint_5",
				"entrypoint_6",
				"entrypoint_7",
			},
		}, {
			name: "mainnet/atomex",
			tree: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			want: []string{
				"initiate",
				"add",
				"redeem",
				"refund",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UnmarshalFromString tree error = %v", err)
				return
			}
			typ, err := tree.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			if got := typ.GetEntrypoints(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypedAst.GetEntrypoints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypedAst_ToMiguel(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name: "atomex storage",
			tree: `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			data: `{"prim":"Pair","args":[{"int":"4"},{"prim":"Unit"}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"big_map","type":"big_map","name":"@big_map_2","value":4},{"prim":"unit","type":"unit","name":"@unit_13"}]}]`,
		}, {
			name: "atomex redeem parameters",
			tree: `{"prim":"bytes","annots":[":secret","%redeem"]}`,
			data: `{"bytes":"8d1c54042ee5a32d3eb5732d86e32efae058d409f32dbb4859142a1589cc4a3e"}`,
			want: `[{"prim":"bytes","type":"bytes","name":"secret","value":"8d1c54042ee5a32d3eb5732d86e32efae058d409f32dbb4859142a1589cc4a3e"}]`,
		}, {
			name: "atomex initiate parameters",
			tree: `{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]}`,
			data: `{"prim":"Pair","args":[{"string":"tz1N5wUqaxLKLDMvLA1D8GJu1twPvjasMMVy"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"758af1ebc445c9561b6bf1f3de4c89c4b1b91c0ffc9d6b8325496734f09ef161"},{"int":"1612869027"}]},{"int":"0"}]}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"initiate","children":[{"prim":"address","type":"address","name":"participant","value":"tz1N5wUqaxLKLDMvLA1D8GJu1twPvjasMMVy"},{"prim":"pair","type":"namedtuple","name":"settings","children":[{"prim":"bytes","type":"bytes","name":"hashed_secret","value":"758af1ebc445c9561b6bf1f3de4c89c4b1b91c0ffc9d6b8325496734f09ef161"},{"prim":"timestamp","type":"timestamp","name":"refund_time","value":"2021-02-09T11:10:27Z"},{"prim":"mutez","type":"mutez","name":"payoff","value":"0"}]}]}]`,
		}, {
			name: "delphinet/KT198xhayXQwcp9Ab1LZPFKuCHJPtQfVQ1Mv/balance_of",
			tree: `{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%requests"]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%request"]},{"prim":"nat","annots":["%balance"]}]}]}],"annots":["%callback"]}],"annots":["%balance_of"]}`,
			data: `{"prim":"Pair","args":[[{"prim":"Pair","args":[{"bytes":"00005f2c9f728eb79637dbea7ed4d30d7a82da1ee2c8"},{"int":"0"}]}],{"bytes":"016172923975c5600aa87af674a5d522a7579eec0000"}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"balance_of","children":[{"prim":"list","type":"list","name":"requests","children":[{"prim":"pair","type":"namedtuple","name":"@pair_3","children":[{"prim":"address","type":"address","name":"owner","value":"tz1UKGLnvAYg7LNiBe5GWkq8oAfFi9DQP9fj"},{"prim":"nat","type":"nat","name":"token_id","value":"0"}]}]},{"prim":"contract","type":"contract","name":"callback","value":"KT1HU2MvEumzrwLoRZpLs1WtdiCnaz7eMAtk"}]}]`,
		}, {
			name: "edonet/KT1ShfJ1EoBVkyTb2uJcs3inS2A1yjiDC8AD",
			tree: `{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"key"},{"prim":"pair","args":[{"prim":"key_hash"},{"prim":"pair","args":[{"prim":"signature"},{"prim":"pair","args":[{"prim":"chain_id"},{"prim":"timestamp"}]}]}]}]}]}`,
			data: `{"prim":"Pair","args":[{"bytes":"01ec634b3c5e0ecfe4899310e4e5e4a0a87b8f117500"},{"prim":"Pair","args":[{"bytes":"00419491b1796b13d756d394ed925c10727bca06e97353c5ca09402a9b6b07abcc"},{"prim":"Pair","args":[{"bytes":"00ccf564a5a0bdb15c3dbdf84d68dacac3e1f968a3"},{"prim":"Pair","args":[{"bytes":"cabde71255b1f1674182cb7f8000903909dbe6dbb6a76afd3376c08a4f827b2bc938ed447f3e592766e89aea89fecfd1e8c8e82c71f60271cd08ac012262d603"},{"prim":"Pair","args":[{"bytes":"7a06a770"},{"int":"1607457231"}]}]}]}]}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"address","type":"address","name":"@address_2","value":"KT1W8fz6dG7Cg6Lnac6dvSD8dV7Ts7q4Pdr8"},{"prim":"key","type":"key","name":"@key_4","value":"edpku976gpuAD2bXyx1XGraeKuCo1gUZ3LAJcHM12W1ecxZwoiu22R"},{"prim":"key_hash","type":"key_hash","name":"@key_hash_6","value":"tz1eKkWU5hGtfLUiqNpucHrXymm83z3DG9Sq"},{"prim":"signature","type":"signature","name":"@signature_8","value":"sigpWi99kgMJEUxDtF9HykqfASTiKiHhuhuBh1YtszXnuDmSfn8u1uPVyiD4JWDeNCxVS81yxZZwpKHnezaUd8s7snjiEGvq"},{"prim":"chain_id","type":"chain_id","name":"@chain_id_10","value":"NetXdQprcVkpaWU"},{"prim":"timestamp","type":"timestamp","name":"@timestamp_11","value":"2020-12-08T19:53:51Z"}]}]`,
		}, {
			name: "edonet/KT1N6VjvuuBfXBbsyMby96zkYeaWuqCto69Q/send",
			tree: `{"prim":"pair","args":[{"prim":"contract","args":[{"prim":"ticket","args":[{"prim":"unit"}]}],"annots":["%destination"]},{"prim":"nat","annots":["%amount"]},{"prim":"address","annots":["%ticketer"]}],"annots":["%send"]}`,
			data: `{"prim":"Pair","args":[{"string":"KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ%receive"},{"prim":"Pair","args":[{"int":"1"},{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"}]}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"send","children":[{"prim":"contract","type":"contract","name":"destination","value":"KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ%receive"},{"prim":"nat","type":"nat","name":"amount","value":"1"},{"prim":"address","type":"address","name":"ticketer","value":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"}]}]`,
		}, {
			name: "delphinet/KT1NQNzrJ1f8Be9kvstfPF9PnbCrHWe6h3jr/openAuction",
			tree: `{"prim":"map","args":[{"prim":"nat"},{"prim":"int"}],"annots":["%openAuction"]}`,
			data: `[{"prim":"Elt","args":[{"int":"0"},{"int":"8"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"7"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"6"}]},{"prim":"Elt","args":[{"int":"3"},{"int":"5"}]},{"prim":"Elt","args":[{"int":"4"},{"int":"4"}]},{"prim":"Elt","args":[{"int":"5"},{"int":"3"}]},{"prim":"Elt","args":[{"int":"6"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"7"},{"int":"1"}]}]`,
			want: `[{"prim":"map","type":"map","name":"openAuction","children":[{"prim":"int","type":"int","name":"0","value":"8"},{"prim":"int","type":"int","name":"1","value":"7"},{"prim":"int","type":"int","name":"2","value":"6"},{"prim":"int","type":"int","name":"3","value":"5"},{"prim":"int","type":"int","name":"4","value":"4"},{"prim":"int","type":"int","name":"5","value":"3"},{"prim":"int","type":"int","name":"6","value":"2"},{"prim":"int","type":"int","name":"7","value":"1"}]}]`,
		}, {
			name: "delphinet/KT1WTYCnCsXpaAx14dpBAQVLnh4HvttLEhQx/mint",
			tree: `{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%value"]}],"annots":["%mint"]}`,
			data: `{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"00007d3756d2acfb6bc5f117b1779f88fca9062559ae"}]},{"prim":"Some","args":[{"int":"3000"}]}]}`,
			want: `[{"prim":"pair","type":"namedtuple","name":"mint","children":[{"prim":"address","type":"address","name":"address","value":"tz1X47T8UYjjHwaNqtAotbn3dewEeE4kGXWU"},{"prim":"nat","type":"nat","name":"value","value":"3000"}]}]`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default",
			tree: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			data: `[{"prim":"Pair","args":[{"bytes":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"Some","args":[{"bytes":"00f1c4ee52908e89b832d47ef72be0d29bf326d245"}]}]}]`,
			want: `[{"prim":"list","type":"list","name":"@list_1","children":[{"prim":"pair","type":"namedtuple","name":"@pair_2","children":[{"prim":"sapling_transaction","type":"sapling_transaction","name":"@sapling_transaction_3","value":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"key_hash","type":"key_hash","name":"@key_hash_5","value":"tz1hgPTMPor2cmqDpGzgwiWJKMPi84HuntYp"}]}]}]`,
		}, {
			name: "simple big map",
			tree: `{"prim": "big_map","args":[{"prim":"int"},{"prim":"string"}]}`,
			data: `[{"prim":"Elt","args":[{"int":"9"},{"string":"test2"}]},{"prim":"Elt","args":[{"int":"10"},{"string":"test"}]}]`,
			want: `[{"prim":"big_map","type":"big_map","name":"@big_map_1","children":[{"prim":"string","type":"string","name":"9","value":"test2"},{"prim":"string","type":"string","name":"10","value":"test"}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UnmarshalFromString tree error = %v", err)
				return
			}
			typ, err := tree.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			var data UntypedAST
			if err := json.UnmarshalFromString(tt.data, &data); err != nil {
				t.Errorf("UnmarshalFromString data error = %v", err)
				return
			}
			if err := typ.Settle(data); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Settle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := typ.ToMiguel()
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.ToMiguel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, err := json.Marshal(got)
			if err != nil {
				t.Errorf("Marshal(got) data error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestTypedAst_Diff(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		curr    string
		prev    string
		want    string
		wantErr bool
	}{
		{
			name: "atomex redeem",
			tree: `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			curr: `{"args":[[{"args":[{"string":"1f4aa7e6f7ad11a90db94a3e2cb2060bcd7dc5db20b2d12b09c0fd92a15bbceb"},null],"prim":"Elt"}],{"prim":"Unit"}],"prim":"Pair"}`,
			prev: `{"args":[[{"args":[{"string":"1f4aa7e6f7ad11a90db94a3e2cb2060bcd7dc5db20b2d12b09c0fd92a15bbceb"},{"args":[{"args":[{"string":"tz1bNL8YciKPtCuKNzQWxVF8Bnm1h3sd8sbB"},{"string":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"}],"prim":"Pair"},{"args":[{"args":[{"int":"200992827"},{"int":"1600334777"}],"prim":"Pair"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}],{"prim":"Unit"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"big_map","type":"big_map","name":"@big_map_2","children":[{"prim":"pair","type":"namedtuple","name":"1f4aa7e6f7ad11a90db94a3e2cb2060bcd7dc5db20b2d12b09c0fd92a15bbceb","diff_type":"delete","children":[{"prim":"pair","type":"namedtuple","name":"recipients","diff_type":"delete","children":[{"prim":"address","type":"address","name":"initiator","diff_type":"delete","value":"tz1bNL8YciKPtCuKNzQWxVF8Bnm1h3sd8sbB"},{"prim":"address","type":"address","name":"participant","diff_type":"delete","value":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"}]},{"prim":"pair","type":"namedtuple","name":"settings","diff_type":"delete","children":[{"prim":"mutez","type":"mutez","name":"amount","diff_type":"delete","value":"200992827"},{"prim":"timestamp","type":"timestamp","name":"refund_time","diff_type":"delete","value":"2020-09-17T09:26:17Z"},{"prim":"mutez","type":"mutez","name":"payoff","diff_type":"delete","value":"0"}]}]}]},{"prim":"unit","type":"unit","name":"@unit_13"}]}`,
		}, {
			name: "atomex initiate",
			tree: `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			curr: `{"args":[[{"args":[{"string":"314d74100a3d994ae0b52114fb414d86001a0e4a1b122ef1d03ed5ad6b7c4f93"},{"args":[{"args":[{"string":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"},{"string":"tz1U2QDpyre9fqCWXG7MwTE38cR5d7qWoYgC"}],"prim":"Pair"},{"args":[{"args":[{"int":"192258291"},{"int":"1602590721"}],"prim":"Pair"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}],{"prim":"Unit"}],"prim":"Pair"}`,
			prev: `{"args":[{"int":"4"},{"prim":"Unit"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"big_map","type":"big_map","name":"@big_map_2","children":[{"prim":"pair","type":"namedtuple","name":"314d74100a3d994ae0b52114fb414d86001a0e4a1b122ef1d03ed5ad6b7c4f93","diff_type":"create","children":[{"prim":"pair","type":"namedtuple","name":"recipients","diff_type":"create","children":[{"prim":"address","type":"address","name":"initiator","diff_type":"create","value":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"},{"prim":"address","type":"address","name":"participant","diff_type":"create","value":"tz1U2QDpyre9fqCWXG7MwTE38cR5d7qWoYgC"}]},{"prim":"pair","type":"namedtuple","name":"settings","diff_type":"create","children":[{"prim":"mutez","type":"mutez","name":"amount","diff_type":"create","value":"192258291"},{"prim":"timestamp","type":"timestamp","name":"refund_time","diff_type":"create","value":"2020-10-13T12:05:21Z"},{"prim":"mutez","type":"mutez","name":"payoff","diff_type":"create","value":"0"}]}]}]},{"prim":"unit","type":"unit","name":"@unit_13"}]}`,
		}, {
			name: "tzbtc transfer",
			tree: `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"bool"}]}]}]}`,
			curr: `{"args":[[{"args":[{"string":"Pair \"ledger\" \"tz1gVJBEoUQKGzwq6CCBpb8QjCQLejmJDSmU\""},{"string":"Pair 1521320 {}"}],"prim":"Elt"},{"args":[{"string":"\"totalSupply\""},{"string":"35793999960"}],"prim":"Elt"},{"args":[{"string":"Pair \"ledger\" \"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf\""},{"string":"Pair 1153132183 {}"}],"prim":"Elt"}],{"args":[[{"prim":"DUP"},{"prim":"CAR"},{"args":[[{"prim":"CDR"}]],"prim":"DIP"},{"prim":"DUP"},{"prim":"DUP"},{"prim":"CAR"},{"args":[[{"prim":"CDR"}]],"prim":"DIP"},{"args":[[{"args":[{"int":"2"},[{"prim":"DUP"}]],"prim":"DIP"},{"args":[{"int":"2"}],"prim":"DIG"}]],"prim":"DIP"},{"args":[{"prim":"string"},{"string":"code"}],"prim":"PUSH"},{"prim":"PAIR"},{"prim":"PACK"},{"prim":"GET"},{"args":[[{"args":[{"args":[{"args":[{"prim":"bytes"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"NONE"}],[{"args":[{"args":[{"args":[{"prim":"bytes"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"UNPACK"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: failed to unpack code"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"prim":"SOME"}]],"prim":"IF_NONE"},{"args":[[{"prim":"DROP"},{"args":[[{"prim":"DUP"},{"args":[{"prim":"bytes"},{"bytes":"05010000000866616c6c6261636b"}],"prim":"PUSH"},{"prim":"GET"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: no field fallback"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"args":[{"args":[{"args":[{"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"pair"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"UNPACK"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: failed to unpack fallback"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"prim":"SWAP"}]],"prim":"DIP"},{"prim":"PAIR"},{"prim":"EXEC"}],[{"args":[[{"prim":"SWAP"},{"prim":"DROP"},{"prim":"PAIR"}]],"prim":"DIP"},{"prim":"SWAP"},{"prim":"EXEC"}]],"prim":"IF_NONE"}],{"args":[{"int":"1"},{"prim":"False"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			prev: `{"args":[[{"args":[{"string":"\"totalSupply\""},{"string":"35793999960"}],"prim":"Elt"},{"args":[{"string":"Pair \"ledger\" \"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf\""},{"string":"Pair 1152132183 {}"}],"prim":"Elt"},{"args":[{"string":"Pair \"ledger\" \"tz1gVJBEoUQKGzwq6CCBpb8QjCQLejmJDSmU\""},{"string":"Pair 2521320 { Elt \"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf\" 1000000 }"}],"prim":"Elt"}],{"args":[[{"prim":"DUP"},{"prim":"CAR"},{"args":[[{"prim":"CDR"}]],"prim":"DIP"},{"prim":"DUP"},{"prim":"DUP"},{"prim":"CAR"},{"args":[[{"prim":"CDR"}]],"prim":"DIP"},{"args":[[{"args":[{"int":"2"},[{"prim":"DUP"}]],"prim":"DIP"},{"args":[{"int":"2"}],"prim":"DIG"}]],"prim":"DIP"},{"args":[{"prim":"string"},{"string":"code"}],"prim":"PUSH"},{"prim":"PAIR"},{"prim":"PACK"},{"prim":"GET"},{"args":[[{"args":[{"args":[{"args":[{"prim":"bytes"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"NONE"}],[{"args":[{"args":[{"args":[{"prim":"bytes"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"UNPACK"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: failed to unpack code"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"prim":"SOME"}]],"prim":"IF_NONE"},{"args":[[{"prim":"DROP"},{"args":[[{"prim":"DUP"},{"args":[{"prim":"bytes"},{"bytes":"05010000000866616c6c6261636b"}],"prim":"PUSH"},{"prim":"GET"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: no field fallback"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"args":[{"args":[{"args":[{"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"pair"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"},{"args":[{"args":[{"prim":"operation"}],"prim":"list"},{"args":[{"prim":"bytes"},{"prim":"bytes"}],"prim":"big_map"}],"prim":"pair"}],"prim":"lambda"}],"prim":"UNPACK"},{"args":[[{"args":[{"prim":"string"},{"string":"UStore: failed to unpack fallback"}],"prim":"PUSH"},{"prim":"FAILWITH"}],[]],"prim":"IF_NONE"},{"prim":"SWAP"}]],"prim":"DIP"},{"prim":"PAIR"},{"prim":"EXEC"}],[{"args":[[{"prim":"SWAP"},{"prim":"DROP"},{"prim":"PAIR"}]],"prim":"DIP"},{"prim":"SWAP"},{"prim":"EXEC"}]],"prim":"IF_NONE"}],{"args":[{"int":"1"},{"prim":"False"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"big_map","type":"big_map","name":"@big_map_2","children":[{"prim":"bytes","type":"bytes","name":"\"totalSupply\"","value":"35793999960"},{"prim":"bytes","type":"bytes","name":"Pair \"ledger\" \"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf\"","from":"Pair 1152132183 {}","diff_type":"update","value":"Pair 1153132183 {}"},{"prim":"bytes","type":"bytes","name":"Pair \"ledger\" \"tz1gVJBEoUQKGzwq6CCBpb8QjCQLejmJDSmU\"","from":"Pair 2521320 { Elt \"KT1DrJV8vhkdLEj76h1H9Q4irZDqAkMPo1Qf\" 1000000 }","diff_type":"update","value":"Pair 1521320 {}"}]},{"prim":"lambda","type":"lambda","name":"@lambda_6","value":"{ DUP ;\n  CAR ;\n  DIP { CDR } ;\n  DUP ;\n  DUP ;\n  CAR ;\n  DIP { CDR } ;\n  DIP { DIP 2 { DUP } ; DIG 2 } ;\n  PUSH string \"code\" ;\n  PAIR ;\n  PACK ;\n  GET ;\n  IF_NONE\n    { NONE (lambda (pair bytes (big_map bytes bytes))\n                   (pair (list operation) (big_map bytes bytes))) }\n    { UNPACK (lambda (pair bytes (big_map bytes bytes))\n                     (pair (list operation) (big_map bytes bytes))) ;\n      IF_NONE { PUSH string \"UStore: failed to unpack code\" ; FAILWITH } {} ;\n      SOME } ;\n  IF_NONE\n    { DROP ;\n      DIP { DUP ;\n            PUSH bytes 0x05010000000866616c6c6261636b ;\n            GET ;\n            IF_NONE { PUSH string \"UStore: no field fallback\" ; FAILWITH } {} ;\n            UNPACK (lambda (pair (pair string bytes) (big_map bytes bytes))\n                           (pair (list operation) (big_map bytes bytes))) ;\n            IF_NONE { PUSH string \"UStore: failed to unpack fallback\" ; FAILWITH } {} ;\n            SWAP } ;\n      PAIR ;\n      EXEC }\n    { DIP { SWAP ; DROP ; PAIR } ; SWAP ; EXEC } }"},{"prim":"nat","type":"nat","name":"@nat_21","value":"1"},{"prim":"bool","type":"bool","name":"@bool_22","value":false}]}`,
		}, {
			name: "edonet/KT1NeBYaiPiQH5jDWs3wJWzFynRsEGeUEUpZ/setList",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"nat"}],"annots":["%list1"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%map1"]}]},{"prim":"set","args":[{"prim":"nat"}],"annots":["%set1"]}]}`,
			curr: `{"args":[{"args":[[{"int":"1"}],[{"args":[{"int":"1"},{"int":"1"}],"prim":"Elt"},{"args":[{"int":"2"},{"int":"1"}],"prim":"Elt"}]],"prim":"Pair"},[{"int":"1"},{"int":"2"},{"int":"3"}]],"prim":"Pair"}`,
			prev: `{"args":[{"args":[[{"int":"1"}],[{"args":[{"int":"1"},{"int":"1"}],"prim":"Elt"},{"args":[{"int":"2"},{"int":"1"}],"prim":"Elt"}]],"prim":"Pair"},[{"int":"1"},{"int":"2"},{"int":"3"}]],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"list","type":"list","name":"list1","children":[{"prim":"nat","type":"nat","name":"@nat_4","value":"1"}]},{"prim":"map","type":"map","name":"map1","children":[{"prim":"mutez","type":"mutez","name":"1","value":"1"},{"prim":"mutez","type":"mutez","name":"2","value":"1"}]},{"prim":"set","type":"set","name":"set1","children":[{"prim":"nat","type":"nat","name":"@nat_9","value":"1"},{"prim":"nat","type":"nat","name":"@nat_9","value":"2"},{"prim":"nat","type":"nat","name":"@nat_9","value":"3"}]}]}`,
		}, {
			name: "mainnet/KT1RA48D7YPmS1bcpfhZKsN6DpZbC4oAxpVW/default",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"int"},{"prim":"map","args":[{"prim":"int"},{"prim":"int"}]}],"annots":["%deck"]},{"prim":"string","annots":["%doneWith"]}]},{"prim":"bool","annots":["%draw"]}]},{"prim":"int","annots":["%nbMoves"]}]},{"prim":"int","annots":["%winner"]}]}`,
			curr: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},[{"prim":"Elt","args":[{"int":"0"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]]},{"prim":"Elt","args":[{"int":"1"},[{"prim":"Elt","args":[{"int":"0"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"1"}]}]]},{"prim":"Elt","args":[{"int":"2"},[{"prim":"Elt","args":[{"int":"0"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"1"}]}]]}],{"string":"https://SmartPy.io"}]},{"prim":"True"}]},{"int":"9"}]},{"int":"0"}]}`,
			prev: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[[{"prim":"Elt","args":[{"int":"0"},[{"prim":"Elt","args":[{"int":"0"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"2"}]}]]},{"prim":"Elt","args":[{"int":"1"},[{"prim":"Elt","args":[{"int":"0"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"1"}]}]]},{"prim":"Elt","args":[{"int":"2"},[{"prim":"Elt","args":[{"int":"0"},{"int":"1"}]},{"prim":"Elt","args":[{"int":"1"},{"int":"2"}]},{"prim":"Elt","args":[{"int":"2"},{"int":"0"}]}]]}],{"string":"https://SmartPy.io"}]},{"prim":"False"}]},{"int":"8"}]},{"int":"0"}]}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"map","type":"map","name":"deck","children":[{"prim":"map","type":"map","name":"0","children":[{"prim":"int","type":"int","name":"0","value":"1"},{"prim":"int","type":"int","name":"1","value":"1"},{"prim":"int","type":"int","name":"2","value":"2"}]},{"prim":"map","type":"map","name":"1","children":[{"prim":"int","type":"int","name":"0","value":"2"},{"prim":"int","type":"int","name":"1","value":"2"},{"prim":"int","type":"int","name":"2","value":"1"}]},{"prim":"map","type":"map","name":"2","children":[{"prim":"int","type":"int","name":"0","value":"1"},{"prim":"int","type":"int","name":"1","value":"2"},{"prim":"int","type":"int","name":"2","from":"0","diff_type":"update","value":"1"}]}]},{"prim":"string","type":"string","name":"doneWith","value":"https://SmartPy.io"},{"prim":"bool","type":"bool","name":"draw","from":false,"diff_type":"update","value":true},{"prim":"int","type":"int","name":"nbMoves","from":"8","diff_type":"update","value":"9"},{"prim":"int","type":"int","name":"winner","value":"0"}]}`,
		}, {
			name: "delphinet/KT1BJfZP7E7e1QXFrEd4QjR53h3dEASfZezQ/init",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%kusd"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%long_bids"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%short_bids"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%option_contract"]},{"prim":"pair","args":[{"prim":"nat","annots":["%long_token_id"]},{"prim":"nat","annots":["%short_token_id"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%bidding_end"]},{"prim":"timestamp","annots":["%trading_end"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_long_bids"]},{"prim":"nat","annots":["%total_short_bids"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%long_price"]},{"prim":"nat","annots":["%short_price"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat","annots":[":winning_side"]}]},{"prim":"address","annots":["%harbinger"]}]},{"prim":"pair","args":[{"prim":"string","annots":["%asset"]},{"prim":"nat","annots":["%target_price"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"bool","annots":["%initialized"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%market_earnings"]},{"prim":"nat","annots":["%creator_earnings"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%cancel_fee"]},{"prim":"nat","annots":["%exercise_fee"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%min_capital"]},{"prim":"nat","annots":["%skew_limit"]}]}]}]}]}]}]}]}`,
			curr: `{"args":[{"args":[{"args":[{"bytes":"01a615eaab5e2e11a2e94f45e9b36e8e57de58729a00"},{"args":[{"int":"38079"},{"int":"38080"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"bytes":"012063789731c24ac7727b63715b49eee6ae17208100"},{"args":[{"int":"0"},{"int":"1"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"1609790087"},{"int":"1609795087"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"prim":"None"},{"bytes":"0182c57db6881913c4f8f31c002f9fbdd4b9b480d900"}],"prim":"Pair"},{"args":[{"string":"BTC-USD"},{"int":"25000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"00001506cf12081e7891fb42f347e17eae387083a0f1"},{"prim":"True"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"6000"},{"int":"3000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"int":"10000000"},{"int":"80"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			prev: `{"args":[{"args":[{"args":[{"bytes":"01a615eaab5e2e11a2e94f45e9b36e8e57de58729a00"},{"args":[{"int":"38079"},{"int":"38080"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"bytes":"01a615eaab5e2e11a2e94f45e9b36e8e57de58729a00"},{"args":[{"int":"0"},{"int":"1"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"1609790087"},{"int":"1609795087"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"prim":"None"},{"bytes":"0182c57db6881913c4f8f31c002f9fbdd4b9b480d900"}],"prim":"Pair"},{"args":[{"string":"BTC-USD"},{"int":"25000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"00001506cf12081e7891fb42f347e17eae387083a0f1"},{"prim":"False"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"6000"},{"int":"3000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"int":"10000000"},{"int":"80"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"address","type":"address","name":"kusd","value":"KT1PiwzQYNnMoy5GHMq4Si6o7zzDjsKgvACr"},{"prim":"big_map","type":"big_map","name":"long_bids","value":38079},{"prim":"big_map","type":"big_map","name":"short_bids","value":38080},{"prim":"address","type":"address","name":"option_contract","from":"KT1PiwzQYNnMoy5GHMq4Si6o7zzDjsKgvACr","diff_type":"update","value":"KT1BY2MGrcpwLBUvYnDcPVhzsqGXXZfR5fXg"},{"prim":"nat","type":"nat","name":"long_token_id","value":"0"},{"prim":"nat","type":"nat","name":"short_token_id","value":"1"},{"prim":"timestamp","type":"timestamp","name":"bidding_end","value":"2021-01-04T19:54:47Z"},{"prim":"timestamp","type":"timestamp","name":"trading_end","value":"2021-01-04T21:18:07Z"},{"prim":"nat","type":"nat","name":"total_long_bids","value":"0"},{"prim":"nat","type":"nat","name":"total_short_bids","value":"0"},{"prim":"nat","type":"nat","name":"long_price","value":"0"},{"prim":"nat","type":"nat","name":"short_price","value":"0"},{"name":"winning_side","value":"None"},{"prim":"address","type":"address","name":"harbinger","value":"KT1LWDzd6mFhjjnb65a1PjHDNZtFKBieTQKH"},{"prim":"string","type":"string","name":"asset","value":"BTC-USD"},{"prim":"nat","type":"nat","name":"target_price","value":"25000"},{"prim":"address","type":"address","name":"admin","value":"tz1MZD3EecfFVHbteFYXZMpFnvFH1g6a2BA1"},{"prim":"bool","type":"bool","name":"initialized","from":false,"diff_type":"update","value":true},{"prim":"nat","type":"nat","name":"market_earnings","value":"0"},{"prim":"nat","type":"nat","name":"creator_earnings","value":"0"},{"prim":"nat","type":"nat","name":"cancel_fee","value":"6000"},{"prim":"nat","type":"nat","name":"exercise_fee","value":"3000"},{"prim":"nat","type":"nat","name":"min_capital","value":"10000000"},{"prim":"nat","type":"nat","name":"skew_limit","value":"80"}]}`,
		}, {
			name: "mainnet/KT1Puc9St8wdNoGtLiD2WXaHbWU7styaxYhD/addLiquidity",
			tree: `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":owner"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":allowance"]}]}]}],"annots":["%accounts"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bool","annots":[":selfIsUpdatingTokenPool"]},{"prim":"pair","args":[{"prim":"bool","annots":[":freezeBaker"]},{"prim":"nat","annots":[":lqtTotal"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":manager"]},{"prim":"address","annots":[":tokenAddress"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokenPool"]},{"prim":"mutez","annots":[":xtzPool"]}]}]}]}]}`,
			curr: `{"args":[[{"args":[{"string":"tz1iMfjNhpcUfDNCFTY5j2S3J6cY9NU8fKuv"},{"args":[{"int":"4919489"},[]],"prim":"Pair"}],"prim":"Elt"}],{"args":[{"args":[{"prim":"False"},{"args":[{"prim":"False"},{"int":"55565098589"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"011b5edfd90cc62eaf4299bdc2eedae4306bea9fe200"},{"bytes":"01813a19aa1cb96fe5b039c0bfc1633e61daf0ae3f00"}],"prim":"Pair"},{"args":[{"int":"169423316128"},{"int":"42961629151"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			prev: `{"args":[{"int":"124"},{"args":[{"args":[{"prim":"False"},{"args":[{"prim":"False"},{"int":"55560179100"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"011b5edfd90cc62eaf4299bdc2eedae4306bea9fe200"},{"bytes":"01813a19aa1cb96fe5b039c0bfc1633e61daf0ae3f00"}],"prim":"Pair"},{"args":[{"int":"169408316131"},{"int":"42957825517"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"big_map","type":"big_map","name":"accounts","children":[{"prim":"pair","type":"namedtuple","name":"tz1iMfjNhpcUfDNCFTY5j2S3J6cY9NU8fKuv","diff_type":"create","children":[{"prim":"nat","type":"nat","name":"balance","diff_type":"create","value":"4919489"},{"prim":"map","type":"map","name":"@map_6","diff_type":"create"}]}]},{"prim":"bool","type":"bool","name":"selfIsUpdatingTokenPool","value":false},{"prim":"bool","type":"bool","name":"freezeBaker","value":false},{"prim":"nat","type":"nat","name":"lqtTotal","from":"55560179100","diff_type":"update","value":"55565098589"},{"prim":"address","type":"address","name":"manager","value":"KT1B5VTw8ZSMnrjhy337CEvAm4tnT8Gu8Geu"},{"prim":"address","type":"address","name":"tokenAddress","value":"KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9"},{"prim":"nat","type":"nat","name":"tokenPool","from":"169408316131","diff_type":"update","value":"169423316128"},{"prim":"mutez","type":"mutez","name":"xtzPool","from":"42957825517","diff_type":"update","value":"42961629151"}]}`,
		}, {
			name: "delphinet/KT1NogASEVoWnYvbjRPbPuTK58Z3ttB8pRDC/accept_ownership",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%default_expiry"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%minting_allowances"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"address"}]},{"prim":"unit"}],"annots":["%operators"]},{"prim":"bool","annots":["%paused"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%permit_counter"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}],"annots":["%expiry"]},{"prim":"map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"timestamp","annots":["%created_at"]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%expiry"]}]}],"annots":["%permits"]}]}],"annots":["%permits"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%master_minter"]},{"prim":"address","annots":["%owner"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%pauser"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%pending_owner"]}]}],"annots":["%roles"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"option","args":[{"prim":"address"}],"annots":["%transferlist_contract"]}]}]}`,
			curr: `{"args":[{"args":[{"args":[{"args":[{"int":"1000"},{"int":"44121"}],"prim":"Pair"},{"args":[{"int":"44122"},[]],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"44123"},{"prim":"False"}],"prim":"Pair"},{"args":[{"int":"11"},[{"args":[{"string":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"args":[{"prim":"None"},[]],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"}],"prim":"Pair"},{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"prim":"None"}],"prim":"Pair"}],"prim":"Pair"},{"int":"1000"}],"prim":"Pair"},{"prim":"None"}],"prim":"Pair"}],"prim":"Pair"}`,
			prev: `{"args":[{"args":[{"args":[{"args":[{"int":"1000"},{"int":"44121"}],"prim":"Pair"},{"args":[{"int":"44122"},[]],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"44123"},{"prim":"False"}],"prim":"Pair"},{"args":[{"int":"11"},[{"args":[{"string":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"args":[{"prim":"None"},[{"args":[{"string":"24b880dafed29ed33e79156e4aecc759927b73439afb91c25663333cc45758d1"},{"args":[{"int":"1610679129"},{"prim":"None"}],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"bytes":"0000f2037cc89912cc457e3d85cfe2bfc471b4b6022d"}],"prim":"Pair"},{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"}],"prim":"Some"}],"prim":"Pair"}],"prim":"Pair"},{"int":"1000"}],"prim":"Pair"},{"prim":"None"}],"prim":"Pair"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"nat","type":"nat","name":"default_expiry","value":"1000"},{"prim":"big_map","type":"big_map","name":"ledger","value":44121},{"prim":"big_map","type":"big_map","name":"metadata","value":44122},{"prim":"map","type":"map","name":"minting_allowances"},{"prim":"big_map","type":"big_map","name":"operators","value":44123},{"prim":"bool","type":"bool","name":"paused","value":false},{"prim":"nat","type":"nat","name":"permit_counter","value":"11"},{"prim":"big_map","type":"big_map","name":"permits","children":[{"prim":"pair","type":"namedtuple","name":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1","children":[{"name":"expiry","value":"None"},{"prim":"map","type":"map","name":"permits","children":[{"prim":"pair","type":"namedtuple","name":"24b880dafed29ed33e79156e4aecc759927b73439afb91c25663333cc45758d1","diff_type":"delete","children":[{"prim":"timestamp","type":"timestamp","name":"created_at","diff_type":"delete","value":"2021-01-15T02:52:09Z"},{"prim":"option","type":"option","name":"expiry","diff_type":"delete","value":"None"}]}]}]}]},{"prim":"pair","type":"namedtuple","name":"roles","children":[{"prim":"address","type":"address","name":"master_minter","value":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"prim":"address","type":"address","name":"owner","from":"tz1hhgPrTFWzXLRyqfWRBDEK5WXFH3FgpFhG","diff_type":"update","value":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"prim":"address","type":"address","name":"pauser","value":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"prim":"address","type":"address","name":"pending_owner","diff_type":"delete","value":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"}]},{"prim":"nat","type":"nat","name":"total_supply","value":"1000"},{"name":"transferlist_contract","value":"None"}]}`,
		}, {
			name: "edonet/KT1QHduqUZi4HRaFRaNAetWZmYTQprugQ2AS/default",
			tree: `{"prim":"option","args":[{"prim":"bls12_381_fr"}]}`,
			curr: `{"args":[{"bytes":"1e00000000000000000000000000000000000000000000000000000000000000"}],"prim":"Some"}`,
			prev: `{"args":[{"bytes":"02e0010000000000000000000000000000000000000000000000000000000000"}],"prim":"Some"}`,
			want: `{"prim":"bls12_381_fr","type":"bls12_381_fr","name":"@bls12_381_fr_2","from":"02e0010000000000000000000000000000000000000000000000000000000000","diff_type":"update","value":"1e00000000000000000000000000000000000000000000000000000000000000"}`,
		}, {
			name: "edonet/KT1Gug8pjcs9qi9VpMoU3E4gjq8fiaT9ZJPd/default",
			tree: `{"prim":"option","args":[{"prim":"bls12_381_g1"}]}`,
			curr: `{"args":[{"bytes":"400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}],"prim":"Some"}`,
			prev: `{"prim":"None"}`,
			want: `{"prim":"bls12_381_g1","type":"bls12_381_g1","name":"@bls12_381_g1_2","diff_type":"create","value":"400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}`,
		}, {
			name: "edonet/KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ/receive",
			tree: `{"prim":"pair","args":[{"prim":"address","annots":["%manager"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"ticket","args":[{"prim":"unit"}]}],"annots":["%tickets"]}]}`,
			curr: `{"args":[{"bytes":"0000d00811680a0689cbf5d73126943268cba8284fd2"},[{"args":[{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"},{"args":[{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"},{"args":[{"prim":"Unit"},{"int":"1"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}`,
			prev: `{"args":[{"bytes":"0000d00811680a0689cbf5d73126943268cba8284fd2"},{"int":"40"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"address","type":"address","name":"manager","value":"tz1ebzubQKGg5AJ2z9Ydun9HzLLy4AzngZq6"},{"prim":"big_map","type":"big_map","name":"tickets","children":[{"prim":"ticket","type":"ticket","name":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x","diff_type":"create","children":[{"prim":"pair","type":"namedtuple","name":"@pair_7","diff_type":"create","children":[{"prim":"address","type":"address","name":"@address_8","diff_type":"create","value":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"},{"prim":"unit","type":"unit","name":"@unit_6","diff_type":"create"},{"prim":"nat","type":"nat","name":"@nat_10","diff_type":"create","value":"1"}]}]}]}]}`,
		}, {
			name: "delphinet/KT1HWqT2YzysjbAnp8iJeF6tgQuMWydtMUgH/setConfirmedContract",
			tree: `{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"bool","annots":["%confirmed"]},{"prim":"pair","args":[{"prim":"nat","annots":["%grade"]},{"prim":"signature","annots":["%migration_sig"]}]}]}],"annots":["%confirmedAddress"]}]}`,
			curr: `{"args":[{"bytes":"000044e6d7af260d4532310ae91cc5643afc275c0a9f"},[{"args":[{"string":"KT19P32o9Mb2QYf6y8ZaDSHJZ2PpKYbrDx33"},{"args":[{"prim":"True"},{"args":[{"int":"1"},{"string":"cf5e8ab9df8d9ef9fb8f19e3312f56ab277017367ea2a8ac99946b4d11f449583a701d35c746566ca31a37c1d2ca460c15886307385bb68519b2df98275dfc04"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}`,
			prev: `{"args":[{"bytes":"000044e6d7af260d4532310ae91cc5643afc275c0a9f"},{"int":"30103"}],"prim":"Pair"}`,
			want: `{"prim":"pair","type":"namedtuple","name":"@pair_1","children":[{"prim":"address","type":"address","name":"admin","value":"tz1RvMBPAGJWewztAivG59ZNEG6FSsfRFwoL"},{"prim":"big_map","type":"big_map","name":"confirmedAddress","children":[{"prim":"pair","type":"namedtuple","name":"KT19P32o9Mb2QYf6y8ZaDSHJZ2PpKYbrDx33","diff_type":"create","children":[{"prim":"bool","type":"bool","name":"confirmed","diff_type":"create","value":true},{"prim":"nat","type":"nat","name":"grade","diff_type":"create","value":"1"},{"prim":"signature","type":"signature","name":"migration_sig","diff_type":"create","value":"cf5e8ab9df8d9ef9fb8f19e3312f56ab277017367ea2a8ac99946b4d11f449583a701d35c746566ca31a37c1d2ca460c15886307385bb68519b2df98275dfc04"}]}]}]}`,
		}, {
			name: "or test",
			tree: `{"prim":"or","args":[{"prim":"int"},{"prim":"string"}]}`,
			curr: `{"prim":"Left","args":[{"int":"10"}]}`,
			prev: `{"prim":"Right","args":[{"string":"test"}]}`,
			want: `{"prim":"or","type":"or","name":"@or_1","diff_type":"update","children":[{"prim":"string","type":"string","name":"@string_3","diff_type":"update","value":"test"}]}`,
		}, {
			name: "or test",
			tree: `{"prim":"or","args":[{"prim":"int"},{"prim":"string"}]}`,
			curr: `{"prim":"Left","args":[{"int":"10"}]}`,
			prev: `{"prim":"Left","args":[{"int":"0"}]}`,
			want: `{"prim":"or","type":"or","name":"@or_1","children":[{"prim":"int","type":"int","name":"@int_2","from":"0","diff_type":"update","value":"10"}]}`,
		}, {
			name: "or test",
			tree: `{"prim":"or","args":[{"prim":"int"},{"prim":"string"}]}`,
			curr: `{"prim":"Right","args":[{"string":"test"}]}`,
			prev: `{"prim":"Left","args":[{"int":"0"}]}`,
			want: `{"prim":"or","type":"or","name":"@or_1","diff_type":"update","children":[{"prim":"int","type":"int","name":"@int_2","diff_type":"update","value":"0"}]}`,
		}, {
			name: "or test",
			tree: `{"prim":"or","args":[{"prim":"int"},{"prim":"string"}]}`,
			curr: `{"prim":"Right","args":[{"string":"test2"}]}`,
			prev: `{"prim":"Right","args":[{"string":"test"}]}`,
			want: `{"prim":"or","type":"or","name":"@or_1","children":[{"prim":"string","type":"string","name":"@string_3","from":"test","diff_type":"update","value":"test2"}]}`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default",
			tree: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			curr: `[{"prim":"Pair","args":[{"bytes":"000002"},{"prim":"Some","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]`,
			prev: `[{"prim":"Pair","args":[{"bytes":"100002"},{"prim":"Some","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]`,
			want: `{"prim":"list","type":"list","name":"@list_1","children":[{"prim":"pair","type":"namedtuple","name":"@pair_2","children":[{"prim":"sapling_transaction","type":"sapling_transaction","name":"@sapling_transaction_3","from":"100002","diff_type":"update","value":"000002"},{"prim":"key_hash","type":"key_hash","name":"@key_hash_5","value":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default lis insert",
			tree: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			curr: `[{"prim":"Pair","args":[{"bytes":"000002"},{"prim":"Some","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]`,
			prev: `[]`,
			want: `{"prim":"list","type":"list","name":"@list_1","children":[{"prim":"pair","type":"namedtuple","name":"@pair_2","diff_type":"create","children":[{"prim":"sapling_transaction","type":"sapling_transaction","name":"@sapling_transaction_3","diff_type":"create","value":"000002"},{"prim":"key_hash","type":"key_hash","name":"@key_hash_5","diff_type":"create","value":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treeA UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeA); err != nil {
				t.Errorf("UnmarshalFromString treeA error = %v", err)
				return
			}
			typA, err := treeA.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			var dataA UntypedAST
			if err := json.UnmarshalFromString(tt.curr, &dataA); err != nil {
				t.Errorf("UnmarshalFromString dataA error = %v", err)
				return
			}
			if err := typA.Settle(dataA); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Settle(a) error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var treeB UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeB); err != nil {
				t.Errorf("UnmarshalFromString treeB error = %v", err)
				return
			}
			typB, err := treeB.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(b) error = %v", err)
				return
			}
			var dataB UntypedAST
			if err := json.UnmarshalFromString(tt.prev, &dataB); err != nil {
				t.Errorf("UnmarshalFromString dataB error = %v", err)
				return
			}
			if err := typB.Settle(dataB); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Settle(b) error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := typA.Diff(typB)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, err := json.Marshal(got)
			if err != nil {
				t.Errorf("Marshal(got) data error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestTypedAst_EnrichBigMap(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		data    string
		bmd     []*types.BigMapDiff
		want    string
		wantErr bool
	}{
		{
			name: "simple",
			tree: `{"prim": "big_map","args":[{"prim":"int"},{"prim":"string"}]}`,
			data: `{"int": "100"}`,
			bmd: []*types.BigMapDiff{
				{
					Ptr:   100,
					Key:   []byte(`{"int":"10"}`),
					Value: []byte(`{"string":"test"}`),
				}, {
					Ptr:   100,
					Key:   []byte(`{"int":"9"}`),
					Value: []byte(`{"string":"test2"}`),
				}, {
					Ptr:   101,
					Key:   []byte(`{"int":"4"}`),
					Value: []byte(`{"string":"test3"}`),
				},
			},
			want: `[{"prim":"Elt","args":[{"int":"9"},{"string":"test2"}]},{"prim":"Elt","args":[{"int":"10"},{"string":"test"}]}]`,
		}, {
			name: "big_map in list",
			tree: `{"prim":"list","args":[{"prim": "big_map","args":[{"prim":"int"},{"prim":"string"}]}]}`,
			data: `[{"int":"100"},{"int":"200"}]`,
			bmd: []*types.BigMapDiff{
				{
					Ptr:   100,
					Key:   []byte(`{"int":"10"}`),
					Value: []byte(`{"string":"test"}`),
				}, {
					Ptr:   100,
					Key:   []byte(`{"int":"9"}`),
					Value: []byte(`{"string":"test2"}`),
				}, {
					Ptr:   101,
					Key:   []byte(`{"int":"4"}`),
					Value: []byte(`{"string":"test3"}`),
				},
			},
			want: `[[{"prim":"Elt","args":[{"int":"9"},{"string":"test2"}]},{"prim":"Elt","args":[{"int":"10"},{"string":"test"}]}],[]]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treeA UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeA); err != nil {
				t.Errorf("UnmarshalFromString treeA error = %v", err)
				return
			}
			typA, err := treeA.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			var dataA UntypedAST
			if err := json.UnmarshalFromString(tt.data, &dataA); err != nil {
				t.Errorf("UnmarshalFromString dataA error = %v", err)
				return
			}
			if err := typA.Settle(dataA); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Settle(a) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := typA.EnrichBigMap(tt.bmd); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.EnrichBigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, err := typA.ToParameters("")
			if err != nil {
				t.Errorf("ToParameters(a) error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestTypedAst_ToBaseNode(t *testing.T) {
	tests := []struct {
		name      string
		tree      string
		data      string
		optimized bool
		wantErr   bool
	}{
		{
			name:      "simple big map",
			tree:      `{"prim":"option", "args":[{"prim": "big_map","args":[{"prim":"int"},{"prim":"string"}]}]}`,
			data:      `{"prim":"Some","args":[[{"prim":"Elt","args":[{"int":"9"},{"string":"test2"}]},{"prim":"Elt","args":[{"int":"10"},{"string":"test"}]}]]}`,
			optimized: true,
		}, {
			name:      "atomex initiate",
			tree:      `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			data:      `{"args":[[{"args":[{"string":"314d74100a3d994ae0b52114fb414d86001a0e4a1b122ef1d03ed5ad6b7c4f93"},{"args":[{"args":[{"string":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"},{"string":"tz1U2QDpyre9fqCWXG7MwTE38cR5d7qWoYgC"}],"prim":"Pair"},{"args":[{"args":[{"int":"192258291"},{"string":"2020-10-13T12:05:21Z"}],"prim":"Pair"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}],{"prim":"Unit"}],"prim":"Pair"}`,
			optimized: false,
		}, {
			name:      "edonet/KT1NeBYaiPiQH5jDWs3wJWzFynRsEGeUEUpZ/setList",
			tree:      `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"nat"}],"annots":["%list1"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%map1"]}]},{"prim":"set","args":[{"prim":"nat"}],"annots":["%set1"]}]}`,
			data:      `{"args":[{"args":[[{"int":"1"}],[{"args":[{"int":"1"},{"int":"1"}],"prim":"Elt"},{"args":[{"int":"2"},{"int":"1"}],"prim":"Elt"}]],"prim":"Pair"},[{"int":"1"},{"int":"2"},{"int":"3"}]],"prim":"Pair"}`,
			optimized: true,
		}, {
			name:      "delphinet/KT1BJfZP7E7e1QXFrEd4QjR53h3dEASfZezQ/init",
			tree:      `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%kusd"]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%long_bids"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%short_bids"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%option_contract"]},{"prim":"pair","args":[{"prim":"nat","annots":["%long_token_id"]},{"prim":"nat","annots":["%short_token_id"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%bidding_end"]},{"prim":"timestamp","annots":["%trading_end"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_long_bids"]},{"prim":"nat","annots":["%total_short_bids"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%long_price"]},{"prim":"nat","annots":["%short_price"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat","annots":[":winning_side"]}]},{"prim":"address","annots":["%harbinger"]}]},{"prim":"pair","args":[{"prim":"string","annots":["%asset"]},{"prim":"nat","annots":["%target_price"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"bool","annots":["%initialized"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%market_earnings"]},{"prim":"nat","annots":["%creator_earnings"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%cancel_fee"]},{"prim":"nat","annots":["%exercise_fee"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%min_capital"]},{"prim":"nat","annots":["%skew_limit"]}]}]}]}]}]}]}]}`,
			data:      `{"args":[{"args":[{"args":[{"bytes":"01a615eaab5e2e11a2e94f45e9b36e8e57de58729a00"},{"args":[{"int":"38079"},{"int":"38080"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"bytes":"012063789731c24ac7727b63715b49eee6ae17208100"},{"args":[{"int":"0"},{"int":"1"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"1609790087"},{"int":"1609795087"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"prim":"None"},{"bytes":"0182c57db6881913c4f8f31c002f9fbdd4b9b480d900"}],"prim":"Pair"},{"args":[{"string":"BTC-USD"},{"int":"25000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"00001506cf12081e7891fb42f347e17eae387083a0f1"},{"prim":"True"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"int":"0"},{"int":"0"}],"prim":"Pair"},{"args":[{"int":"6000"},{"int":"3000"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"int":"10000000"},{"int":"80"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			optimized: true,
		}, {
			name:      "mainnet/KT1Puc9St8wdNoGtLiD2WXaHbWU7styaxYhD/addLiquidity",
			tree:      `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":owner"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":allowance"]}]}]}],"annots":["%accounts"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bool","annots":[":selfIsUpdatingTokenPool"]},{"prim":"pair","args":[{"prim":"bool","annots":[":freezeBaker"]},{"prim":"nat","annots":[":lqtTotal"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":manager"]},{"prim":"address","annots":[":tokenAddress"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokenPool"]},{"prim":"mutez","annots":[":xtzPool"]}]}]}]}]}`,
			data:      `{"args":[[{"args":[{"string":"tz1iMfjNhpcUfDNCFTY5j2S3J6cY9NU8fKuv"},{"args":[{"int":"4919489"},[]],"prim":"Pair"}],"prim":"Elt"}],{"args":[{"args":[{"prim":"False"},{"args":[{"prim":"False"},{"int":"55565098589"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"bytes":"011b5edfd90cc62eaf4299bdc2eedae4306bea9fe200"},{"bytes":"01813a19aa1cb96fe5b039c0bfc1633e61daf0ae3f00"}],"prim":"Pair"},{"args":[{"int":"169423316128"},{"int":"42961629151"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"}`,
			optimized: false,
		}, {
			name:      "delphinet/KT1NogASEVoWnYvbjRPbPuTK58Z3ttB8pRDC/accept_ownership",
			tree:      `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%default_expiry"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%ledger"]}]},{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%metadata"]},{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}],"annots":["%minting_allowances"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"address"}]},{"prim":"unit"}],"annots":["%operators"]},{"prim":"bool","annots":["%paused"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%permit_counter"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}],"annots":["%expiry"]},{"prim":"map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"timestamp","annots":["%created_at"]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%expiry"]}]}],"annots":["%permits"]}]}],"annots":["%permits"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%master_minter"]},{"prim":"address","annots":["%owner"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%pauser"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%pending_owner"]}]}],"annots":["%roles"]},{"prim":"nat","annots":["%total_supply"]}]},{"prim":"option","args":[{"prim":"address"}],"annots":["%transferlist_contract"]}]}]}`,
			data:      `{"args":[{"args":[{"args":[{"args":[{"int":"1000"},{"int":"44121"}],"prim":"Pair"},{"args":[{"int":"44122"},[]],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"int":"44123"},{"prim":"False"}],"prim":"Pair"},{"args":[{"int":"11"},[{"args":[{"string":"tz1PuYDXAkEyNo1oyuBqHSD4edXXqBvWt4A1"},{"args":[{"prim":"None"},[]],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}],"prim":"Pair"}],"prim":"Pair"},{"args":[{"args":[{"args":[{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"}],"prim":"Pair"},{"args":[{"bytes":"00002ecf635a4b1508e6effe97aea69dcd99c1113dc0"},{"prim":"None"}],"prim":"Pair"}],"prim":"Pair"},{"int":"1000"}],"prim":"Pair"},{"prim":"None"}],"prim":"Pair"}],"prim":"Pair"}`,
			optimized: false,
		}, {
			name:      "edonet/KT1QHduqUZi4HRaFRaNAetWZmYTQprugQ2AS/default",
			tree:      `{"prim":"option","args":[{"prim":"bls12_381_fr"}]}`,
			data:      `{"args":[{"bytes":"1e00000000000000000000000000000000000000000000000000000000000000"}],"prim":"Some"}`,
			optimized: true,
		}, {
			name:      "edonet/KT1Gug8pjcs9qi9VpMoU3E4gjq8fiaT9ZJPd/default",
			tree:      `{"prim":"option","args":[{"prim":"bls12_381_g1"}]}`,
			data:      `{"args":[{"bytes":"400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}],"prim":"Some"}`,
			optimized: true,
		}, {
			name:      "edonet/KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ/receive",
			tree:      `{"prim":"pair","args":[{"prim":"address","annots":["%manager"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"ticket","args":[{"prim":"unit"}]}],"annots":["%tickets"]}]}`,
			data:      `{"args":[{"bytes":"0000d00811680a0689cbf5d73126943268cba8284fd2"},[{"args":[{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"},{"args":[{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"},{"args":[{"prim":"Unit"},{"int":"1"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}`,
			optimized: false,
		}, {
			name:      "delphinet/KT1HWqT2YzysjbAnp8iJeF6tgQuMWydtMUgH/setConfirmedContract",
			tree:      `{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"bool","annots":["%confirmed"]},{"prim":"pair","args":[{"prim":"nat","annots":["%grade"]},{"prim":"signature","annots":["%migration_sig"]}]}]}],"annots":["%confirmedAddress"]}]}`,
			data:      `{"args":[{"bytes":"000044e6d7af260d4532310ae91cc5643afc275c0a9f"},[{"args":[{"string":"KT19P32o9Mb2QYf6y8ZaDSHJZ2PpKYbrDx33"},{"args":[{"prim":"True"},{"args":[{"int":"1"},{"string":"cf5e8ab9df8d9ef9fb8f19e3312f56ab277017367ea2a8ac99946b4d11f449583a701d35c746566ca31a37c1d2ca460c15886307385bb68519b2df98275dfc04"}],"prim":"Pair"}],"prim":"Pair"}],"prim":"Elt"}]],"prim":"Pair"}`,
			optimized: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treeA UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeA); err != nil {
				t.Errorf("UnmarshalFromString treeA error = %v", err)
				return
			}
			a, err := treeA.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			var dataA UntypedAST
			if err := json.UnmarshalFromString(tt.data, &dataA); err != nil {
				t.Errorf("UnmarshalFromString dataA error = %v", err)
				return
			}
			if err := a.Settle(dataA); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.Settle(a) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := a.ToBaseNode(tt.optimized)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.ToBaseNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			bGot, err := json.Marshal(got)
			if err != nil {
				t.Errorf("Marshal(got) error = %v", err)
				return
			}
			var mGot map[string]interface{}
			if err := json.Unmarshal(bGot, &mGot); err != nil {
				t.Errorf("Unmarshal(got) error = %v", err)
				return
			}
			var mWant map[string]interface{}
			if err := json.UnmarshalFromString(tt.data, &mWant); err != nil {
				t.Errorf("UnmarshalFromString(want) error = %v", err)
				return
			}
			assert.Equal(t, mWant, mGot)
		})
	}
}

func TestTypedAst_FromJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name: "atomex initiate",
			tree: `{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]}`,
			data: `{"participant": "tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a", "hashed_secret": "314d74100a3d994ae0b52114fb414d86001a0e4a1b122ef1d03ed5ad6b7c4f93", "refund_time": "2020-10-13T12:05:21Z", "payoff": 800}`,
			want: `{"prim":"Pair","args":[{"string":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"314d74100a3d994ae0b52114fb414d86001a0e4a1b122ef1d03ed5ad6b7c4f93"},{"int":"1602590721"}]},{"int":"800"}]}]}`,
		}, {
			name: "delphinet/KT1VZj8kJYcpqr3fAPC8sLjSNuDvYDz5V39Z/mainParameter",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}],"annots":["%operation"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"mutez"}],"annots":["%transferTokens"]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"key"}]}],"annots":["%changeKeys"]}]}]}]},{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%mainParameter"]}`,
			data: `{"@nat_3":10,"@or_4":{"@nat_6":0,"schemaKey":"L","@address_7":"KT1VZj8kJYcpqr3fAPC8sLjSNuDvYDz5V39Z"},"@list_16":[{"@option_17":{"schemaKey":"some","@signature_18":"d218139a182d8932bf99ffcff72074f6451fe0126e74c43faee7ce65e6dd5b2dcce959d3c6f44c0820adabecfb269e4366b540c5841491bedd31d76e892de30f"}},{"@option_17":{"schemaKey":"none"}}]}`,
			want: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"10"},{"prim":"Left","args":[{"prim":"Pair","args":[{"int":"0"},{"string":"KT1VZj8kJYcpqr3fAPC8sLjSNuDvYDz5V39Z"}]}]}]},[{"prim":"Some","args":[{"string":"sigqUVuz98T1vsmQzGv46JKcShb5ups1pZvHTLLptnPToEjjLH6rzdYZX3Y6V1PpYbDFtW3miegcAGrEnBSka1gHhyCqCvmW"}]},{"prim":"None"}]]}`,
		}, {
			name: "mainnet/KT1Ty2uAmF5JxWyeGrVpk17MEyzVB8cXs8aJ/deposit",
			tree: `{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"bool"},{"prim":"pair","args":[{"prim":"int"},{"prim":"int"}]}]}]}]}]}],"annots":["%deposit"]}`,
			data: `{"@nat_6":10,"@nat_8":20,"@bool_10":true,"@int_12":30,"@int_13":40,"@address_2":"KT1Ty2uAmF5JxWyeGrVpk17MEyzVB8cXs8aJ","@address_4":"KT1Ty2uAmF5JxWyeGrVpk17MEyzVB8cXs8aJ"}`,
			want: `{"prim":"Pair","args":[{"string":"KT1Ty2uAmF5JxWyeGrVpk17MEyzVB8cXs8aJ"},{"prim":"Pair","args":[{"string":"KT1Ty2uAmF5JxWyeGrVpk17MEyzVB8cXs8aJ"},{"prim":"Pair","args":[{"int":"10"},{"prim":"Pair","args":[{"int":"20"},{"prim":"Pair","args":[{"prim":"True"},{"prim":"Pair","args":[{"int":"30"},{"int":"40"}]}]}]}]}]}]}`,
		}, {
			name: "tzbtc epwApplyMigration",
			tree: `{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationscript","%epwApplyMigration"]}`,
			data: `{"migrationscript":"{CDR;CAR}"}`,
			want: `[[{"prim":"CDR"},{"prim":"CAR"}]]`,
		}, {
			name: "delphinet/KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r/update_record",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%address"]},{"prim":"bool","annots":["%bool"]}]},{"prim":"or","args":[{"prim":"bytes","annots":["%bytes"]},{"prim":"int","annots":["%int"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"key","annots":["%key"]},{"prim":"key_hash","annots":["%key_hash"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%nat"]},{"prim":"signature","annots":["%signature"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"string","annots":["%string"]},{"prim":"mutez","annots":["%tez"]}]},{"prim":"timestamp","annots":["%timestamp"]}]}]}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"bytes","annots":["%name"]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%ttl"]}],"annots":["%update_record"]}`,
			data: `{"ttl":{"schemaKey":"none"},"data":[{"@or_8":{"bool":true,"schemaKey":"LLLR"},"@string_7":"test"}],"address":{"schemaKey":"none"},"name":"00","owner":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}`,
			want: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"None"},[{"prim":"Elt","args":[{"string":"test"},{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"True"}]}]}]}]}]}]]},{"prim":"Pair","args":[{"bytes":"00"},{"string":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}]}]},{"prim":"None"}]}`,
		}, {
			name: "delphinet/KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r/update_record modified with big_map",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%address"]},{"prim":"bool","annots":["%bool"]}]},{"prim":"or","args":[{"prim":"bytes","annots":["%bytes"]},{"prim":"int","annots":["%int"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"key","annots":["%key"]},{"prim":"key_hash","annots":["%key_hash"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%nat"]},{"prim":"signature","annots":["%signature"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"string","annots":["%string"]},{"prim":"mutez","annots":["%tez"]}]},{"prim":"timestamp","annots":["%timestamp"]}]}]}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"bytes","annots":["%name"]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%ttl"]}],"annots":["%update_record"]}`,
			data: `{"ttl":{"schemaKey":"none"},"data":[{"@or_8":{"bool":true,"schemaKey":"LLLR"},"@string_7":"test"}],"address":{"schemaKey":"none"},"name":"00","owner":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}`,
			want: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"None"},[{"prim":"Elt","args":[{"string":"test"},{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"True"}]}]}]}]}]}]]},{"prim":"Pair","args":[{"bytes":"00"},{"string":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}]}]},{"prim":"None"}]}`,
		}, {
			name: "delphinet/KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r/update_record",
			tree: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}],"annots":["%address"]},{"prim":"map","args":[{"prim":"string"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%address"]},{"prim":"bool","annots":["%bool"]}]},{"prim":"or","args":[{"prim":"bytes","annots":["%bytes"]},{"prim":"int","annots":["%int"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"key","annots":["%key"]},{"prim":"key_hash","annots":["%key_hash"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%nat"]},{"prim":"signature","annots":["%signature"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"string","annots":["%string"]},{"prim":"mutez","annots":["%tez"]}]},{"prim":"timestamp","annots":["%timestamp"]}]}]}],"annots":["%data"]}]},{"prim":"pair","args":[{"prim":"bytes","annots":["%name"]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"option","args":[{"prim":"nat"}],"annots":["%ttl"]}],"annots":["%update_record"]}`,
			data: `{"address":{"schemaKey":"some","@address_5":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"},"data":[{"@or_8":{"schemaKey":"LRLR","key_hash":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"},"@string_7":"test"}],"ttl":{"schemaKey":"some","@nat_33":100},"name":"00","owner":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}`,
			want: `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}]},[{"prim":"Elt","args":[{"string":"test"},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]}]}]}]]},{"prim":"Pair","args":[{"bytes":"00"},{"string":"KT1GM8hwKseJjArkcuw99QyBBC9dswxXr66r"}]}]},{"prim":"Some","args":[{"int":"100"}]}]}`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default",
			tree: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			data: `{"@list_1":[{"@option_4":{"schemaKey":"some","@key_hash_5":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}, "@sapling_transaction_3":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"}]}`,
			want: `[{"prim":"Pair","args":[{"bytes":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"Some","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]`,
		}, {
			name: "edonet/KT1KntQCDntGvRpLdjykNXJBHJs8rLU9WaBV/default modified set",
			tree: `{"prim":"set","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}`,
			data: `{"@set_1":[{"@option_4":{"schemaKey":"some","@key_hash_5":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}, "@sapling_transaction_3":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"}]}`,
			want: `[{"prim":"Pair","args":[{"bytes":"000002c0db3146a94b750c32d29554da38676439454600485d341403d1e1360732b9dd5baa2ff48826cd9f8d090b01a94c9cef44c0a43fd40a599bc32f6c52d7e6925dd1eb059a1bb5a190242986f9d29962c90fd6b885eac540e49b21cd8b2ef165522d98f55950e232608fda87cbacc6768bb3497985df91412ddc45fd9264d5580e859ecf5e2238761754dd4963bdad0c068cb3f44e10fdfbccf27075024a1c36bf449ae03d2b3e1e292d1a3a07b5ad38633d84c668e6b16fc587f0d91ececad71c0b08986ef29f21fdb14075305540236225b474561d216ae56c065247bcf6c436c22b847e3b0a15ea2f3015c97f252373218fad40b7c3dc46a53ece93f1f699e58674e565d7a1e49fd9edb558382f7dc224d033b71df3011704432673d0144a9da2fa39f585df620016bf745636108147899e1e420e33d742a64727dc7790d205cd431b72d05be6f47b99f9ca985a61b88e1ea09691154e85c42372147e3dc08d0262a154e2e440eb2f337f57f1f0cc5a0dc4d56c16cb78057680b34cc286b1079475d024019313bbff3bdd9a1587fe80f724e656e10e5b20c2ae4364699f8405414ccdbf61fb1f712877d79938ee87f2d77fdd8431a182481cccbc2f89f3c2736aa956745389d03c28676fcbf1f62a723f9c56d751b7b9116dc3a6bf2c324fa58311a2310328ee0c2d12212f626aff96289048f2403e61e9808b3bf6e71be1d91115b473f056abdcebaa7e8518a75b49629e2960269921e7347bd3278410632a8b0946f45799515d1afef676ed8d274fdc2960ffd6ea606274c4602f9b8056180d347a454893605db1a509dec4a98007c19499f5ff8565aaaa19aff03a44ab20674d91113434e3f7eb50d50385ce3ffe1a3e635e74fd1dc36d27a39517e36a83303bcf8add2ff896f27e32479fe94a25f1e16c1ab2ca2d0666f9ece9423699fa4444c3b7a2d861ac9b357b1ceb3a16977d8c89ccebb6a75ce5e39fbfb38895c007000001f322b175583f68f44b97079b9a5eb82d8d79797b911dbda323c6be8456c5a4f23ca38c3e4f488a12980b93dbe4a12f8e54d426103170796d53ec257816e5ff4a25277763fdf0fc14091fd444ba4142f541b255425add66a61aa3b4445e09a9f3f1c8b85fe65c300f5f7b706effbd70cf2295f6d18f28ea982588c3d42863946d3a11772864770b1dce2725ab9316dd776a0a89c2f95027aa0208a6f4624421a6fd211d6cf8848ff191cd161418d1427a818b0f538c8a467732fcb47a67ec621577c31f8360c939776271d4ece94fb600d283c5696d0cc7b969fdf8cdd8685486f67fa52b989223e3a2be4d4c73932e74dcb52b2e581d20a1d6b2d2b600c2905b494a51ad6e29aacbd8d9ce7bca324951c5aefafeafa88627e2aff917d2b37d6f960000004fe3d6d3399ec4adb31f8cc93dec11897f1fe0e2724767edc3d503e1a2856205cef3abb8f12e0b1838834c0c5ae745d0f6f0180c4f1500b110944a2f52eb691c6439cf70626448f300792b7faa651efc455af64027c51a4e70e479001a1be194b8e857634b2c092f094cfa011cddd527eaaaecb2a5b11dd77c9e2937020ab5c7870930f5f9092b207aa2d5955d906ee33689f60957211dd81df3d5fd4b2992657ddf262fe8a44ab0e627fa2d0acb4198662e44ca82296f550120ebd31ed34fc574cddf38aa381ef1a75455e55d74e790cda0b9b2d6b868ff1431bddc11128ef26a1269c68a38b042853ec3406b5479b5c181d28941111a895ff0fc5d53f59fb00d39beb449b516b6b91cfe8c3a0828060000000000c65d40e9f836128ca1d7d717961ba86286807deeda12894a0dc92b74d9f7e16c08592b"},{"prim":"Some","args":[{"string":"expruN32WETsB2Dx1AynDmMufVr1As9qdnjRxKQ82rk2qZ4uxuKVMK"}]}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treeA UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeA); err != nil {
				t.Errorf("UnmarshalFromString treeA error = %v", err)
				return
			}
			a, err := treeA.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			var m map[string]interface{}
			if err := json.UnmarshalFromString(tt.data, &m); err != nil {
				t.Errorf("UnmarshalFromString(want) error = %v", err)
				return
			}
			if err := a.FromJSONSchema(m); (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.FromJSONSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
			b, err := a.ToParameters("")
			if err != nil {
				t.Errorf("ToParameters(a) error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(b))
		})
	}
}

func TestTypedAst_FindByName(t *testing.T) {
	tests := []struct {
		name         string
		tree         string
		fieldName    string
		isEntrypoint bool
		want         string
	}{
		{
			name:      "atomex/redeem",
			tree:      `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			fieldName: "redeem",
			want:      `{"prim":"bytes","annots":[":secret","%redeem"]}`,
		}, {
			name:      "atomex/initiate",
			tree:      `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			fieldName: "initiate",
			want:      `{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]}`,
		}, {
			name:      "atomex/add",
			tree:      `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			fieldName: "add",
			want:      `{"prim":"bytes","annots":[":hashed_secret","%add"]}`,
		}, {
			name:      "atomex/refund",
			tree:      `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			fieldName: "refund",
			want:      `{"prim":"bytes","annots":[":hashed_secret","%refund"]}`,
		}, {
			name:      "atomex/storage",
			tree:      `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			fieldName: "participant",
			want:      `{"prim":"address","annots":["%participant"]}`,
		}, {
			name:      "atomex/storage",
			tree:      `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			fieldName: "unknown",
			want:      "null",
		}, {
			name:      "atomex/storage",
			tree:      `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			fieldName: "unknown",
			want:      "null",
		}, {
			name:      "atomex/storage with modified map",
			tree:      `{"prim":"pair","args":[{"prim":"map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%initiator"]},{"prim":"address","annots":["%participant"]}],"annots":["%recipients"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%amount"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}]}]},{"prim":"unit"}]}`,
			fieldName: "unknown",
			want:      "null",
		}, {
			name:         "mainnet/KT1TnwBxgK4ayHuxrti6KKkJpWBHXBYRCX6H",
			tree:         `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%deposit"]},{"prim":"mutez","annots":["%reRoll"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%investor"]},{"prim":"address","annots":["%referrar"]}],"annots":["%register"]},{"prim":"or","args":[{"prim":"key_hash","annots":["%updateBaker"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%devFee"]},{"prim":"nat","annots":["%distributionPoolFee"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%referralFee"]},{"prim":"nat","annots":["%rewardPoolFee"]}]}],"annots":["%deposit"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%devFee"]},{"prim":"nat","annots":["%distributionPoolFee"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%referralFee"]},{"prim":"nat","annots":["%rewardPoolFee"]}]}],"annots":["%withdraw"]}],"annots":["%updateConfig"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%updateInvestorBaker"]},{"prim":"address","annots":["%updateOwner"]}]},{"prim":"or","args":[{"prim":"int","annots":["%updateRewardPoolReleaseDayCount"]},{"prim":"or","args":[{"prim":"mutez","annots":["%withdraw"]},{"prim":"unit","annots":["%withdrawDevFee"]}]}]}]}]}`,
			fieldName:    "withdraw",
			isEntrypoint: true,
			want:         `{"prim":"mutez","annots":["%withdraw"]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var treeA UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &treeA); err != nil {
				t.Errorf("UnmarshalFromString treeA error = %v", err)
				return
			}
			a, err := treeA.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			got := a.FindByName(tt.fieldName, tt.isEntrypoint)
			s, err := json.MarshalToString(got)
			if err != nil {
				t.Errorf("ToParameters(a) error = %v", err)
				return
			}
			s = strings.ReplaceAll(s, " ", "")
			assert.Equal(t, tt.want, s)
		})
	}
}

func TestTypedAst_FromParameters(t *testing.T) {
	tests := []struct {
		name       string
		tree       string
		entrypoint string
		data       string
		want       string
		wantErr    bool
	}{
		{
			name:       "atomex/initiate",
			tree:       `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			entrypoint: "default",
			data:       `{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"tz1aKTCbAUuea2RV9kxqRVRg3HT7f1RKnp6a"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"bytes":"182684bb81ef4008ed0acb6b7c40f3f3ddc08bda2ca11f4d2c31c43a8cff387e"},{"int":"1613402127"}]},{"int":"0"}]}]}]}]}`,
		}, {
			name:       "atomex/redeem",
			tree:       `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%participant"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%hashed_secret"]},{"prim":"timestamp","annots":["%refund_time"]}]},{"prim":"mutez","annots":["%payoff"]}],"annots":["%settings"]}],"annots":[":initiate","%initiate"]},{"prim":"bytes","annots":[":hashed_secret","%add"]}],"annots":["%fund"]},{"prim":"or","args":[{"prim":"bytes","annots":[":secret","%redeem"]},{"prim":"bytes","annots":[":hashed_secret","%refund"]}],"annots":["%withdraw"]}]}`,
			entrypoint: "default",
			data:       `{"prim":"Right","args":[{"prim":"Left","args":[{"bytes":"a5af9b02d19cc7bee9d52c9264feec28e0704e6434259926eb8693e8358d4d21"}]}]}`,
		}, {
			name:       "dexter/xtzToToken",
			tree:       `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"pair","args":[{"prim":"nat","annots":[":allowance"]},{"prim":"nat","annots":[":currentAllowance"]}]}],"annots":["%approve"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"nat","annots":[":minLqtMinted"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":maxTokensDeposited"]},{"prim":"timestamp","annots":[":deadline"]}]}],"annots":["%addLiquidity"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":lqtBurned"]}]}]},{"prim":"pair","args":[{"prim":"mutez","annots":[":minXtzWithdrawn"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensWithdrawn"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%removeLiquidity"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensBought"]},{"prim":"timestamp","annots":[":deadline"]}]}],"annots":["%xtzToToken"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":to"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokensSold"]},{"prim":"pair","args":[{"prim":"mutez","annots":[":minXtzBought"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%tokenToXtz"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":outputDexterContract"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensBought"]},{"prim":"address","annots":[":owner"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokensSold"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%tokenToToken"]},{"prim":"or","args":[{"prim":"key_hash","annots":["%updateTokenPool"]},{"prim":"nat","annots":["%updateTokenPoolInternal"]}]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}]},{"prim":"bool"}],"annots":["%setBaker"]},{"prim":"or","args":[{"prim":"address","annots":["%setManager"]},{"prim":"unit","annots":["%default"]}]}]}]}]}`,
			entrypoint: "xtzToToken",
			data:       `{"prim":"Pair","args":[{"string":"tz1KvTmY1k7nr2D62LNugeq4uWYTt513e5Mx"},{"prim":"Pair","args":[{"int":"2120690073714895"},{"int":"1613402584"}]}]}`,
		}, {
			name:       "dexter/tokenToXtz",
			tree:       `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"pair","args":[{"prim":"nat","annots":[":allowance"]},{"prim":"nat","annots":[":currentAllowance"]}]}],"annots":["%approve"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"nat","annots":[":minLqtMinted"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":maxTokensDeposited"]},{"prim":"timestamp","annots":[":deadline"]}]}],"annots":["%addLiquidity"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":lqtBurned"]}]}]},{"prim":"pair","args":[{"prim":"mutez","annots":[":minXtzWithdrawn"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensWithdrawn"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%removeLiquidity"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensBought"]},{"prim":"timestamp","annots":[":deadline"]}]}],"annots":["%xtzToToken"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":to"]}]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokensSold"]},{"prim":"pair","args":[{"prim":"mutez","annots":[":minXtzBought"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%tokenToXtz"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":outputDexterContract"]},{"prim":"pair","args":[{"prim":"nat","annots":[":minTokensBought"]},{"prim":"address","annots":[":owner"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"pair","args":[{"prim":"nat","annots":[":tokensSold"]},{"prim":"timestamp","annots":[":deadline"]}]}]}],"annots":["%tokenToToken"]},{"prim":"or","args":[{"prim":"key_hash","annots":["%updateTokenPool"]},{"prim":"nat","annots":["%updateTokenPoolInternal"]}]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"key_hash"}]},{"prim":"bool"}],"annots":["%setBaker"]},{"prim":"or","args":[{"prim":"address","annots":["%setManager"]},{"prim":"unit","annots":["%default"]}]}]}]}]}`,
			entrypoint: "tokenToXtz",
			data:       `{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1Ub7v6eoec6KdB5VTCqtAEVrQsjj1ZTFTk"},{"string":"tz1Ub7v6eoec6KdB5VTCqtAEVrQsjj1ZTFTk"}]},{"prim":"Pair","args":[{"int":"100000000000000000"},{"prim":"Pair","args":[{"int":"33682887"},{"int":"1613393537"}]}]}]}`,
		}, {
			name:       "aspen/transfer",
			tree:       `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%spender"]},{"prim":"nat","annots":["%value"]}],"annots":["%approve"]},{"prim":"unit","annots":["%finishIssuance"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%spender"]}]},{"prim":"contract","args":[{"prim":"nat"}],"annots":["%callback"]}],"annots":["%getAllowance"]},{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"contract","args":[{"prim":"nat"}],"annots":["%callback"]}],"annots":["%getBalance"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalSupply"]},{"prim":"nat","annots":["%issueTokens"]}]},{"prim":"or","args":[{"prim":"address","annots":["%setAdmin"]},{"prim":"address","annots":["%setIssuer"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"bool","annots":["%setPaused"]},{"prim":"address","annots":["%setRegistry"]}]},{"prim":"or","args":[{"prim":"address","annots":["%setRules"]},{"prim":"pair","args":[{"prim":"address","annots":["%from"]},{"prim":"pair","args":[{"prim":"address","annots":["%to"]},{"prim":"nat","annots":["%value"]}]}],"annots":["%transfer"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%from"]},{"prim":"pair","args":[{"prim":"address","annots":["%to"]},{"prim":"nat","annots":["%value"]}]}],"annots":["%transferOverride"]}]}]}`,
			entrypoint: "default",
			data:       `{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Pair","args":[{"string":"tz1escro5Ni7Y5re6WpNzU4BtNNtB75G6g1C"},{"prim":"Pair","args":[{"string":"tz1aPeXr4238SL8JMzm9tDiHrqXnScHjKvtJ"},{"int":"1100000"}]}]}]}]}]}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewTypedAstFromString(tt.tree)
			if err != nil {
				t.Errorf("NewTypedAstFromString error = %v", err)
				return
			}
			p := &types.Parameters{
				Entrypoint: tt.entrypoint,
				Value:      []byte(tt.data),
			}
			got, err := a.FromParameters(p)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.FromParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			res, err := got.ToParameters(DocsFull)
			if err != nil {
				t.Errorf("ToParameters() error = %v", err)
				return
			}
			if tt.want == "" {
				tt.want = tt.data
			}
			assert.Equal(t, tt.want, string(res))
		})
	}
}

func TestTypedAst_GetEntrypointsDocs(t *testing.T) {
	tests := []struct {
		name    string
		tree    string
		result  string
		wantErr bool
	}{
		{
			name:   "mainnet/KT1VsSxSXUkgw6zkBGgUuDXXuJs9ToPqkrCg/VestedFunds4",
			tree:   `{"prim":"or","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"contract","args":[{"prim":"unit"}],"annots":["%dest"]},{"prim":"mutez","annots":["%transfer_amount"]}],"annots":["%Transfer"]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"contract","args":[{"prim":"unit"}],"annots":["%pour_dest"]},{"prim":"key","annots":["%pour_authorizer"]}]}],"annots":["%Set_pour"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"key"}],"annots":["%signatories"]},{"prim":"nat","annots":["%group_threshold"]}]}],"annots":["%key_groups"]},{"prim":"nat","annots":["%overall_threshold"]}],"annots":["%Set_keys"]},{"prim":"option","args":[{"prim":"key_hash","annots":["%new_delegate"]}],"annots":["%Set_delegate"]}]}],"annots":["%action_input"]},{"prim":"list","args":[{"prim":"list","args":[{"prim":"option","args":[{"prim":"signature"}]}]}],"annots":["%signatures"]}],"annots":["%Action"]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"signature","annots":["%pour_auth"]},{"prim":"mutez","annots":["%pour_amount"]}],"annots":["%Pour"]}]}]}`,
			result: `[{"name":"Action","typedef":[{"name":"Action","type":"pair","args":[{"key":"action_input","value":"$action_input"},{"key":"signatures","value":"list(list(option(signature)))"}]},{"name":"action_input","type":"or","args":[{"key":"Transfer","value":"$Transfer"},{"key":"Set_pour","value":"option($Set_pour)"},{"key":"Set_keys","value":"$Set_keys"},{"key":"Set_delegate","value":"option(key_hash)"}]},{"name":"Transfer","type":"pair","args":[{"key":"dest","value":"contract(unit)"},{"key":"transfer_amount","value":"mutez"}]},{"name":"Set_pour","type":"pair","args":[{"key":"pour_dest","value":"contract(unit)"},{"key":"pour_authorizer","value":"key"}]},{"name":"Set_keys","type":"pair","args":[{"key":"key_groups","value":"list($key_groups_item)"},{"key":"overall_threshold","value":"nat"}]},{"name":"key_groups_item","type":"pair","args":[{"key":"signatories","value":"list(key)"},{"key":"group_threshold","value":"nat"}]},{"name":"signatures_item","type":"list(option(signature))"}]},{"name":"entrypoint_1","typedef":[{"name":"Pour","type":"pair","args":[{"key":"pour_auth","value":"signature"},{"key":"pour_amount","value":"mutez"}]}]}]`,
		}, {
			name:   "mainnet/KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn/tzBTC",
			tree:   `{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getVersion"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":spender"]}]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getAllowance"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getBalance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalSupply"]},{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalMinted"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat"}]}],"annots":["%getTotalBurned"]},{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address"}]}],"annots":["%getOwner"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address"}]}],"annots":["%getRedeemAddress"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"list","args":[{"prim":"nat"}]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"string"},{"prim":"pair","args":[{"prim":"string"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"string"},{"prim":"string"}]}]}]}]}]}]}]}],"annots":["%getTokenMetadata"]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}],"annots":["%run"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":currentVersion"]},{"prim":"nat","annots":[":newVersion"]}]},{"prim":"pair","args":[{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationScript"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newCode"]},{"prim":"option","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}]}],"annots":[":newPermCode"]}]}]}],"annots":["%upgrade"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat","annots":[":current"]},{"prim":"nat","annots":[":new"]}],"annots":["%epwBeginUpgrade"]},{"prim":"lambda","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}],"annots":[":migrationscript","%epwApplyMigration"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]}]}],"annots":[":contractcode","%epwSetCode"]},{"prim":"unit","annots":["%epwFinishUpgrade"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]},{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":["%approve"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}],"annots":["%mint"]},{"prim":"nat","annots":[":value","%burn"]}]},{"prim":"or","args":[{"prim":"address","annots":[":operator","%addOperator"]},{"prim":"address","annots":[":operator","%removeOperator"]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":[":redeem","%setRedeemAddress"]},{"prim":"unit","annots":["%pause"]}]},{"prim":"or","args":[{"prim":"unit","annots":["%unpause"]},{"prim":"or","args":[{"prim":"address","annots":[":newOwner","%transferOwnership"]},{"prim":"unit","annots":["%acceptOwnership"]}]}]}]}]}],"annots":["%safeEntrypoints"]}]}]}]}]}`,
			result: `[{"name":"getVersion","typedef":[{"name":"getVersion","type":"pair","args":[{"key":"@unit_5","value":"unit"},{"key":"@contract_6","value":"contract(nat)"}]}]},{"name":"getAllowance","typedef":[{"name":"getAllowance","type":"pair","args":[{"key":"owner","value":"address"},{"key":"spender","value":"address"},{"key":"@contract_12","value":"contract(nat)"}]}]},{"name":"getBalance","typedef":[{"name":"getBalance","type":"pair","args":[{"key":"owner","value":"address"},{"key":"@contract_17","value":"contract(nat)"}]}]},{"name":"getTotalSupply","typedef":[{"name":"getTotalSupply","type":"pair","args":[{"key":"@unit_21","value":"unit"},{"key":"@contract_22","value":"contract(nat)"}]}]},{"name":"getTotalMinted","typedef":[{"name":"getTotalMinted","type":"pair","args":[{"key":"@unit_25","value":"unit"},{"key":"@contract_26","value":"contract(nat)"}]}]},{"name":"getTotalBurned","typedef":[{"name":"getTotalBurned","type":"pair","args":[{"key":"@unit_31","value":"unit"},{"key":"@contract_32","value":"contract(nat)"}]}]},{"name":"getOwner","typedef":[{"name":"getOwner","type":"pair","args":[{"key":"@unit_35","value":"unit"},{"key":"@contract_36","value":"contract(address)"}]}]},{"name":"getRedeemAddress","typedef":[{"name":"getRedeemAddress","type":"pair","args":[{"key":"@unit_40","value":"unit"},{"key":"@contract_41","value":"contract(address)"}]}]},{"name":"getTokenMetadata","typedef":[{"name":"getTokenMetadata","type":"pair","args":[{"key":"@list_45","value":"list(nat)"},{"key":"@contract_47","value":"contract($contract_47_param)"}]},{"name":"@contract_47_param","type":"list (pair nat (pair string (pair string (pair nat (map string string)))))"}]},{"name":"run","typedef":[{"name":"run","type":"pair","args":[{"key":"@string_65","value":"string"},{"key":"@bytes_66","value":"bytes"}]}]},{"name":"upgrade","typedef":[{"name":"upgrade","type":"pair","args":[{"key":"currentVersion","value":"nat"},{"key":"newVersion","value":"nat"},{"key":"migrationScript","value":"$migrationScript"},{"key":"newCode","value":"option($newCode)"},{"key":"newPermCode","value":"option($newPermCode)"}]},{"name":"migrationScript","type":"lambda","args":[{"key":"input","value":"big_map bytes bytes"},{"key":"return","value":"big_map bytes bytes"}]},{"name":"newCode","type":"lambda","args":[{"key":"input","value":"pair (pair string bytes) (big_map bytes bytes)"},{"key":"return","value":"pair (list operation) (big_map bytes bytes)"}]},{"name":"newPermCode","type":"lambda","args":[{"key":"input","value":"pair unit (big_map bytes bytes)"},{"key":"return","value":"pair (list operation) (big_map bytes bytes)"}]}]},{"name":"epwBeginUpgrade","typedef":[{"name":"epwBeginUpgrade","type":"pair","args":[{"key":"current","value":"nat"},{"key":"new","value":"nat"}]}]},{"name":"epwApplyMigration","typedef":[{"name":"epwApplyMigration","type":"lambda"}]},{"name":"epwSetCode","typedef":[{"name":"epwSetCode","type":"lambda"}]},{"name":"epwFinishUpgrade","typedef":[{"name":"epwFinishUpgrade","type":"unit"}]},{"name":"transfer","typedef":[{"name":"transfer","type":"pair","args":[{"key":"from","value":"address"},{"key":"to","value":"address"},{"key":"value","value":"nat"}]}]},{"name":"approve","typedef":[{"name":"approve","type":"pair","args":[{"key":"spender","value":"address"},{"key":"value","value":"nat"}]}]},{"name":"mint","typedef":[{"name":"mint","type":"pair","args":[{"key":"to","value":"address"},{"key":"value","value":"nat"}]}]},{"name":"burn","typedef":[{"name":"burn","type":"nat"}]},{"name":"addOperator","typedef":[{"name":"addOperator","type":"address"}]},{"name":"removeOperator","typedef":[{"name":"removeOperator","type":"address"}]},{"name":"setRedeemAddress","typedef":[{"name":"setRedeemAddress","type":"address"}]},{"name":"pause","typedef":[{"name":"pause","type":"unit"}]},{"name":"unpause","typedef":[{"name":"unpause","type":"unit"}]},{"name":"transferOwnership","typedef":[{"name":"transferOwnership","type":"address"}]},{"name":"acceptOwnership","typedef":[{"name":"acceptOwnership","type":"unit"}]}]`,
		}, {
			name:   "mainnet/KT1U1JZaXoG4u1EPnhHL4R4otzkWc1L34q3c/Equisafe-KYC-registrar",
			tree:   `{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}],"annots":["%1"]}]}],"annots":["%addMembers"]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country_invest_limit"]},{"prim":"nat","annots":["%min_rating"]}]},{"prim":"pair","args":[{"prim":"map","args":[{"prim":"nat"},{"prim":"nat"}],"annots":["%rating_restrictions"]},{"prim":"timestamp","annots":["%vesting"]}]}]}],"annots":["%1"]}]},{"prim":"bool","annots":["%2"]}],"annots":["%checkMember"]}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%0"]},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%country"]},{"prim":"timestamp","annots":["%expires"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rating"]},{"prim":"nat","annots":["%region"]}]}]},{"prim":"bool","annots":["%restricted"]}]},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%1"]}],"annots":["%getMember"]},{"prim":"list","args":[{"prim":"address"}],"annots":["%removeMembers"]}]}]}`,
			result: `[{"name":"addMembers","typedef":[{"name":"addMembers","type":"list($addMembers_item)"},{"name":"addMembers_item","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"$1"}]},{"name":"1","type":"pair","args":[{"key":"country","value":"nat"},{"key":"expires","value":"timestamp"},{"key":"rating","value":"nat"},{"key":"region","value":"nat"},{"key":"restricted","value":"bool"}]}]},{"name":"checkMember","typedef":[{"name":"checkMember","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"map(nat, $1_value)"},{"key":"2","value":"bool"}]},{"name":"1_value","type":"pair","args":[{"key":"country_invest_limit","value":"nat"},{"key":"min_rating","value":"nat"},{"key":"rating_restrictions","value":"map(nat, nat)"},{"key":"vesting","value":"timestamp"}]}]},{"name":"getMember","typedef":[{"name":"getMember","type":"pair","args":[{"key":"0","value":"address"},{"key":"1","value":"$1"}]},{"name":"1","type":"lambda","args":[{"key":"input","value":"pair (pair (pair (nat %country) (timestamp %expires)) (pair (nat %rating) (nat %region))) (bool %restricted)"},{"key":"return","value":"list operation"}]}]},{"name":"removeMembers","typedef":[{"name":"removeMembers","type":"list(address)"}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UnmarshalFromString tree error = %v", err)
				return
			}
			a, err := tree.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST(a) error = %v", err)
				return
			}
			got, err := a.GetEntrypointsDocs()
			if (err != nil) != tt.wantErr {
				t.Errorf("TypedAst.GetEntrypointsDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotStr, err := json.MarshalToString(got)
			if err != nil {
				t.Errorf("MarshalToString(fot) error = %v", err)
				return
			}
			assert.Equal(t, tt.result, gotStr)
		})
	}
}
