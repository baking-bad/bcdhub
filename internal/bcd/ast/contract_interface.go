package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast/interfaces"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// contract tags
const (
	ContractTagFA1           = "fa1"
	ContractTagFA1_2         = "fa1-2"
	ContractTagFA2           = "fa2"
	ContractTagViewNat       = "view_nat"
	ContractTagViewAddress   = "view_address"
	ContractTagViewBalanceOf = "view_balance_of"
)

type contractInterface struct {
	Entrypoints map[string]Node
	IsRoot      bool
}

var interfaceTrees = map[string]contractInterface{}

// FindContractInterfaces -
func FindContractInterfaces(tree *TypedAst) []string {
	if initInterfaceTrees() != nil {
		return nil
	}
	tags := make([]string, 0)
	for tag := range interfaceTrees {
		if FindContractInterface(tree, tag) {
			tags = append(tags, tag)
		}
	}
	return tags
}

func findViewContractInterfaces(tree *TypedAst) []string {
	if initInterfaceTrees() != nil {
		return nil
	}
	tags := make([]string, 0)
	for _, tag := range []string{ContractTagViewNat, ContractTagViewAddress, ContractTagViewBalanceOf} {
		if FindContractInterface(tree, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

// FindContractInterface -
func FindContractInterface(tree *TypedAst, name string) bool {
	if initInterfaceTrees() != nil {
		return false
	}
	if contract, ok := interfaceTrees[name]; ok {
		return findEntrypoints(tree, contract, nil)
	}
	return false
}

func findEntrypoints(tree *TypedAst, ci contractInterface, exists map[string]struct{}) bool {
	if ci.IsRoot {
		if len(tree.Nodes) != 1 || len(ci.Entrypoints) != 1 {
			return false
		}
		return tree.Nodes[0].EqualType(ci.Entrypoints[consts.DefaultEntrypoint])
	}

	if exists == nil {
		exists = make(map[string]struct{})
	}

	for i := range tree.Nodes {
		if tree.Nodes[i].IsPrim(consts.OR) {
			or := tree.Nodes[i].(*Or)
			orTree := &TypedAst{
				Nodes: []Node{or.LeftType, or.RightType},
			}
			if findEntrypoints(orTree, ci, exists) {
				return true
			}
			continue
		}

		for name, subTree := range ci.Entrypoints {
			if _, ok := exists[name]; !ok && tree.Nodes[i].EqualType(subTree) {
				exists[name] = struct{}{}
			}
		}

		if len(exists) == len(ci.Entrypoints) {
			return true
		}
	}

	return false
}

func initInterfaceTrees() error {
	if len(interfaceTrees) > 0 {
		return nil
	}

	all, err := interfaces.GetAll()
	if err != nil {
		return err
	}
	for name, data := range all {
		ci := contractInterface{
			Entrypoints: make(map[string]Node),
			IsRoot:      data.IsRoot,
		}

		for key, str := range data.Entrypoints {
			var tree UntypedAST
			if err := json.Unmarshal(str, &tree); err != nil {
				return err
			}
			t, err := tree.ToTypedAST()
			if err != nil {
				return err
			}
			ci.Entrypoints[key] = t.Nodes[0]
		}
		interfaceTrees[name] = ci
	}
	return nil
}
