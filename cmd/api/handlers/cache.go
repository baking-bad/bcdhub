package handlers

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

func (ctx *Context) getAlias(network, address string) string {
	key := fmt.Sprintf("alias:%s:%s", network, address)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TZIP.Get(network, address)
	})
	if err != nil {
		return ""
	}
	return item.Value().(*tzip.TZIP).Name
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

func (ctx *Context) getCurrentBlock(network string) (block.Block, error) {
	key := fmt.Sprintf("block:%s", network)
	item, err := ctx.Cache.Fetch(key, time.Second*15, func() (interface{}, error) {
		return ctx.Blocks.Last(network)
	})
	if err != nil {
		return block.Block{}, err
	}
	return item.Value().(block.Block), nil
}

func (ctx *Context) getTezosBalance(network, address string, level int64) (int64, error) {
	key := fmt.Sprintf("tezos_balance:%s:%s:%d", network, address, level)
	item, err := ctx.Cache.Fetch(key, 30*time.Second, func() (interface{}, error) {
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return 0, err
		}
		return rpc.GetContractBalance(address, level)
	})
	if err != nil {
		return 0, err
	}
	return item.Value().(int64), nil
}
