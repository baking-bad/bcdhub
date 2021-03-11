package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"golang.org/x/crypto/blake2b"
)

// Pack -
func Pack(node Base) (string, error) {
	data, err := Forge(node, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", forge.PackPrefix, data), nil
}

// BigMapKeyHashFromNode -
func BigMapKeyHashFromNode(node Node) (string, error) {
	nodeBase, err := node.ToBaseNode(true)
	if err != nil {
		return "", err
	}

	return BigMapKeyHash(nodeBase)
}

// BigMapKeyHash -
func BigMapKeyHash(node *base.Node) (string, error) {
	data, err := forge.Forge(node)
	if err != nil {
		return "", err
	}
	blakeHash := blake2b.Sum256(append([]byte{forge.PackPrefixByte}, data...))
	return encoding.EncodeBase58(blakeHash[:], []byte(encoding.PrefixScriptExpr))
}

// BigMapKeyHashFromString -
func BigMapKeyHashFromString(str string) (string, error) {
	var node base.Node
	if err := json.UnmarshalFromString(str, &node); err != nil {
		return "", err
	}
	return BigMapKeyHash(&node)
}
