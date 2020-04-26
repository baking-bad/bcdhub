package migrations

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	ES   *elastic.Elastic
	RPCs map[string]noderpc.Pool
	DB   database.DB

	Config config.Config
}

// NewContext - creates migration context from config
func NewContext(cfg config.Config) (*Context, error) {
	es, err := elastic.New([]string{cfg.Elastic.URI})
	if err != nil {
		return nil, err
	}
	networks := make([]string, 0)
	for k := range cfg.RPC {
		networks = append(networks, k)
	}

	RPCs := make(map[string]noderpc.Pool)
	for network, rpcProvider := range cfg.RPC {
		RPCs[network] = noderpc.NewPool([]string{rpcProvider.URI}, time.Second*time.Duration(rpcProvider.Timeout))
	}

	db, err := database.New(cfg.DB.ConnString)
	if err != nil {
		return nil, err
	}

	return &Context{
		ES:     es,
		RPCs:   RPCs,
		DB:     db,
		Config: cfg,
	}, nil
}

// Close -
func (ctx *Context) Close() {
	ctx.DB.Close()
}

// GetRPC -
func (ctx *Context) GetRPC(network string) (noderpc.Pool, error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return nil, fmt.Errorf("Unknown RPC network: %s", network)
	}
	return rpc, nil
}
