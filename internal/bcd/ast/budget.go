package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// MaxInstantiatedNodes - upper bound for the number of type nodes
// FromJSONSchema may copy while filling a single input.
//
// Collections (list, set, map, big_map) copy their element type once per
// element. Both the element type size and the element count are
// attacker-controlled on API endpoints, and an element can be as small as
// `{}`, so the copies are paid from a shared budget. Copying is
// reflection-based (about a microsecond per node), so the budget also bounds
// the CPU and memory a rejected request can burn.
const MaxInstantiatedNodes = 1_000_000

// nodeBudget is shared by all collections of one tree being filled from JSON.
type nodeBudget struct {
	left  int
	types map[Node]elementType
}

type elementType struct {
	size           int
	hasCollections bool
}

func newNodeBudget() *nodeBudget {
	return &nodeBudget{left: MaxInstantiatedNodes}
}

// instantiate copies an element type of a collection, paying for it from the
// budget. Collections inside the copy share the same budget. Without a budget
// (trees not filled from JSON) it is a plain Copy.
func (b *nodeBudget) instantiate(typ Node) (Node, error) {
	if b == nil {
		return Copy(typ), nil
	}
	info, ok := b.types[typ]
	if !ok {
		info = elementType{size: typeSize(typ), hasCollections: hasCollections(typ)}
		if b.types == nil {
			b.types = make(map[Node]elementType)
		}
		b.types[typ] = info
	}
	if info.size > b.left {
		return nil, errors.Wrapf(consts.ErrValidation, "input is too large: it would instantiate more than %d nodes", MaxInstantiatedNodes)
	}
	b.left -= info.size
	node := Copy(typ)
	if info.hasCollections {
		attachBudget(node, b)
	}
	return node, nil
}

func hasCollections(node Node) bool {
	switch node.(type) {
	case *List, *Set, *Map, *BigMap:
		return true
	}
	for _, child := range childTypes(node) {
		if hasCollections(child) {
			return true
		}
	}
	return false
}

// attachBudget gives every collection in the type tree the budget to pay for
// its element copies from.
func attachBudget(node Node, b *nodeBudget) {
	switch t := node.(type) {
	case *List:
		t.budget = b
	case *Set:
		t.budget = b
	case *Map:
		t.budget = b
	case *BigMap:
		t.budget = b
	}
	for _, child := range childTypes(node) {
		attachBudget(child, b)
	}
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
