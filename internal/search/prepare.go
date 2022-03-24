package search

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Prepare -
func Prepare(network types.Network, items []models.Model) []Data {
	data := make([]Data, 0)

	for i := range items {
		switch val := items[i].(type) {
		case *contract.Contract:
			var c Contract
			c.Prepare(network, val)
			data = append(data, &c)
		case *bigmapdiff.BigMapDiff:
			var bmd BigMapDiff
			bmd.Prepare(network, val)
			data = append(data, &bmd)
		case *operation.Operation:
			var op Operation
			op.Prepare(network, val)
			data = append(data, &op)
		case *tokenmetadata.TokenMetadata:
			var token Token
			token.Prepare(network, val)
			data = append(data, &token)
		case *cm.ContractMetadata:
			var m Metadata
			m.Prepare(network, val)
			data = append(data, &m)
		}
	}

	return data
}
