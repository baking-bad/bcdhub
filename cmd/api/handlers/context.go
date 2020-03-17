package handlers

import (
	"github.com/baking-bad/bcdhub/cmd/api/oauth"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	ES    *elastic.Elastic
	RPCs  map[string]noderpc.Pool
	Dir   string
	DB    database.DB
	OAUTH oauth.Config
}

// NewContext -
func NewContext(e *elastic.Elastic, rpcs map[string]noderpc.Pool, dir string, db database.DB, oauth oauth.Config) *Context {
	return &Context{
		ES:    e,
		RPCs:  rpcs,
		Dir:   dir,
		DB:    db,
		OAUTH: oauth,
	}
}
