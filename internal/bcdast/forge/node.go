package forge

import "math/big"

// Node -
type Node struct {
	Prim   string
	Args   []*Node
	Annots []string

	StringValue *string
	BytesValue  *string
	IntValue    *big.Int
}
