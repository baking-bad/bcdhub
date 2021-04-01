package search

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Prepare -
func Prepare(items []models.Model) []Data {
	data := make([]Data, 0)

	for i := range items {
		switch val := items[i].(type) {
		case *contract.Contract:
			var c Contract
			c.Prepare(val)
			data = append(data, &c)
		case *bigmapdiff.BigMapDiff:
			var bmd BigMapDiff
			bmd.Prepare(val)
			data = append(data, &bmd)
		case *tezosdomain.TezosDomain:
			var td Domain
			td.Prepare(val)
			data = append(data, &td)
		case *operation.Operation:
			var op Operation
			op.Prepare(val)
			data = append(data, &op)
		case *tokenmetadata.TokenMetadata:
			var token Token
			token.Prepare(val)
			data = append(data, &token)
		case *tzip.TZIP:
			var m Metadata
			m.Prepare(val)
			data = append(data, &m)
		}
	}

	return data
}
