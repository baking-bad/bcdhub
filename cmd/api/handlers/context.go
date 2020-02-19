package handlers

import (
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	ES   *elastic.Elastic
	RPCs map[string]*noderpc.NodeRPC

	Dir string
}

// NewContext -
func NewContext(e *elastic.Elastic, rpcs map[string]*noderpc.NodeRPC, dir string) *Context {
	return &Context{
		ES:   e,
		RPCs: rpcs,
		Dir:  dir,
	}
}
