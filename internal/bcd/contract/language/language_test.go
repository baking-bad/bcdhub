package language

import (
	"testing"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func TestDetectSmartPy(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		res   string
	}{
		{
			name:  "SmartPy Value",
			value: "SmartPy is awesome",
			res:   LangSmartPy,
		},
		{
			name:  "SmartPy Value",
			value: "start self. end",
			res:   LangSmartPy,
		},
		{
			name:  "SmartPy Value",
			value: "start sp. end",
			res:   LangSmartPy,
		},
		{
			name:  "SmartPy Value",
			value: "WrongCondition",
			res:   LangSmartPy,
		},
		{
			name:  "SmartPy Value",
			value: `Get-item:123`,
			res:   LangSmartPy,
		},
		{
			name:  "SmartPy Value",
			value: `Get-item:123a`,
			res:   LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			node := &base.Node{StringValue: &tt.value}
			if result := GetFromCode(node); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestDetectLiquidity(t *testing.T) {
	testCases := []struct {
		name string
		n    *base.Node
		res  string
	}{
		{
			name: "Liquidity Annotation",
			n: &base.Node{
				Prim:   "address",
				Annots: []string{"%0 _slash_"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: &base.Node{
				Prim:   "address",
				Annots: []string{"_slash_"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: &base.Node{
				Prim:   "address",
				Annots: []string{":_entries"},
			},
			res: LangLiquidity,
		},
		{
			name: "Liquidity Annotation",
			n: &base.Node{
				Prim:   "address",
				Annots: []string{`@\w+_slash_1`},
			},
			res: LangLiquidity,
		},
		{
			name: "Not Liquidity Annotation",
			n: &base.Node{
				Prim:   "address",
				Annots: []string{"123"},
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
		name  string
		value string
		res   string
	}{
		{
			name:  "Ligo Annotation",
			value: `{"prim":"address","annots":["%0"]}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Annotation",
			value: `{"prim":"address","annots":["%1"]}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Annotation",
			value: `{"prim":"address","annots":["%3"]}`,
			res:   LangLigo,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["%3s"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["%%3"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["%abc"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["%-42"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["abc"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Not Ligo",
			value: `{"prim":"address","annots":["0"]}`,
			res:   LangUnknown,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"GET_FORCE"}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"get_force"}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"MAP FIND"}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"start get_entrypoint end"}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"get_contract end"}`,
			res:   LangLigo,
		},
		{
			name:  "Ligo Value",
			value: `{"string":"failed assertion"}`,
			res:   LangLigo,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var node base.Node
			if err := json.UnmarshalFromString(tt.value, &node); err != nil {
				t.Errorf("UnmarshalFromString error=%s", err)
				return
			}
			if result := GetFromCode(&node); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestDetectLorentz(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		res   string
	}{
		{
			name:  "Lorentz Value",
			value: "UStore",
			res:   LangLorentz,
		},
		{
			name:  "Lorentz Value",
			value: "something UStore strange",
			res:   LangLorentz,
		},
		{
			name:  "Lorentz Value",
			value: "123 UStore",
			res:   LangLorentz,
		},
		{
			name:  "Not Lorentz Value",
			value: "start end",
			res:   LangUnknown,
		},
		{
			name:  "Not Lorentz Value",
			value: "ustore",
			res:   LangUnknown,
		},
		{
			name:  "Not Lorentz Value",
			value: "Ustore",
			res:   LangUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			node := &base.Node{StringValue: &tt.value}
			if result := GetFromCode(node); result != tt.res {
				t.Errorf("Invalid result.\nGot: %v\nExpected: %v", result, tt.res)
			}
		})
	}
}

func TestGetFromParameter(t *testing.T) {
	testCases := []struct {
		name string
		n    *base.Node
		res  string
	}{
		{
			name: "liquidity entrypoints",
			n: &base.Node{
				Annots: []string{"%_Liq_entry"},
			},
			res: LangLiquidity,
		},
		{
			name: "lorentz entrypoints",
			n: &base.Node{
				Annots: []string{"%epwBeginUpgrade"},
			},
			res: LangLorentz,
		},
		{
			name: "lorentz entrypoints",
			n: &base.Node{
				Annots: []string{"%epwApplyMigration"},
			},
			res: LangLorentz,
		},
		{
			name: "lorentz entrypoints",
			n: &base.Node{
				Annots: []string{"%epwSetCode"},
			},
			res: LangLorentz,
		},
		{
			name: "random entrypoints",
			n: &base.Node{
				Annots: []string{"%setCode"},
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
			var node base.Node
			if err := json.UnmarshalFromString(tt.input, &node); err != nil {
				t.Errorf("UnmarshalFromString error=%s", err)
				return
			}
			if got := GetFromFirstPrim(&node); got != tt.want {
				t.Errorf("GetFromFirstPrim invalid. expected: %v, got: %v", tt.want, got)
			}
		})
	}
}
