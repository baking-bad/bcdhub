package handlers

import "github.com/aopoltorzhicky/bcdhub/internal/elastic"

import "github.com/aopoltorzhicky/bcdhub/internal/noderpc"

// Context -
type Context struct {
	ES   *elastic.Elastic
	RPCs map[string]*noderpc.NodeRPC
}

// NewContext -
func NewContext(e *elastic.Elastic, rpcs map[string]*noderpc.NodeRPC) *Context {
	return &Context{
		ES:   e,
		RPCs: rpcs,
	}
}
