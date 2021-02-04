package normalize

import (
	"reflect"
	"testing"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/stretchr/testify/assert"
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
			want: `{"prim": "pair", "args": [ {"prim": "int"}, {"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}]}`,
		}, {
			name: "comb 1 with annots",
			typ:  `{"prim": "pair", "annots": ["%test"], "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "string"}]}`,
			want: `{"prim": "pair", "annots": ["%test"], "args": [{"prim": "int"}, {"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}]}`,
		}, {
			name: "pair",
			typ:  `{"prim":"pair","args":[{"prim": "string"},{"prim": "int"}]}`,
			want: `{"prim":"pair","args":[{"prim": "string"},{"prim": "int"}]}`,
		}, {
			name: "comb 2",
			typ:  `{"prim": "Left", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "string"}]}]}`,
			want: `{"prim": "Left", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args":[{"prim": "int"}, {"prim": "string"}]}]}]}`,
		}, {
			name:    "Invalid data",
			typ:     `10`,
			wantErr: true,
		}, {
			name: "prim storage",
			typ:  `{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":user"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":[":approvals"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}]}`,
			want: `{"prim":"storage","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address","annots":[":user"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":[":approvals"]}]}],"annots":["%ledger"]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}]}`,
		}, {
			name: "map",
			typ:  `{"prim": "map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			want: `{"prim": "map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}]}]}`,
		}, {
			name: "list",
			typ:  `{"prim": "list", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			want: `{"prim": "list", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}]}`,
		}, {
			name: "option",
			typ:  `{"prim": "option", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			want: `{"prim": "option", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}]}]}`,
		}, {
			name: "big_map",
			typ:  `{"prim": "big_map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			want: `{"prim": "big_map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}]}]}]}]}`,
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
			typ:  `{"prim": "pair", "args":[{"prim": "int"}, {"prim": "pair", "args":[{"prim": "bytes"}, {"prim": "pair", "args":[{"prim": "bool"}, {"prim": "int"}]}]}]}`,
			want: `{"prim": "Pair", "args":[{"int": "0"}, {"prim": "Pair", "args":[{"bytes":"0000b5dc83da2da6bc59b5564eeac9760ff19a6280fc"}, {"prim": "Pair", "args":[{"prim": "False"}, {"int":"10000"}]}]}]}`,
		}, {
			name: "map",
			typ:  `{"prim": "map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			data: `[{"prim": "Elt", "args": [[{"int": "0"}, {"int": "1"}, {"int": "2"}], [{"int": "0"}, {"int": "1"}, {"int": "2"}, {"int": "3"}]]},{"prim": "Elt", "args": [{"prim": "Pair", "args": [{"int": "4"}, {"int": "5"}, {"int": "6"}]}, [{"int": "0"}, {"int": "1"}, {"int": "2"}, {"int": "3"}]]}]`,
			want: `[{"args": [{"args": [{"int": "0"},
										{"args": [{"int": "1"}, {"int": "2"}], "prim": "Pair"}],
									"prim": "Pair"},
									{"args": [{"int": "0"},
										{"args": [{"int": "1"},
										{"args": [{"int": "2"}, {"int": "3"}], "prim": "Pair"}],
										"prim": "Pair"}],
									"prim": "Pair"}],
									"prim": "Elt"},
									{"args": [{"args": [{"int": "4"},
										{"args": [{"int": "5"}, {"int": "6"}], "prim": "Pair"}],
									"prim": "Pair"},
									{"args": [{"int": "0"},
										{"args": [{"int": "1"},
										{"args": [{"int": "2"}, {"int": "3"}], "prim": "Pair"}],
										"prim": "Pair"}],
									"prim": "Pair"}],
									"prim": "Elt"}]`,
		}, {
			name: "list",
			typ:  `{"prim": "list", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			data: `[[{"int": "0"}, {"int": "1"}, {"int": "2"}],{"prim": "Pair", "args": [{"int": "4"}, {"int": "5"}, {"int": "6"}]}]`,
			want: `[{"args": [{"int": "0"},{"args": [{"int": "1"}, {"int": "2"}], "prim": "Pair"}],"prim": "Pair"},{"args": [{"int": "4"},{"args": [{"int": "5"}, {"int": "6"}], "prim": "Pair"}],"prim": "Pair"}]`,
		}, {
			name: "option",
			typ:  `{"prim": "option", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			data: `{"prim": "Some", "args": [[{"int": "0"}, {"int": "1"}, {"int": "2"}, {"int": "3"}]]}`,
			want: `{"args": [{"args": [{"int": "0"},{"args": [{"int": "1"},{"args": [{"int": "2"}, {"int": "3"}], "prim": "Pair"}],"prim": "Pair"}],"prim": "Pair"}],"prim": "Some"}`,
		}, {
			name: "bigmap ptr",
			typ:  `{"prim": "big_map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			data: `{"int": "10"}`,
			want: `{"int": "10"}`,
		}, {
			name: "bigmap array",
			typ:  `{"prim": "big_map", "args": [{"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}, {"prim": "pair", "args": [{"prim": "int"}, {"prim": "int"}, {"prim": "int"}, {"prim": "int"}]}]}`,
			data: `[{"prim": "Elt", "args": [[{"int": "0"}, {"int": "1"}, {"int": "2"}], [{"int": "0"}, {"int": "1"}, {"int": "2"}, {"int": "3"}]]},{"prim": "Elt", "args": [{"prim": "Pair", "args": [{"int": "4"}, {"int": "5"}, {"int": "6"}]}, [{"int": "0"}, {"int": "1"}, {"int": "2"}, {"int": "3"}]]}]`,
			want: `[{"args": [{"args": [{"int": "0"},
										{"args": [{"int": "1"}, {"int": "2"}], "prim": "Pair"}],
									"prim": "Pair"},
									{"args": [{"int": "0"},
										{"args": [{"int": "1"},
										{"args": [{"int": "2"}, {"int": "3"}], "prim": "Pair"}],
										"prim": "Pair"}],
									"prim": "Pair"}],
									"prim": "Elt"},
									{"args": [{"args": [{"int": "4"},
										{"args": [{"int": "5"}, {"int": "6"}], "prim": "Pair"}],
									"prim": "Pair"},
									{"args": [{"int": "0"},
										{"args": [{"int": "1"},
										{"args": [{"int": "2"}, {"int": "3"}], "prim": "Pair"}],
										"prim": "Pair"}],
									"prim": "Pair"}],
									"prim": "Elt"}]`,
		}, {
			name: "KT1GjKvUhpJLDaAHifnohmLjEfvn4fCkhKbs",
			typ:  `{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"nat"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"map"}],"prim":"pair"}],"prim":"big_map"},{"args":[{"args":[{"prim":"address"},{"prim":"bool"}],"prim":"pair"},{"args":[{"prim":"nat"},{"args":[{"prim":"address"},{"prim":"address"}],"prim":"or"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}`,
			data: `[{"int":"1"},{"args":[{"bytes":"000235bad5bc6e2f470762a82340d11b2bbf6c3a84b2"},{"prim":"False"}],"prim":"Pair"},{"int":"17"},{"args":[{"bytes":"000235bad5bc6e2f470762a82340d11b2bbf6c3a84b2"}],"prim":"Left"}]`,
			want: `{"args": [{"int": "1"},{"args": [{"args": [{"bytes": "000235bad5bc6e2f470762a82340d11b2bbf6c3a84b2"},{"prim": "False"}],"prim": "Pair"},{"args": [{"int": "17"},{"args": [{"bytes": "000235bad5bc6e2f470762a82340d11b2bbf6c3a84b2"}],"prim": "Left"}],"prim": "Pair"}],"prim": "Pair"}],"prim": "Pair"}`,
		}, {
			name: "KT1GjKvUhpJLDaAHifnohmLjEfvn4fCkhKbs 2",
			typ:  `{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]}]}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"bool"}]},{"prim":"nat"},{"prim":"or","args":[{"prim":"address"},{"prim":"address"}]}]}`,
			data: `{"prim":"Pair","args":[{"int":"1"},{"prim":"Pair","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"},{"prim":"False"}]},{"int":"17"},{"prim":"Left","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"}]}]}`,
			want: `{"prim":"Pair","args":[{"int":"1"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"},{"prim":"False"}]},{"prim":"Pair","args":[{"int":"17"},{"prim":"Left","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"}]}]}]}]}`,
		}, {
			name: "KT1GjKvUhpJLDaAHifnohmLjEfvn4fCkhKbs default",
			typ:  `{"args":[{"args":[{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"args":[{"args":[{"args":[{"prim":"address"},{"prim":"address"}],"prim":"pair"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"}],"prim":"or"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"nat"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"bool"},{"prim":"address"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"args":[{"prim":"unit"},{"args":[{"prim":"address"}],"prim":"contract"}],"prim":"pair"},{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"}],"prim":"or"},{"args":[{"args":[{"prim":"address"},{"prim":"nat"}],"prim":"pair"},{"prim":"address"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"}`,
			data: `{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"},{"prim":"Pair","args":[{"string":"tz1ZfrERcALBwmAqwonRXYVQBDT9BjNjBHJu"},{"int":"1"}]}]}]}]}]}`,
			want: `{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"tz3RE9FM2HK2aSpoDHsQZaqM1PsqrAZR7JqX"},{"prim":"Pair","args":[{"string":"tz1ZfrERcALBwmAqwonRXYVQBDT9BjNjBHJu"},{"int":"1"}]}]}]}]}]}`,
		}, {
			name: "KT1N6VjvuuBfXBbsyMby96zkYeaWuqCto69Q receive",
			typ:  `{"annots":["%receive"],"args":[{"prim":"unit"}],"prim":"ticket"}`,
			data: `{"prim":"Pair","args":[{"bytes":"01aaa4f29006915e1c7b6867024c3fa73337caab3700"},{"prim":"Pair","args":[{"prim":"Unit"},{"int":"2"}]}]}`,
			want: `{"prim":"Pair","args":[{"bytes":"01aaa4f29006915e1c7b6867024c3fa73337caab3700"},{"prim":"Pair","args":[{"prim":"Unit"},{"int":"2"}]}]}`,
		}, {
			name: "KT1N6VjvuuBfXBbsyMby96zkYeaWuqCto69Q send",
			typ:  `{"annots":["%send"],"args":[{"annots":["%destination"],"args":[{"args":[{"prim":"unit"}],"prim":"ticket"}],"prim":"contract"},{"annots":["%amount"],"prim":"nat"},{"annots":["%ticketer"],"prim":"address"}],"prim":"pair"}`,
			data: `{"prim":"Pair","args":[{"string":"KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ%receive"},{"prim":"Pair","args":[{"int":"1"},{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"}]}]}`,
			want: `{"prim":"Pair","args":[{"string":"KT1AqgENraEg8oro9gJ61mocjRLGBBkya4DQ%receive"},{"prim":"Pair","args":[{"int":"1"},{"string":"KT1Q9438XGRGQmWFEuoi5heQiASA5eszRi2x"}]}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := gjson.Parse(tt.data)
			typ := gjson.Parse(tt.typ)

			normalizedTyp, err := Type(typ)
			if err != nil {
				t.Errorf("Type() error = %v", err)
				return
			}
			got, err := Data(data, normalizedTyp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			var want interface{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Errorf("Unmarshal(want) error = %v", err)
				return
			}
			var gotMap interface{}
			if err := json.Unmarshal([]byte(got.Raw), &gotMap); err != nil {
				logger.Debug(got.Raw)
				t.Errorf("Unmarshal(got) error = %v", err)
				return
			}
			assert.Equal(t, want, gotMap)
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
						"int": "10",
					},
				},
			},
			typ: `{"prim": "option", "args":[{"prim": "int"}]}`,
			want: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"int": "10",
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
								"int": "10",
							},
							map[string]interface{}{
								"int": "11",
							},
							map[string]interface{}{
								"int": "12",
							},
						},
					},
				},
			},
			typ: `{"prim": "option", "args":[{"prim": "pair", "args":[{"prim": "int"}, {"prim": "pair", "args":[{"prim": "int"},{"prim": "int"}]}]}]}`,
			want: map[string]interface{}{
				"prim": "Some",
				"args": []interface{}{
					map[string]interface{}{
						"prim": "Pair",
						"args": []interface{}{
							map[string]interface{}{
								"int": "10",
							},
							map[string]interface{}{
								"prim": "Pair",
								"args": []interface{}{
									map[string]interface{}{
										"int": "11",
									},
									map[string]interface{}{
										"int": "12",
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
