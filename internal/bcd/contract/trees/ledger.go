package trees

import "github.com/baking-bad/bcdhub/internal/bcd/ast"

var (
	NewNftLedgerSingleAsset, _ = ast.NewTypedAstFromString(`{"prim":"big_map","args":[{"prim":"address"},{"prim":"nat"}]}`)
	NewNftLedgerAsset, _       = ast.NewTypedAstFromString(`{"prim":"big_map","args":[{"prim":"nat"},{"prim":"address"}]}`)
	NewNftLedgerMultiAsset, _  = ast.NewTypedAstFromString(`{"prim":"big_map","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}`)
)
