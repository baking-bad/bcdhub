package main

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	DB  database.DB
	ES  *elastic.Elastic
	RPC map[string]noderpc.Pool
	MQ  *mq.MQ
}

func newContext(cfg config) (*Context, error) {
	es := elastic.WaitNew([]string{cfg.Search.URI})
	RPCs := createRPCs(cfg)
	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		return nil, err
	}

	db, err := database.New(cfg.DB.URI)
	if err != nil {
		return nil, err
	}

	return &Context{
		DB:  db,
		ES:  es,
		RPC: RPCs,
		MQ:  messageQueue,
	}, nil
}

func (ctx *Context) close() {
	ctx.MQ.Close()
}

func createRPCs(cfg config) map[string]noderpc.Pool {
	rpc := make(map[string]noderpc.Pool)
	for network, hosts := range cfg.NodeRPC {
		rpc[network] = noderpc.NewPool(hosts, time.Second*30)
	}
	return rpc
}

// func (ctx *Context) getRPC(network string) (noderpc.Pool, error) {
// 	if rpc, ok := ctx.RPC[network]; ok {
// 		return rpc, nil
// 	}
// 	return nil, fmt.Errorf("Unknown rpc network %s", network)
// }
