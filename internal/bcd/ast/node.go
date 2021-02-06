package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

func toBaseNodeInt(val *base.BigInt) *base.Node {
	return &base.Node{
		IntValue: val,
	}
}

func toBaseNodeString(val string) *base.Node {
	return &base.Node{
		StringValue: &val,
	}
}

func toBaseNodeBytes(val string) *base.Node {
	return &base.Node{
		BytesValue: &val,
	}
}

func mapToBaseNodes(data map[Node]Node, optimized bool) (*base.Node, error) {
	node := new(base.Node)
	node.Prim = base.PrimArray
	node.Args = make([]*base.Node, 0)
	for key, value := range data {
		keyNode, err := key.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		valueNode, err := value.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = append(node.Args, &base.Node{
			Prim: consts.Elt,
			Args: []*base.Node{
				keyNode, valueNode,
			},
		})
	}
	return node, nil
}

func arrayToBaseNode(data []Node, optimized bool) (*base.Node, error) {
	node := new(base.Node)
	node.Prim = base.PrimArray
	node.Args = make([]*base.Node, 0)
	for i := range data {
		arg, err := data[i].ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = append(node.Args, arg)
	}
	return node, nil
}
