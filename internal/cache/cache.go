package cache

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/karlseguin/ccache"
)

// Cache -
type Cache struct {
	*ccache.Cache
}

// NewCache -
func NewCache() *Cache {
	return &Cache{
		ccache.New(ccache.Configure().MaxSize(100000)),
	}
}

// AliasKey -
func (cache *Cache) AliasKey(network types.Network, address string) string {
	return fmt.Sprintf("alias:%d:%s", network, address)
}

// ContractMetadataKey -
func (cache *Cache) ContractMetadataKey(network types.Network, address string) string {
	return fmt.Sprintf("contract_metadata:%d:%s", network, address)
}

// ContractKey -
func (cache *Cache) ContractKey(network types.Network, address string) string {
	return fmt.Sprintf("contract:%d:%s", network, address)
}

// TokenMetadataKey -
func (cache *Cache) TokenMetadataKey(network types.Network, address string, tokenID uint64) string {
	return fmt.Sprintf("token_metadata:%d:%s:%d", network, address, tokenID)
}

// BlockKey -
func (cache *Cache) BlockKey(network types.Network) string {
	return fmt.Sprintf("block:%d", network)
}

// TezosBalanceKey -
func (cache *Cache) TezosBalanceKey(network types.Network, address string, level int64) string {
	return fmt.Sprintf("tezos_balance:%d:%s:%d", network, address, level)
}

// ScriptKey -
func (cache *Cache) ScriptKey(network types.Network, address string) string {
	return fmt.Sprintf("script:%d:%s", network, address)
}

// ScriptBytesKey -
func (cache *Cache) ScriptBytesKey(network types.Network, address string) string {
	return fmt.Sprintf("script_bytes:%d:%s", network, address)
}
