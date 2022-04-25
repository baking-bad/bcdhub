package search

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Prepare -
func Prepare(network types.Network, model models.Model) Data {
	switch val := model.(type) {
	case *contract.Contract:
		return NewContract(network, val)
	case *bigmapdiff.BigMapDiff:
		return NewBigMapDiff(network, val)
	case *operation.Operation:
		return NewOperation(network, val)
	case *tokenmetadata.TokenMetadata:
		return NewToken(network, val)
	case *contract_metadata.ContractMetadata:
		return NewMetadata(network, val)
	}

	return nil
}
