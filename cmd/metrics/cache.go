package main

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

func getStorageType(bmd bigmapdiff.BigMapDiff) (*ast.TypedAst, error) {
	item, err := ctx.cache.Fetch(fmt.Sprintf("%s:%s", bmd.Network, bmd.Contract), time.Minute, func() (interface{}, error) {
		data, err := fetch.Contract(bmd.Contract, bmd.Network, bmd.Protocol, ctx.SharePath)
		if err != nil {
			return nil, err
		}
		script, err := ast.NewScript(data)
		if err != nil {
			return nil, err
		}

		return script.StorageType()
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(*ast.TypedAst), nil
}
