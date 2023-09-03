package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/karlseguin/ccache"
	"github.com/microcosm-cc/bluemonday"
)

// Cache -
type Cache struct {
	*ccache.Cache
	rpc noderpc.INode

	accounts  account.Repository
	contracts contract.Repository
	protocols protocol.Repository
	sanitizer *bluemonday.Policy
}

// NewCache -
func NewCache(rpc noderpc.INode, accounts account.Repository, contracts contract.Repository, protocols protocol.Repository) *Cache {
	sanitizer := bluemonday.UGCPolicy()
	sanitizer.AllowAttrs("em")
	return &Cache{
		ccache.New(ccache.Configure().MaxSize(1000)),
		rpc,
		accounts,
		contracts,
		protocols,
		sanitizer,
	}
}

// Alias -
func (cache *Cache) Alias(address string) string {
	if !bcd.IsContract(address) {
		return ""
	}
	key := fmt.Sprintf("alias:%s", address)
	item, err := cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		acc, err := cache.accounts.Get(address)
		if err == nil && acc.Alias != "" {
			return acc.Alias, nil
		}

		return "", err
	})
	if err != nil {
		return ""
	}

	if data, ok := item.Value().(string); ok && data != "" {
		return cache.sanitizer.Sanitize(data)
	}
	return ""
}

// ContractTags -
func (cache *Cache) ContractTags(address string) (types.Tags, error) {
	if !bcd.IsContract(address) {
		return 0, nil
	}

	key := fmt.Sprintf("contract:%s", address)
	item, err := cache.Fetch(key, time.Minute*10, func() (interface{}, error) {
		c, err := cache.contracts.Get(address)
		if err != nil {
			return 0, err
		}
		return c.Tags, nil
	})
	if err != nil {
		return 0, err
	}
	return item.Value().(types.Tags), nil
}

// TezosBalance -
func (cache *Cache) TezosBalance(ctx context.Context, address string, level int64) (int64, error) {
	key := fmt.Sprintf("tezos_balance:%s:%d", address, level)
	item, err := cache.Fetch(key, 30*time.Second, func() (interface{}, error) {
		return cache.rpc.GetContractBalance(ctx, address, level)
	})
	if err != nil {
		return 0, err
	}
	return item.Value().(int64), nil
}

// StorageTypeBytes -
func (cache *Cache) StorageTypeBytes(address, symLink string) ([]byte, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("storage:%s", address)
	item, err := cache.Fetch(key, 5*time.Minute, func() (interface{}, error) {
		return cache.contracts.ScriptPart(address, symLink, consts.STORAGE)
	})
	if err != nil {
		return nil, err
	}
	return item.Value().([]byte), nil
}

// ProtocolByID -
func (cache *Cache) ProtocolByID(id int64) (protocol.Protocol, error) {
	key := fmt.Sprintf("protocol_id:%d", id)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return cache.protocols.GetByID(id)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value().(protocol.Protocol), nil
}
