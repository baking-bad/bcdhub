package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/contract/trees"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Tag -
func Tag(bigMap *bigmap.BigMap) {
	var key ast.UntypedAST
	if err := json.Unmarshal(bigMap.KeyType, &key); err != nil {
		return
	}
	keyType, err := key.ToTypedAST()
	if err != nil {
		return
	}
	var value ast.UntypedAST
	if err := json.Unmarshal(bigMap.ValueType, &value); err != nil {
		return
	}
	valueType, err := value.ToTypedAST()
	if err != nil {
		return
	}

	ledger(bigMap, keyType, valueType)
	contractMetadata(bigMap, keyType, valueType)
	tokenMetadata(bigMap, keyType, valueType)
}

func ledger(bigMap *bigmap.BigMap, keyType, valueType *ast.TypedAst) {
	if bigMap.Name != "ledger" {
		return
	}

	if keyType.EqualType(trees.Nat) && valueType.EqualType(trees.Address) {
		bigMap.Tags.Set(types.LedgerTag)
		return
	}

	if keyType.EqualType(trees.Address) && valueType.EqualType(trees.Nat) {
		bigMap.Tags.Set(types.LedgerTag)
		return
	}

	if keyType.EqualType(trees.Token) && valueType.EqualType(trees.Nat) {
		bigMap.Tags.Set(types.LedgerTag)
		return
	}
}

func contractMetadata(bigMap *bigmap.BigMap, keyType, valueType *ast.TypedAst) {
	if bigMap.Name != "metadata" {
		return
	}

	if keyType.EqualType(trees.String) && valueType.EqualType(trees.Bytes) {
		bigMap.Tags.Set(types.ContractMetadataTag)
		return
	}
}

func tokenMetadata(bigMap *bigmap.BigMap, keyType, valueType *ast.TypedAst) {
	if bigMap.Name != "token_metadata" {
		return
	}

	if keyType.EqualType(trees.Nat) && valueType.EqualType(trees.TokenMetadata) {
		bigMap.Tags.Set(types.TokenMetadataTag)
		return
	}
}
