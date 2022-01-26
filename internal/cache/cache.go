package cache

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/karlseguin/ccache"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
)

// Cache -
type Cache struct {
	*ccache.Cache
	rpc map[types.Network]noderpc.INode

	blocks    block.Repository
	accounts  account.Repository
	contracts contract.Repository
	protocols protocol.Repository
	sanitizer *bluemonday.Policy
	tzip      contract_metadata.Repository
}

// NewCache -
func NewCache(rpc map[types.Network]noderpc.INode, blocks block.Repository, accounts account.Repository, contracts contract.Repository, protocols protocol.Repository, cm contract_metadata.Repository, sanitizer *bluemonday.Policy) *Cache {
	return &Cache{
		ccache.New(ccache.Configure().MaxSize(100000)),
		rpc,
		blocks,
		accounts,
		contracts,
		protocols,
		sanitizer,
		cm,
	}
}

// Alias -
func (cache *Cache) Alias(network types.Network, address string) string {
	if !bcd.IsContract(address) {
		return ""
	}
	key := fmt.Sprintf("alias:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		acc, err := cache.accounts.Get(network, address)
		if err == nil && acc.Alias != "" {
			return acc.Alias, nil
		}

		cm, err := cache.tzip.Get(network, address)
		if err == nil {
			return cm.Name, nil
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
func (cache *Cache) ContractMetadata(network types.Network, address string) (*contract_metadata.ContractMetadata, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}
	key := fmt.Sprintf("contract_metadata:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return cache.tzip.Get(network, address)
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(*contract_metadata.ContractMetadata), nil
}

// Events -
func (cache *Cache) Events(network types.Network, address string) (contract_metadata.Events, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}
	key := fmt.Sprintf("contract_metadata:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return cache.tzip.Events(network, address)
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(contract_metadata.Events), nil
}

// Contract -
func (cache *Cache) Contract(network types.Network, address string) (*contract.Contract, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("contract:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Minute*10, func() (interface{}, error) {
		return cache.contracts.Get(network, address)
	})
	if err != nil {
		return nil, err
	}
	cntr := item.Value().(contract.Contract)
	return &cntr, nil
}

// ProjectIDByHash -
func (cache *Cache) ProjectIDByHash(hash string) string {
	return fmt.Sprintf("project_id:%s", hash)
}

// CurrentBlock -
func (cache *Cache) CurrentBlock(network types.Network) (block.Block, error) {
	key := fmt.Sprintf("block:%d", network)
	item, err := cache.Fetch(key, time.Second*15, func() (interface{}, error) {
		return cache.blocks.Last(network)
	})
	if err != nil {
		return block.Block{}, err
	}
	return item.Value().(block.Block), nil
}

//nolint
// TezosBalance -
func (cache *Cache) TezosBalance(network types.Network, address string, level int64) (int64, error) {
	node, ok := cache.rpc[network]
	if !ok {
		return 0, errors.Errorf("unknown network: %s", network.String())
	}

	key := fmt.Sprintf("tezos_balance:%d:%s:%d", network, address, level)
	item, err := cache.Fetch(key, 30*time.Second, func() (interface{}, error) {

		return node.GetContractBalance(address, level)
	})
	if err != nil {
		return 0, err
	}
	return item.Value().(int64), nil
}

// ScriptBytes -
func (cache *Cache) ScriptBytes(network types.Network, address, symLink string) ([]byte, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("script_bytes:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		script, err := cache.contracts.Script(network, address, symLink)
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

// StorageType -
func (cache *Cache) StorageType(network types.Network, address, symLink string) (*ast.TypedAst, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := fmt.Sprintf("storage:%d:%s", network, address)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		data, err := cache.contracts.ScriptPart(network, address, symLink, consts.STORAGE)
		if err != nil {
			return nil, err
		}
		return ast.NewTypedAstFromBytes(data)
	})
	if err != nil {
		return nil, err
	}
	return item.Value().(*ast.TypedAst), nil
}

// ProtocolByID -
func (cache *Cache) ProtocolByID(network types.Network, id int64) (protocol.Protocol, error) {
	key := fmt.Sprintf("protocol_id:%d:%d", network, id)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return cache.protocols.GetByID(id)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value().(protocol.Protocol), nil
}

// ProtocolByHash -
func (cache *Cache) ProtocolByHash(network types.Network, hash string) (protocol.Protocol, error) {
	key := fmt.Sprintf("protocol_hash:%d:%s", network, hash)
	item, err := cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return cache.protocols.Get(network, hash, -1)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value().(protocol.Protocol), nil
}
