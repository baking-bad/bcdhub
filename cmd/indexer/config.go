package main

import "log"

type config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	TzKT    map[string]interface{} `json:"tzkt"`
	TzStats map[string]interface{} `json:"tzstats"`
	Indexer string                 `json:"indexer"`
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
	UpdateTimer int64 `json:"update_timer"`
}

func (cfg config) print() {
	log.Print("-----------CONFIG-----------")
	log.Printf("Indexer: %s", cfg.Indexer)
	for _, node := range cfg.NodeRPC {
		log.Printf("Node: [%s] %s", node.Network, node.Host)
	}
	log.Printf("Update every %d second", cfg.UpdateTimer)
	log.Printf("Elastic URI: %s", cfg.Search.URI)
}
