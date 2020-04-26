package handlers

import (
	"github.com/baking-bad/bcdhub/cmd/api/oauth"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// Context -
type Context struct {
	ES       *elastic.Elastic
	RPCs     map[string]noderpc.Pool
	Dir      string
	DB       database.DB
	OAUTH    oauth.Config
	TzKTSvcs map[string]*tzkt.ServicesTzKT
}

// NewContext -
func NewContext(e *elastic.Elastic, rpcs map[string]noderpc.Pool, svcs map[string]*tzkt.ServicesTzKT, dir string, db database.DB, oauth oauth.Config) (*Context, error) {
	networks := make([]string, 0)
	for k := range rpcs {
		networks = append(networks, k)
	}
	return &Context{
		ES:       e,
		RPCs:     rpcs,
		Dir:      dir,
		DB:       db,
		OAUTH:    oauth,
		TzKTSvcs: svcs,
	}, nil
}
