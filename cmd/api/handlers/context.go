package handlers

import "github.com/aopoltorzhicky/bcdhub/internal/elastic"

// Context -
type Context struct {
	ES *elastic.Elastic
}

// NewContext -
func NewContext(e *elastic.Elastic) *Context {
	return &Context{
		ES: e,
	}
}
