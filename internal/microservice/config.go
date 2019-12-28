package microservice

import (
	"log"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

type config struct {
	Db struct {
		URI string `json:"uri"`
		Log bool   `json:"log"`
	} `json:"db"`
	Mq struct {
		URI   string `json:"uri"`
		Queue string `json:"queue"`
	} `json:"mq"`
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
}

func (cfg config) print() {
	log.Print("-----------CONFIG-----------")
	log.Printf("DB log: %v", cfg.Db.Log)
	for _, node := range cfg.NodeRPC {
		log.Printf("Node: [%s] %s", node.Network, node.Host)
	}

	log.Printf("Message queue: %s", cfg.Mq.Queue)
}

func (cfg config) createRPCs() map[string]*noderpc.NodeRPC {
	rpc := make(map[string]*noderpc.NodeRPC)
	for i := range cfg.NodeRPC {
		nodeCfg := cfg.NodeRPC[i]
		rpc[nodeCfg.Network] = noderpc.NewNodeRPC(nodeCfg.Host, nodeCfg.Network)
		rpc[nodeCfg.Network].SetTimeout(time.Second * 30)
	}
	return rpc
}
