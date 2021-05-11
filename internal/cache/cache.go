package cache

import (
	"fmt"

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
func (cache *Cache) AliasKey(network, address string) string {
	return fmt.Sprintf("alias:%s:%s", network, address)
}

// ContractKey -
func (cache *Cache) ContractKey(network, address string) string {
	return fmt.Sprintf("contract:%s:%s", network, address)
}

// TokenMetadataKey -
func (cache *Cache) TokenMetadataKey(network, address string, tokenID uint64) string {
	return fmt.Sprintf("token_metadata:%s:%s:%d", network, address, tokenID)
}

// BlockKey -
func (cache *Cache) BlockKey(network string) string {
	return fmt.Sprintf("block:%s", network)
}

// TezosBalanceKey -
func (cache *Cache) TezosBalanceKey(network, address string, level int64) string {
	return fmt.Sprintf("tezos_balance:%s:%s:%d", network, address, level)
}
