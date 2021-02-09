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
								Const: "left",
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
								Const: "right",
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

func TestTypedAst_Docs(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
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
			name: "mainnet/KT1ChNsEFxwyCbJyWGSL3KdjeXE28AY1Kaog/BakersRegistry",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"key_hash","annots":["%delegate"]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes","annots":["%bakerName"]},{"prim":"bool","annots":["%openForDelegation"]}]},{"prim":"bytes","annots":["%bakerOffchainRegistryUrl"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%split"]},{"prim":"list","args":[{"prim":"address"}],"annots":["%bakerPaysFromAccounts"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%minDelegation"]},{"prim":"bool","annots":["%subtractPayoutsLessThanMin"]}]},{"prim":"pair","args":[{"prim":"int","annots":["%payoutDelay"]},{"prim":"pair","args":[{"prim":"nat","annots":["%payoutFrequency"]},{"prim":"int","annots":["%minPayout"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bool","annots":["%bakerChargesTransactionFee"]},{"prim":"nat","annots":["%paymentConfigMask"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%overDelegationThreshold"]},{"prim":"bool","annots":["%subtractRewardsFromUninvitedDelegation"]}]}]}]}]}]}],"annots":["%data"]},{"prim":"option","args":[{"prim":"address"}],"annots":["%reporterAccount"]}]}],"annots":["%set_data"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"mutez","annots":["%signup_fee"]},{"prim":"mutez","annots":["%update_fee"]}],"annots":["%set_fees"]},{"prim":"contract","args":[{"prim":"unit"}],"annots":["%withdraw"]}]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"set_data","value":"$set_data"},{"key":"set_fees","value":"$set_fees"},{"key":"withdraw","value":"contract(unit)"}]},{"name":"set_data","type":"pair","args":[{"key":"delegate","value":"key_hash"},{"key":"data","value":"option($data)"},{"key":"reporterAccount","value":"option(address)"}]},{"name":"data","type":"pair","args":[{"key":"bakerName","value":"bytes"},{"key":"openForDelegation","value":"bool"},{"key":"bakerOffchainRegistryUrl","value":"bytes"},{"key":"split","value":"nat"},{"key":"bakerPaysFromAccounts","value":"list(address)"},{"key":"minDelegation","value":"nat"},{"key":"subtractPayoutsLessThanMin","value":"bool"},{"key":"payoutDelay","value":"int"},{"key":"payoutFrequency","value":"nat"},{"key":"minPayout","value":"int"},{"key":"bakerChargesTransactionFee","value":"bool"},{"key":"paymentConfigMask","value":"nat"},{"key":"overDelegationThreshold","value":"nat"},{"key":"subtractRewardsFromUninvitedDelegation","value":"bool"}]},{"name":"set_fees","type":"pair","args":[{"key":"signup_fee","value":"mutez"},{"key":"update_fee","value":"mutez"}]}]`,
		}, {
			name: "edonet/KT1D7MfG9CEBav7TXsa4xbPL3QZgR5eEgx7g/ticket",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"ticket","args":[{"prim":"address"}]},{"prim":"nat"}],"annots":["%sendConvertedBalance"]},{"prim":"mutez","annots":["%setConversionRate"]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"sendConvertedBalance","value":"$sendConvertedBalance"},{"key":"setConversionRate","value":"mutez"}]},{"name":"sendConvertedBalance","type":"pair","args":[{"key":"@ticket_3","value":"ticket(address)"},{"key":"@nat_5","value":"nat"}]}]`,
		}, {
			name: "edonet/KT1MaW1LQ77YpZwtmrb4aHBUteqPN91AruWB/sapling_state",
			data: `[{"prim":"parameter","args":[{"prim":"sapling_state","args":[{"int":"8"}]}]}]`,
			want: `[{"name":"@sapling_state_1","type":"sapling_state(8)"}]`,
		}, {
			name: "edonet/KT1PbFKg3mgJAadojxPjh3EQSLNsuAkYjNnQ/sapling_transaction",
			data: `[{"prim":"parameter","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"sapling_transaction","args":[{"int":"8"}]},{"prim":"option","args":[{"prim":"key_hash"}]}]}]}]}]`,
			want: `[{"name":"@list_1","type":"list($list_1_item)"},{"name":"@list_1_item","type":"pair","args":[{"key":"@sapling_transaction_3","value":"sapling_transaction(8)"},{"key":"@option_5","value":"option(key_hash)"}]}]`,
		}, {
			name: "unknown contract 1",
			data: `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"nat"}],"annots":["%setList"]},{"prim":"map","args":[{"prim":"nat"},{"prim":"mutez"}],"annots":["%setMap"]}]},{"prim":"set","args":[{"prim":"nat"}],"annots":["%setSet"]}]}]}]`,
			want: `[{"name":"@or_1","type":"or","args":[{"key":"setList","value":"list(nat)"},{"key":"setMap","value":"map(nat, mutez)"},{"key":"setSet","value":"set(nat)"}]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			untyped, err := NewScript([]byte(tt.data))
			if err != nil {
				t.Errorf("NewScript() error = %v", err)
				return
			}
			a, err := untyped.Parameter.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			got, err := a.Docs(DocsFull)
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
		want    bool
		wantErr bool
	}{
		{
			name: "simple true",
			typ:  `[{"prim": "string"}]`,
			a:    `{"string": "test"}`,
			b:    `{"string": "test"}`,
			want: true,
		}, {
			name: "simple false",
			typ:  `[{"prim": "string"}]`,
			a:    `{"string": "test"}`,
			b:    `{"string": "another"}`,
			want: false,
		}, {
			name: "pair with option None true",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			want: true,
		}, {
			name: "pair with option Some true",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"bytes": "0000cd1a410ffd5315ded34337f5f76edff48a13999a"}]}]}`,
			want: true,
		}, {
			name: "pair with option false",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "None"}]}`,
			want: false,
		}, {
			name: "pair with option Some false",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "address"}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "tz1eLWfccL46VAUjtyz9kEKgzuKnwyZH4rTA"}]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[{"string": "KT1BUKeJTemAaVBfRz6cqxeUBQGQqMxfG19A"}]}]}`,
			want: false,
		}, {
			name: "pair with option uncomparable false",
			typ:  `[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "option", "args":[{"prim": "set", "args":[{"prim": "unit"}]}]}]}]`,
			a:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[[]]}]}`,
			b:    `{"prim": "Pair", "args":[{"int": "100"}, {"prim": "Some", "args":[[]]}]}`,
			want: false,
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
