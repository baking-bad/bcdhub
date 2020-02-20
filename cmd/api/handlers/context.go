package handlers

import (
	"github.com/aopoltorzhicky/bcdhub/cmd/api/oauth"
	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	ES    *elastic.Elastic
	RPCs  map[string]*noderpc.NodeRPC
	Dir   string
	DB    database.DB
	OAUTH oauth.Config
}

// NewContext -
func NewContext(e *elastic.Elastic, rpcs map[string]*noderpc.NodeRPC, dir string, db database.DB, oauth oauth.Config) *Context {
	return &Context{
		ES:    e,
		RPCs:  rpcs,
		Dir:   dir,
		DB:    db,
		OAUTH: oauth,
	}
}
