package main

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
)

func saveSearchModels(searcher search.Searcher, items []models.Model) error {
	data := search.Prepare(items)

	for i := range data {
		switch typ := data[i].(type) {
		case *search.Contract:
			typ.Alias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Address)
			typ.DelegateAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Delegate)
		case *search.Operation:
			typ.SourceAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Source)
			typ.DestinationAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Destination)
			typ.DelegateAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Delegate)
		}
	}

	return searcher.Save(data)
}
