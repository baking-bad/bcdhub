package main

import "log"

type config struct {
	Db struct {
		URI string `json:"uri"`
		Log bool   `json:"log"`
	} `json:"db"`
	Mq struct {
		URI    string   `json:"uri"`
		Queues []string `json:"queues"`
	} `json:"mq"`
	TzKT    indexerConfig `json:"tzkt"`
	TzStats indexerConfig `json:"tzstats"`
	Indexer string        `json:"indexer"`
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
	UpdateTimer int64 `json:"update_timer"`
}

type indexerConfig struct {
	MainNet      string `json:"mainnet"`
	ZeroNet      string `json:"zeronet,omitempty"`
	BabylonNet   string `json:"babylonnet,omitempty"`
	CarthagenNet string `json:"carthagenet,omitempty"`
	Timeout      int64  `json:"timeout,omitempty"`
}

func (cfg config) print() {
	log.Print("-----------CONFIG-----------")
	log.Printf("DB log: %v", cfg.Db.Log)
	log.Printf("Indexer: %s", cfg.Indexer)
	for _, node := range cfg.NodeRPC {
		log.Printf("Node: [%s] %s", node.Network, node.Host)
	}
	log.Printf("Update every %d second", cfg.UpdateTimer)
}
