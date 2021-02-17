package contract

import (
	jsoniter "github.com/json-iterator/go"

	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type contractData struct {
	Code    stdJSON.RawMessage `json:"code"`
	Storage stdJSON.RawMessage `json:"storage"`
}

// Parser -
type Parser struct {
	Code    *ast.Script
	Storage ast.UntypedAST

	Language           types.Set
	FailStrings        types.Set
	Tags               types.Set
	Annotations        types.Set
	HardcodedAddresses types.Set
	Hash               string
}

// NewParser -
func NewParser(data []byte) (*Parser, error) {
	var cd contractData
	if err := json.Unmarshal(data, &cd); err != nil {
		return nil, err
	}

	script, err := ast.NewScript(cd.Code)
	if err != nil {
		return nil, err
	}

	var storage ast.UntypedAST
	if err := json.Unmarshal(cd.Storage, &storage); err != nil {
		return nil, err
	}

	hardcoded, err := findHardcodedAddresses(cd.Code)
	if err != nil {
		return nil, err
	}

	hash, err := computeHash(cd.Code)
	if err != nil {
		return nil, err
	}

	return &Parser{
		Code:               script,
		Storage:            storage,
		Language:           make(types.Set),
		FailStrings:        make(types.Set),
		Tags:               make(types.Set),
		Annotations:        make(types.Set),
		HardcodedAddresses: hardcoded,
		Hash:               hash,
	}, nil
}

// Parse -
func (p *Parser) Parse() error {
	return nil
}
