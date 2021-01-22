package forge

import "math/big"

// Node -
type Node struct {
	Prim   string
	Args   []Node
	Annots []string

	StringValue *string
	BytesValue  *string
	IntValue    *big.Int

	argsCount int
	hasAnnots bool
}

func newNode(argsCount int, hasAnnots bool) *Node {
	return &Node{
		argsCount: argsCount,
		hasAnnots: hasAnnots,
	}
}

// Unforge -
func (n *Node) Unforge(data []byte) (int, error) {
	return 0, nil
}
