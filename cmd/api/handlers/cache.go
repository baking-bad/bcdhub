package handlers

import (
	"fmt"
	"time"
)

func (ctx *Context) getAlias(network, address string) string {
	key := fmt.Sprintf("aliases:%s", network)
	item, err := ctx.Cache.Fetch(key, time.Minute*30, func() (interface{}, error) {
		return ctx.TZIP.GetAliasesMap(network)
	})
	if err != nil {
		return ""
	}
	aliases := item.Value().(map[string]string)
	alias, ok := aliases[address]
	if !ok {
		return ""
	}
	return alias
}
