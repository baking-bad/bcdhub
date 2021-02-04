package normalize

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

func TestType(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		want    string
		wantErr bool
	}{
		{
			name: "comb 1",
			typ:  `{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "string"}]}`,
			want: `{"prim": "pair", "args": [{"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}, {"prim": "int"}]}`,
		}, {
			name: "comb 1 with annots",
			typ:  `{"prim": "pair", "annots": ["%test"], "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "string"}]}`,
			want: `{"prim": "pair", "annots": ["%test"], "args": [{"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}, {"prim": "int"}]}`,
		}, {
			name: "pair",
			typ:  `{"prim":"pair","args":[{"prim": "string"},{"prim": "int"}]}`,
			want: `{"prim":"pair","args":[{"prim": "string"},{"prim": "int"}]}`,
		}, {
			name: "comb 2",
			typ:  `{"prim": "Left", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "string"}]}]}`,
			want: `{"prim": "Left", "args": [{"prim": "pair", "args": [{"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}, {"prim": "int"}]}]}`,
		}, {
			name:    "Invalid data",
			typ:     `10`,
			wantErr: true,
		}, {
			name: "prim storage",
			typ:  `{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":user"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":[":approvals"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}]}`,
			want: `{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":user"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":[":approvals"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.typ)
			got, err := Type(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Type() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			want := make(map[string]interface{})
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Errorf("Unmarshal(want) error = %v", err)
				return
			}
			gotMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(got.Raw), &gotMap); err != nil {
				t.Errorf("Unmarshal(got) error = %v", err)
				return
			}
			if !reflect.DeepEqual(gotMap, want) {
				t.Errorf("Type() = %v, want %v", gotMap, want)
			}
		})
	}
}

func TestData(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		typ     string
		want    string
		wantErr bool
	}{
		{
			name: "top level comb",
			data: `[{"int":"0"},{"bytes":"0000b5dc83da2da6bc59b5564eeac9760ff19a6280fc"},{"prim":"False"},{"int":"10000"}]`,
			typ:  `{"prim": "pair", "args":[{"prim": "pair", "args":[{"prim": "pair", "args":[{"prim": "bool"}, {"prim": "int"}]}, {"prim": "bytes"}]}, {"prim": "int"}]}`,
			want: `{"prim": "Pair", "args":[{"prim": "Pair", "args":[{"prim": "Pair", "args":[{"prim": "False"}, {"int":"10000"}]}, {"bytes":"0000b5dc83da2da6bc59b5564eeac9760ff19a6280fc"}]}, {"int": "0"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.data)
			typ := gjson.Parse(tt.typ)
			got, err := Data(data, typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			want := make(map[string]interface{})
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Errorf("Unmarshal(want) error = %v", err)
				return
			}
			gotMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(got.Raw), &gotMap); err != nil {
				t.Errorf("Unmarshal(got) error = %v", err)
				return
			}
			if !reflect.DeepEqual(gotMap, want) {
				t.Errorf("Type() = %v, want %v", gotMap, want)
			}
		})
	}
}

func Test_processOptionValue(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		typ     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "None",
			data: map[string]interface{}{
				"prim": "None",
			},
			typ: `{"prim": "option", "args":[{"prim": "int"}]}`,
			want: map[string]interface{}{
				"prim": "None",
			},
		}, {
			name: "Some without combs",
			data: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"int": 10,
					},
				},
			},
			typ: `{"prim": "option", "args":[{"prim": "int"}]}`,
			want: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"int": 10,
					},
				},
			},
		}, {
			name:    "nil data",
			typ:     `{"prim": "int"}`,
			wantErr: true,
		}, {
			name: "Some with combs",
			data: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"prim": "Pair",
						"args": []interface{}{
							map[string]interface{}{
								"int": 10,
							},
							map[string]interface{}{
								"int": 11,
							},
							map[string]interface{}{
								"int": 12,
							},
						},
					},
				},
			},
			typ: `{"prim": "option", "args":[{"prim": "pair", "args":[{"prim": "pair", "args":[{"prim": "int"},{"prim": "int"}]}, {"prim": "int"}]}]}`,
			want: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"prim": "Pair",
						"args": []interface{}{
							map[string]interface{}{
								"prim": "Pair",
								"args": []interface{}{
									map[string]interface{}{
										"int": 11,
									},
									map[string]interface{}{
										"int": 12,
									},
								},
							},
							map[string]interface{}{
								"int": 10,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := gjson.Parse(tt.typ)
			got, err := processOptionValue(tt.data, typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("processOptionValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processOptionValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processMapValue(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		typ     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "simple",
			typ:  `{"prim": "map", "args":[{"prim": "address","annots": [":spender"]},{"prim": "nat","annots": [":value"]}]}`,
			data: []interface{}{},
			want: []interface{}{},
		}, {
			name: "simple with data",
			typ:  `{"prim": "map", "args":[{"prim": "address","annots": [":spender"]},{"prim": "nat","annots": [":value"]}]}`,
			data: []interface{}{
				map[string]interface{}{
					"prim": "Elt",
					"args": []interface{}{
						map[string]interface{}{
							"string": "address",
						},
						map[string]interface{}{
							"int": 12,
						},
					},
				},
				map[string]interface{}{
					"prim": "Elt",
					"args": []interface{}{
						map[string]interface{}{
							"string": "address2",
						},
						map[string]interface{}{
							"int": 10,
						},
					},
				},
			},
			want: []interface{}{
				map[string]interface{}{
					"prim": "Elt",
					"args": []interface{}{
						map[string]interface{}{
							"string": "address",
						},
						map[string]interface{}{
							"int": 12,
						},
					},
				},
				map[string]interface{}{
					"prim": "Elt",
					"args": []interface{}{
						map[string]interface{}{
							"string": "address2",
						},
						map[string]interface{}{
							"int": 10,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := gjson.Parse(tt.typ)
			got, err := processMapValue(tt.data, typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMapValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processMapValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
