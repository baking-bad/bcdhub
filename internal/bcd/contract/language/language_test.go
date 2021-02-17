package language

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/tidwall/gjson"
)

func TestDetectSmartPy(t *testing.T) {
	testCases := []struct {
		name string
		n    node.Node
		res  string
	}{
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}("SmartPy is awesome"),
				Type:  consts.KeyString,
			},
			res: LangSmartPy,
		},
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}("start self. end"),
				Type:  consts.KeyString,
			},
			res: LangSmartPy,
		},
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}("start sp. end"),
				Type:  consts.KeyString,
			},
			res: LangSmartPy,
		},
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}("WrongCondition"),
				Type:  consts.KeyString,
			},
			res: LangSmartPy,
		},
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}(`Get-item:123`),
				Type:  consts.KeyString,
			},
			res: LangSmartPy,
		},
		{
			name: "SmartPy Value",
			n: node.Node{
				Value: interface{}(`Get-item:123a`),
				Type:  consts.KeyString,
			},
			res: LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := GetFromCode(tt.n); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestDetectLiquidity(t *testing.T) {
	testCases := []struct {
		name string
		n    node.Node
		res  string
	}{
		{
			name: "Liquidity Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%0 _slash_"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"_slash_"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{":_entries"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{`@\w+_slash_1`},
			},
			res: LangLiquidity,
		},
		{
			name: "Not Liquidity Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"123"},
			},
			res: LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := GetFromCode(tt.n); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestDetectLIGO(t *testing.T) {
	testCases := []struct {
		name string
		n    node.Node
		res  string
	}{
		{
			name: "Ligo Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%0"},
			},
			res: LangLigo,
		},
		{
			name: "Ligo Annotation",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%1"},
			},
			res: LangLigo,
		},
		{
			name: "Ligo Annotation",
			n: node.Node{
				Prim:        "nat",
				Annotations: []string{"%3"},
			},
			res: LangLigo,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%3s"},
			},
			res: LangUnknown,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%%3"},
			},
			res: LangUnknown,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%abc"},
			},
			res: LangUnknown,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"%-42"},
			},
			res: LangUnknown,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"abc"},
			},
			res: LangUnknown,
		},
		{
			name: "Not Ligo",
			n: node.Node{
				Prim:        "address",
				Annotations: []string{"0"},
			},
			res: LangUnknown,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("GET_FORCE"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("get_force"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("MAP FIND"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("start get_entrypoint end"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("get_contract end"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
		{
			name: "Ligo Value",
			n: node.Node{
				Value: interface{}("failed assertion"),
				Type:  consts.KeyString,
			},
			res: LangLigo,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := GetFromCode(tt.n); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestDetectLorentz(t *testing.T) {
	testCases := []struct {
		name string
		n    node.Node
		res  string
	}{
		{
			name: "Lorentz Value",
			n: node.Node{
				Value: interface{}("UStore"),
				Type:  consts.KeyString,
			},
			res: LangLorentz,
		},
		{
			name: "Lorentz Value",
			n: node.Node{
				Value: interface{}("something UStore strange"),
				Type:  consts.KeyString,
			},
			res: LangLorentz,
		},
		{
			name: "Lorentz Value",
			n: node.Node{
				Value: interface{}("123 UStore"),
				Type:  consts.KeyString,
			},
			res: LangLorentz,
		},
		{
			name: "Not Lorentz Value",
			n: node.Node{
				Value: interface{}("start end"),
				Type:  consts.KeyString,
			},
			res: LangUnknown,
		},
		{
			name: "Not Lorentz Value",
			n: node.Node{
				Value: interface{}("ustore"),
				Type:  consts.KeyString,
			},
			res: LangUnknown,
		},
		{
			name: "Not Lorentz Value",
			n: node.Node{
				Value: interface{}("Ustore"),
				Type:  consts.KeyString,
			},
			res: LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := GetFromCode(tt.n); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestGetFromParameter(t *testing.T) {
	testCases := []struct {
		name string
		n    node.Node
		res  string
	}{
		{
			name: "liquidity entrypoints",
			n: node.Node{
				Annotations: []string{"%_Liq_entry"},
			},
			res: LangLiquidity,
		},
		{
			name: "lorentz entrypoints",
			n: node.Node{
				Annotations: []string{"%epwBeginUpgrade"},
			},
			res: LangLorentz,
		},
		{
			name: "lorentz entrypoints",
			n: node.Node{
				Annotations: []string{"%epwApplyMigration"},
			},
			res: LangLorentz,
		},
		{
			name: "lorentz entrypoints",
			n: node.Node{
				Annotations: []string{"%epwSetCode"},
			},
			res: LangLorentz,
		},
		{
			name: "random entrypoints",
			n: node.Node{
				Annotations: []string{"%setCode"},
			},
			res: LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if result := GetFromParameter(tt.n); result != tt.res {
				t.Errorf("Invalid result.\nGot:%v\nExpected:%v", result, tt.res)
			}
		})
	}
}

func TestGetFromFirstPrim(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lorentz",
			input: `{"prim":"CAST"}`,
			want:  LangLorentz,
		},
		{
			name:  "michelson",
			input: `{"prim":"pair"}`,
			want:  LangUnknown,
		},
		{
			name:  "michelson",
			input: `[{"prim": "CAST"},{"prim": "bool"}]`,
			want:  LangUnknown,
		},
		{
			name:  "michelson",
			input: `[[{"prim": "nat"},{"prim": "CAST"}]]`,
			want:  LangUnknown,
		},
		{
			name:  "scaml",
			input: `[]`,
			want:  LangSCaml,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			parsed := gjson.Parse(tt.input)
			if got := GetFromFirstPrim(parsed); got != tt.want {
				t.Errorf("GetFromFirstPrim invalid. expected: %v, got: %v", tt.want, got)
			}
		})
	}
}
