package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	ccache "github.com/karlseguin/ccache/v3"
	"github.com/microcosm-cc/bluemonday"
)

// Cache -
type Cache struct {
	rpc noderpc.INode

	accounts  account.Repository
	contracts contract.Repository
	protocols protocol.Repository
	sanitizer *bluemonday.Policy

	contractTags *ccache.Cache[types.Tags]
	balances     *ccache.Cache[int64]
	storageType  *ccache.Cache[[]byte]
	scriptBytes  *ccache.Cache[[]byte]
	scripts      *ccache.Cache[contract.Script]
	protocolById *ccache.Cache[protocol.Protocol]
}

// NewCache -
func NewCache(rpc noderpc.INode, accounts account.Repository, contracts contract.Repository, protocols protocol.Repository) *Cache {
	sanitizer := bluemonday.UGCPolicy()
	sanitizer.AllowAttrs("em")
	return &Cache{
		contractTags: ccache.New(ccache.Configure[types.Tags]().MaxSize(1000)),
		balances:     ccache.New(ccache.Configure[int64]().MaxSize(1000)),
		storageType:  ccache.New(ccache.Configure[[]byte]().MaxSize(1000)),
		scriptBytes:  ccache.New(ccache.Configure[[]byte]().MaxSize(1000)),
		scripts:      ccache.New(ccache.Configure[contract.Script]().MaxSize(1000)),
		protocolById: ccache.New(ccache.Configure[protocol.Protocol]().MaxSize(1000)),
		rpc:          rpc,
		accounts:     accounts,
		contracts:    contracts,
		protocols:    protocols,
		sanitizer:    sanitizer,
	}
}

// ContractTags -
func (cache *Cache) ContractTags(ctx context.Context, address string) (types.Tags, error) {
	if !bcd.IsContract(address) {
		return 0, nil
	}

	item, err := cache.contractTags.Fetch(address, time.Minute*10, func() (types.Tags, error) {
		c, err := cache.contracts.Get(ctx, address)
		if err != nil {
			return 0, err
		}
		return c.Tags, nil
	})
	if err != nil {
		cache.contractTags.Delete(address)
		return 0, err
	}
	return item.Value(), nil
}

// TezosBalance -
func (cache *Cache) TezosBalance(ctx context.Context, address string, level int64) (int64, error) {
	key := fmt.Sprintf("%s:%d", address, level)
	item, err := cache.balances.Fetch(key, 10*time.Second, func() (int64, error) {
		return cache.rpc.GetContractBalance(ctx, address, level)
	})
	if err != nil {
		cache.balances.Delete(key)
		return 0, err
	}
	return item.Value(), nil
}

// StorageTypeBytes -
func (cache *Cache) StorageTypeBytes(ctx context.Context, address, symLink string) ([]byte, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("%s:%s", address, symLink)
	item, err := cache.storageType.Fetch(key, 5*time.Minute, func() ([]byte, error) {
		return cache.contracts.ScriptPart(ctx, address, symLink, consts.STORAGE)
	})
	if err != nil {
		cache.storageType.Delete(key)
		return nil, err
	}
	return item.Value(), nil
}

// ProtocolByID -
func (cache *Cache) ProtocolByID(ctx context.Context, id int64) (protocol.Protocol, error) {
	item, err := cache.protocolById.Fetch(strconv.FormatInt(id, 10), time.Hour, func() (protocol.Protocol, error) {
		return cache.protocols.GetByID(ctx, id)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value(), nil
}

func (cache *Cache) Script(ctx context.Context, address, symLink string) (contract.Script, error) {
	key := fmt.Sprintf("%s:%s", address, symLink)
	item, err := cache.scripts.Fetch(key, time.Hour, func() (contract.Script, error) {
		return cache.contracts.Script(ctx, address, symLink)
	})
	if err != nil {
		cache.scripts.Delete(key)
		return contract.Script{}, err
	}
	return item.Value(), nil
}

func (cache *Cache) ScriptBytes(ctx context.Context, address, symLink string) ([]byte, error) {
	key := fmt.Sprintf("%s:%s", address, symLink)
	item, err := cache.scriptBytes.Fetch(key, time.Hour, func() ([]byte, error) {
		script, err := cache.contracts.Script(ctx, address, symLink)
		if err != nil {
			return nil, err
		}
		return script.Full()
	})
	if err != nil {
		cache.scriptBytes.Delete(key)
		return nil, err
	}
	return item.Value(), nil
}
