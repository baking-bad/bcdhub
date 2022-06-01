package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
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

	accounts         account.Repository
	contracts        contract.Repository
	protocols        protocol.Repository
	sanitizer        *bluemonday.Policy
	contractMetadata contract_metadata.Repository
}

// NewCache -
func NewCache(rpc noderpc.INode, accounts account.Repository, contracts contract.Repository, protocols protocol.Repository, cm contract_metadata.Repository, sanitizer *bluemonday.Policy) *Cache {
	return &Cache{
		ccache.New(ccache.Configure().MaxSize(1000)),
		rpc,
		accounts,
		contracts,
		protocols,
		sanitizer,
		cm,
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

		cm, err := cache.contractMetadata.Get(address)
		if err == nil {
			if cm.Name != consts.Unknown {
				return cm.Name, nil
			}
			return "", nil
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

// ContractMetadata -
func (cache *Cache) ContractMetadata(address string) (*contract_metadata.ContractMetadata, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}
	key := fmt.Sprintf("contract_metadata:%s", address)
	item, err := cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return cache.contractMetadata.Get(address)
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(*contract_metadata.ContractMetadata), nil
}

// Events -
func (cache *Cache) Events(address string) (contract_metadata.Events, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}
	key := fmt.Sprintf("contract_metadata:%s", address)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return cache.contractMetadata.Events(address)
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(contract_metadata.Events), nil
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

// ScriptBytes -
func (cache *Cache) ScriptBytes(address, symLink string) ([]byte, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("script_bytes:%s", address)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		script, err := cache.contracts.Script(address, symLink)
		if err != nil {
			return nil, err
		}
		return script.Full()
	})
	if err != nil {
		return nil, err
	}
	return item.Value().([]byte), nil
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
