package tokenbalance

import (
	"math/big"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var parsers = map[string][]Parser{
	SingleAssetBalanceUpdates: {
		NewSingleAssetBalance(),
		NewSingleAssetUpdate(),
	},
	MultiAssetBalanceUpdates: {
		NewMultiAssetBalance(),
		NewMultiAssetUpdate(),
	},
}

// Parser -
type Parser interface {
	GetReturnType() *ast.TypedAst
	Parse(item []byte) ([]TokenBalance, error)
}

// TokenBalance -
type TokenBalance struct {
	Address string
	TokenID int64
	Value   *big.Int
}

// GetParser -
func GetParser(name string, returnType *ast.TypedAst) (Parser, error) {
	p, ok := parsers[NormalizeName(name)]
	if !ok {
		for _, ps := range parsers {
			item, err := findParser(ps, returnType)
			if err == nil {
				return item, nil
			}
		}
		return nil, errors.Wrap(ErrUnknownParser, name)
	}

	return findParser(p, returnType)
}

// GetParserForBigMap -
func GetParserForBigMap(returnType *ast.TypedAst) (Parser, error) {
	if returnType == nil {
		return nil, nil
	}
	bm := returnType.Nodes[0].(*ast.BigMap)
	var s strings.Builder
	s.WriteString(`{"prim":"map","args":[`)
	b, err := json.Marshal(bm.KeyType)
	if err != nil {
		return nil, err
	}
	if _, err := s.Write(b); err != nil {
		return nil, err
	}
	s.WriteByte(',')
	bValue, err := json.Marshal(bm.ValueType)
	if err != nil {
		return nil, err
	}
	if _, err := s.Write(bValue); err != nil {
		return nil, err
	}
	s.WriteString(`]}`)

	node, err := ast.NewTypedAstFromString(s.String())
	if err != nil {
		return nil, err
	}
	return GetParser("", node)
}

func findParser(p []Parser, returnType *ast.TypedAst) (Parser, error) {
	for i := range p {
		if returnType.EqualType(p[i].GetReturnType()) {
			return p[i], nil
		}
	}
	return nil, errors.Errorf("Invalid parser`s return type: %s", returnType)
}

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}
