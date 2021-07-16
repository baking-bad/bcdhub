package tokenbalance

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var parsersEvents = map[string][]Parser{
	SingleAssetBalanceUpdates: {
		NewSingleAssetBalance(),
		NewSingleAssetUpdate(),
	},
	MultiAssetBalanceUpdates: {
		NewMultiAssetBalance(),
		NewMultiAssetUpdate(),
	},
	NftAssetBalanceUpdates: {
		NewNftAssetOption(),
	},
}

var parsersBigMap = map[string][]Parser{
	SingleAssetBalanceUpdates: {
		NewSingleAssetBalance(),
		NewSingleAssetUpdate(),
	},
	MultiAssetBalanceUpdates: {
		NewMultiAssetBalance(),
		NewMultiAssetUpdate(),
	},
	NftAssetBalanceUpdates: {
		NewNftAsset(),
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
	TokenID uint64
	Value   decimal.Decimal

	IsExclusiveNFT bool
}

// GetParserForEvents -
func GetParserForEvents(name string, returnType *ast.TypedAst) (Parser, error) {
	return getParser(parsersEvents, name, returnType)
}

func getParser(parsers map[string][]Parser, name string, returnType *ast.TypedAst) (Parser, error) {
	if parsers == nil {
		return nil, errors.Wrap(ErrUnknownParser, name)
	}

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
func GetParserForBigMap(keyType, valueType []byte) (Parser, error) {
	if keyType == nil && valueType == nil {
		return nil, nil
	}
	var s strings.Builder
	s.WriteString(`{"prim":"map","args":[`)
	if _, err := s.Write(keyType); err != nil {
		return nil, err
	}
	s.WriteByte(',')
	if _, err := s.Write(valueType); err != nil {
		return nil, err
	}
	s.WriteString(`]}`)

	node, err := ast.NewTypedAstFromString(s.String())
	if err != nil {
		return nil, err
	}
	return getParser(parsersBigMap, "", node)
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

func getMap(retType *ast.TypedAst, data []byte) (*ast.Map, error) {
	var node ast.UntypedAST
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	newNode := ast.Copy(retType.Nodes[0])
	if err := newNode.ParseValue(node[0]); err != nil {
		return nil, err
	}

	return newNode.(*ast.Map), nil
}
