package main

import (
	"log"

	"github.com/fatih/color"
)

type config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	TzKT        map[string]string   `json:"tzkt"`
	TzStats     map[string]string   `json:"tzstats"`
	Indexer     string              `json:"indexer"`
	NodeRPC     map[string][]string `json:"nodes"`
	UpdateTimer int64               `json:"update_timer"`
	Mq          struct {
		URI    string   `json:"uri"`
		Queues []string `json:"queues"`
	} `json:"mq"`
	Sentry struct {
		Project string `json:"project"`
		Env     string `json:"env"`
		DSN     string `json:"dsn"`
		Debug   bool   `json:"debug"`
	} `json:"sentry"`
}

func (cfg config) print() {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	log.Print("-----------CONFIG-----------")
	log.Printf("Indexer: %s", green(cfg.Indexer))
	for network, hosts := range cfg.NodeRPC {
		log.Printf("Nodes %s: %v", green(network), hosts)
	}
	log.Printf("Update every %s second", blue(cfg.UpdateTimer))
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	log.Print("----------------------------")
}
