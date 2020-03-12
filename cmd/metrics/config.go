package main

import (
	"log"

	"github.com/fatih/color"
)

type config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	UpdateTimer int64 `json:"update_timer"`
	Mq          struct {
		URI    string   `json:"uri"`
		Queues []string `json:"queues"`
	} `json:"mq"`
	DB struct {
		URI string `json:"uri"`
	}
	NodeRPC map[string][]string `json:"nodes"`
	Sentry  struct {
		Project string `json:"project"`
		DSN     string `json:"dsn"`
		Debug   bool   `json:"debug"`
	} `json:"sentry"`
}

func (cfg config) print() {
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	log.Print("-----------CONFIG-----------")
	for network, hosts := range cfg.NodeRPC {
		log.Printf("Nodes %s: %v", green(network), hosts)
	}
	log.Printf("Update every %s second", blue(cfg.UpdateTimer))
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	log.Printf("DB URI: %s", blue(cfg.DB.URI))
	log.Print("----------------------------")
}
