package config

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// CachedAlias -
func (ctx *Context) CachedAlias(network types.Network, address string) string {
	if !bcd.IsContract(address) {
		return ""
	}
	key := ctx.Cache.AliasKey(network, address)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TZIP.Get(network, address)
	})
	if err != nil {
		return ""
	}

	if data, ok := item.Value().(*tzip.TZIP); ok && data != nil {
		return data.Name
	}
	return ""
}

// CachedContractMetadata -
func (ctx *Context) CachedContractMetadata(network types.Network, address string) (*tzip.TZIP, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}
	key := ctx.Cache.ContractMetadataKey(network, address)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TZIP.Get(network, address)
	})
	if err != nil {
		return nil, err
	}

	return item.Value().(*tzip.TZIP), nil
}

// CachedCurrentBlock -
func (ctx *Context) CachedCurrentBlock(network types.Network) (block.Block, error) {
	key := ctx.Cache.BlockKey(network)
	item, err := ctx.Cache.Fetch(key, time.Second*15, func() (interface{}, error) {
		return ctx.Blocks.Last(network)
	})
	if err != nil {
		return block.Block{}, err
	}
	return item.Value().(block.Block), nil
}

// CachedTezosBalance -
func (ctx *Context) CachedTezosBalance(network types.Network, address string, level int64) (int64, error) {
	key := ctx.Cache.TezosBalanceKey(network, address, level)
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

// CachedContract -
func (ctx *Context) CachedContract(network types.Network, address string) (*contract.Contract, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := ctx.Cache.ContractKey(network, address)
	item, err := ctx.Cache.Fetch(key, time.Minute*10, func() (interface{}, error) {
		return ctx.Contracts.Get(network, address)
	})
	if err != nil {
		return nil, err
	}
	cntr := item.Value().(contract.Contract)
	return &cntr, nil
}

// CachedScript -
func (ctx *Context) CachedScript(network types.Network, address, symLink string) (*ast.Script, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := ctx.Cache.ScriptKey(network, address)
	item, err := ctx.Cache.Fetch(key, time.Hour, func() (interface{}, error) {
		script, err := ctx.CachedScriptBytes(network, address, symLink)
		if err != nil {
			return nil, err
		}
		return ast.NewScriptWithoutCode(script)
	})
	if err != nil {
		return nil, err
	}
	return item.Value().(*ast.Script), nil
}

// CachedScriptBytes -
func (ctx *Context) CachedScriptBytes(network types.Network, address, symLink string) ([]byte, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := ctx.Cache.ScriptBytesKey(network, address)
	item, err := ctx.Cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return fetch.ContractBySymLink(network, address, symLink, ctx.SharePath)
	})
	if err != nil {
		return nil, err
	}
	return item.Value().([]byte), nil
}

// CachedStorageType -
func (ctx *Context) CachedStorageType(network types.Network, address, symLink string) (*ast.TypedAst, error) {
	if !bcd.IsContract(address) {
		return nil, nil
	}

	key := ctx.Cache.StorageType(network, address)
	item, err := ctx.Cache.Fetch(key, time.Hour, func() (interface{}, error) {
		data, err := ctx.CachedScriptBytes(network, address, symLink)
		if err != nil {
			return nil, err
		}
		script, err := ast.NewScriptWithoutCode(data)
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

// CachedProtocolByHash -
func (ctx *Context) CachedProtocolByHash(network types.Network, hash string) (protocol.Protocol, error) {
	key := ctx.Cache.ProtocolByIDKey(network, hash)
	item, err := ctx.Cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return ctx.Protocols.Get(network, hash, -1)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value().(protocol.Protocol), nil
}

// CachedProtocolByID -
func (ctx *Context) CachedProtocolByID(network types.Network, id int64) (protocol.Protocol, error) {
	key := ctx.Cache.ProtocolByHashKey(network, id)
	item, err := ctx.Cache.Fetch(key, time.Hour, func() (interface{}, error) {
		return ctx.Protocols.GetByID(id)
	})
	if err != nil {
		return protocol.Protocol{}, err
	}
	return item.Value().(protocol.Protocol), nil
}
