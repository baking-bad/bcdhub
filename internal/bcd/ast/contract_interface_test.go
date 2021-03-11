package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindContractInterface(t *testing.T) {
	tests := []struct {
		name          string
		tree          string
		interfaceName string
		want          bool
	}{
		{
			name:          "view_nat",
			tree:          `{"prim":"nat"}`,
			interfaceName: "view_nat",
			want:          true,
		}, {
			name:          "not view_nat",
			tree:          `{"prim":"string"}`,
			interfaceName: "view_nat",
			want:          false,
		}, {
			name:          "view_address",
			tree:          `{"prim":"address"}`,
			interfaceName: "view_address",
			want:          true,
		}, {
			name:          "not view_address",
			tree:          `{"prim":"string"}`,
			interfaceName: "view_address",
			want:          false,
		}, {
			name:          "view_balance_of",
			tree:          `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%request"]},{"prim":"nat","annots":["%balance"]}]}]}`,
			interfaceName: "view_balance_of",
			want:          true,
		}, {
			name:          "not view_balance_of",
			tree:          `{"prim":"string"}`,
			interfaceName: "view_balance_of",
			want:          false,
		}, {
			name:          "fa1",
			tree:          `{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":["%approve"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":spender"]}]},{"prim":"contract","args":[{"prim":"nat","annots":[":remaining"]}]}],"annots":["%getAllowance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"contract","args":[{"prim":"nat","annots":[":balance"]}]}],"annots":["%getBalance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat","annots":[":totalSupply"]}]}],"annots":["%getTotalSupply"]},{"prim":"or","args":[{"prim":"bool","annots":["%setPause"]},{"prim":"or","args":[{"prim":"address","annots":["%setAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address","annots":[":administrator"]}]}],"annots":["%getAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}],"annots":["%mint"]},{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"nat","annots":[":value"]}],"annots":["%burn"]}]}]}]}]}]}]}]}]}]}`,
			interfaceName: "fa1",
			want:          true,
		}, {
			name:          "not fa1",
			tree:          `{"prim":"string"}`,
			interfaceName: "fa1",
			want:          false,
		}, {
			name:          "fa1-2",
			tree:          `{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":["%approve"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":spender"]}]},{"prim":"contract","args":[{"prim":"nat","annots":[":remaining"]}]}],"annots":["%getAllowance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"contract","args":[{"prim":"nat","annots":[":balance"]}]}],"annots":["%getBalance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat","annots":[":totalSupply"]}]}],"annots":["%getTotalSupply"]},{"prim":"or","args":[{"prim":"bool","annots":["%setPause"]},{"prim":"or","args":[{"prim":"address","annots":["%setAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address","annots":[":administrator"]}]}],"annots":["%getAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}],"annots":["%mint"]},{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"nat","annots":[":value"]}],"annots":["%burn"]}]}]}]}]}]}]}]}]}]}`,
			interfaceName: "fa1-2",
			want:          true,
		}, {
			name:          "not fa1-2",
			tree:          `{"prim":"string"}`,
			interfaceName: "fa1-2",
			want:          false,
		}, {
			name:          "fa2",
			tree:          `{"prim":"or","args":[{"prim":"or","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}]}],"annots":["%txs"]}]}],"annots":["%transfer"]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%requests"]},{"prim":"contract","args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}],"annots":["%request"]},{"prim":"nat","annots":["%balance"]}]}]}],"annots":["%callback"]}],"annots":["%balance_of"]}]},{"prim":"or","args":[{"prim":"contract","args":[{"prim":"address"}],"annots":["%token_metadata_registry"]},{"prim":"list","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%add_operator"]},{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}],"annots":["%remove_operator"]}]}],"annots":["%update_operators"]}]}]}`,
			interfaceName: "fa2",
			want:          true,
		}, {
			name:          "not fa2",
			tree:          `{"prim":"string"}`,
			interfaceName: "fa2",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UnmarshalFromString() error = %v", err)
				return
			}

			typedTree, err := tree.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}

			if got := FindContractInterface(typedTree, tt.interfaceName); got != tt.want {
				t.Errorf("FindContractInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindContractInterfaces(t *testing.T) {
	tests := []struct {
		name string
		tree string
		want []string
	}{
		{
			name: "fa1, fa1-2",
			tree: `{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}],"annots":["%transfer"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}],"annots":["%approve"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"address","annots":[":spender"]}]},{"prim":"contract","args":[{"prim":"nat","annots":[":remaining"]}]}],"annots":["%getAllowance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":owner"]},{"prim":"contract","args":[{"prim":"nat","annots":[":balance"]}]}],"annots":["%getBalance"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"nat","annots":[":totalSupply"]}]}],"annots":["%getTotalSupply"]},{"prim":"or","args":[{"prim":"bool","annots":["%setPause"]},{"prim":"or","args":[{"prim":"address","annots":["%setAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"unit"},{"prim":"contract","args":[{"prim":"address","annots":[":administrator"]}]}],"annots":["%getAdministrator"]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}],"annots":["%mint"]},{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"nat","annots":[":value"]}],"annots":["%burn"]}]}]}]}]}]}]}]}]}]}`,
			want: []string{"fa1-2", "fa1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree UntypedAST
			if err := json.UnmarshalFromString(tt.tree, &tree); err != nil {
				t.Errorf("UnmarshalFromString() error = %v", err)
				return
			}

			typedTree, err := tree.ToTypedAST()
			if err != nil {
				t.Errorf("ToTypedAST() error = %v", err)
				return
			}
			got := FindContractInterfaces(typedTree)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
