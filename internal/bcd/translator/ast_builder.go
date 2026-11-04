package translator

import (
	"github.com/pkg/errors"
	"github.com/yhirose/go-peg"
)

// enableAst makes the parser build the same AST as peg.Parser.EnableAst, but
// without line/column info. go-peg computes it for every node by scanning the
// input from the very beginning, which makes parsing quadratic in the input
// size. The translator never reads Ln/Col.
func enableAst(parser *peg.Parser) error {
	if err := parser.EnableAst(); err != nil {
		return err
	}

	// Whether a rule is a token is not exported by go-peg: probe the action
	// EnableAst installed. A token action ignores child values, a non-token
	// one turns them into child nodes.
	probe := &peg.Values{Vs: []peg.Any{&peg.Ast{}}}
	for name, rule := range parser.Grammar {
		res, err := rule.Action(probe, nil)
		if err != nil {
			return err
		}
		ast, ok := res.(*peg.Ast)
		if !ok {
			return errors.Errorf("unexpected AST action result for rule %s: %T", name, res)
		}

		if len(ast.Nodes) == 0 {
			rule.Action = tokenAction(name)
		} else {
			rule.Action = nodeAction(name)
		}
	}
	return nil
}

func tokenAction(name string) peg.Action {
	return func(v *peg.Values, _ peg.Any) (peg.Any, error) {
		return &peg.Ast{S: v.S, Name: name, Token: v.Token()}, nil
	}
}

func nodeAction(name string) peg.Action {
	return func(v *peg.Values, _ peg.Any) (peg.Any, error) {
		ast := &peg.Ast{S: v.S, Name: name}
		if len(v.Vs) > 0 {
			ast.Nodes = make([]*peg.Ast, 0, len(v.Vs))
		}
		for _, val := range v.Vs {
			node, ok := val.(*peg.Ast)
			if !ok {
				return nil, errors.Errorf("unexpected value in rule %s: %T", name, val)
			}
			node.Parent = ast
			ast.Nodes = append(ast.Nodes, node)
		}
		return ast, nil
	}
}
