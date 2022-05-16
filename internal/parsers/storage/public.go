package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// Errors -
var (
	ErrInvalidPath          = errors.Errorf("Invalid path")
	ErrPathIsNotPointer     = errors.Errorf("Path is not pointer")
	ErrPointerAlreadyExists = errors.Errorf("Pointer already exists")
)

// Enrich -
func Enrich(storage *ast.TypedAst, bmd []bigmapdiff.BigMapDiff, skipEmpty, unpack bool) error {
	if len(bmd) == 0 {
		return nil
	}
	if !storage.IsSettled() {
		return ErrTreeIsNotSettled
	}

	data := prepareBigMapDiffsToEnrich(bmd, skipEmpty)
	return storage.EnrichBigMap(data)
}

// EnrichFromState -
func EnrichFromState(storage *ast.TypedAst, bmd []bigmapdiff.BigMapState, skipEmpty, unpack bool) error {
	if len(bmd) == 0 {
		return nil
	}
	if !storage.IsSettled() {
		return ErrTreeIsNotSettled
	}

	data := prepareBigMapStatesToEnrich(bmd, skipEmpty)
	return storage.EnrichBigMap(data)
}

// MakeStorageParser -
func MakeStorageParser(repo bigmapdiff.Repository, rpc noderpc.INode, protocol string) (Parser, error) {
	switch protocol {
	case
		"ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im",
		"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i",
		"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV",
		"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp",
		"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex",
		"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P",
		"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt",
		"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
		"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd",
		"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY":
		return NewAlpha(), nil
	case
		"PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx",
		"PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
		"PsiThaCaT47Zboaw71QWScM8sXeMM7bbQFncK9FLqYc6EKdpjVP",
		"Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
		"PrrUA9dCzbqBzugjQyw65HLHKjhH3HMFSLLHLZjj5rkmkG13Fej",
		"PsrsRVg1Gycjn5LvMtoYSQah1znvYmGp8bHLxwYLBZaYFf2CEkV",
		"PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
		"PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV",
		"PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY":
		return NewLazyBabylon(repo, rpc), nil
	case
		"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS",
		"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU",
		"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN",
		"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
		"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq",
		"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA":
		return NewBabylon(repo, rpc), nil
	default:
		return nil, errors.Errorf("unknown protocol for storage parser: %s", protocol)
	}
}
