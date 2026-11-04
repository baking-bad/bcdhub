package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// MaxInstantiatedNodes - upper bound for the number of nodes FromJSONSchema
// may instantiate for a single input.
//
// Collections (list, set, map, big_map) copy their element type once per
// element, so the work is the element type size multiplied by the element
// count. Both are attacker-controlled on API endpoints, and an element can be
// as small as `{}`, so the product is checked before any copying starts.
const MaxInstantiatedNodes = 1_000_000

// checkInstantiationBudget estimates, from above, how many nodes filling typ
// from data would instantiate: every array element in data becomes at most one
// copy of the largest collection element type in typ.
func checkInstantiationBudget(typ Node, data any) error {
	elementSize := maxElementTypeSize(typ)
	if elementSize == 0 {
		return nil
	}
	elements := countArrayElements(data)
	if elements > 0 && elementSize > MaxInstantiatedNodes/elements {
		return errors.Wrapf(consts.ErrValidation, "input is too large: %d elements of up to %d nodes each exceed the limit of %d nodes", elements, elementSize, MaxInstantiatedNodes)
	}
	return nil
}

// childTypes returns the type nodes directly nested in node.
func childTypes(node Node) []Node {
	switch t := node.(type) {
	case *Pair:
		return t.Args
	case *SectionType:
		return t.Args
	case *Or:
		return []Node{t.LeftType, t.RightType}
	case *Map:
		return []Node{t.KeyType, t.ValueType}
	case *BigMap:
		return []Node{t.KeyType, t.ValueType}
	case *Lambda:
		return []Node{t.InputType, t.ReturnType}
	case *List:
		return []Node{t.Type}
	case *Set:
		return []Node{t.Type}
	case *Option:
		return []Node{t.Type}
	case *Contract:
		return []Node{t.Type}
	case *Ticket:
		return []Node{t.Type, t.PairedType}
	default:
		return nil
	}
}

// typeSize counts the nodes of a type tree, i.e. the cost of one Copy of it.
func typeSize(node Node) int {
	if node == nil {
		return 0
	}
	size := 1
	for _, child := range childTypes(node) {
		size += typeSize(child)
	}
	return size
}

// maxElementTypeSize returns the size of the largest element type of any
// collection in the type tree (key and value together for maps).
func maxElementTypeSize(node Node) int {
	if node == nil {
		return 0
	}
	var size int
	switch t := node.(type) {
	case *List:
		size = typeSize(t.Type)
	case *Set:
		size = typeSize(t.Type)
	case *Map:
		size = typeSize(t.KeyType) + typeSize(t.ValueType)
	case *BigMap:
		size = typeSize(t.KeyType) + typeSize(t.ValueType)
	}
	for _, child := range childTypes(node) {
		size = max(size, maxElementTypeSize(child))
	}
	return size
}

// countArrayElements counts elements of all JSON arrays nested in data.
func countArrayElements(data any) int {
	switch v := data.(type) {
	case []any:
		count := len(v)
		for i := range v {
			count += countArrayElements(v[i])
		}
		return count
	case map[string]any:
		var count int
		for _, item := range v {
			count += countArrayElements(item)
		}
		return count
	default:
		return 0
	}
}
