package handlers

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
)

func (ctx *Context) getAlias(network, address string) string {
	key := fmt.Sprintf("alias:%s:%s", network, address)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TZIP.Get(network, address)
	})
	if err != nil {
		return ""
	}
	return item.Value().(string)
}

func (ctx *Context) getTokenMetadata(network, address string, tokenID uint64) (*tokenmetadata.TokenMetadata, error) {
	key := fmt.Sprintf("token_metadata:%s:%s:%d", network, address, tokenID)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TokenMetadata.GetOne(network, address, tokenID)
	})
	if err != nil {
		return nil, err
	}
	return item.Value().(*tokenmetadata.TokenMetadata), nil
}
