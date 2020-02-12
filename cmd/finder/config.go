package main

import (
	"log"

	"github.com/fatih/color"
)

type config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
	UpdateTimer int64 `json:"update_timer"`
}

func (cfg config) print() {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	log.Print("-----------CONFIG-----------")
	for _, node := range cfg.NodeRPC {
		log.Printf("Node: [%s] %s", green(node.Network), node.Host)
	}
	log.Printf("Update every %s second", blue(cfg.UpdateTimer))
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	log.Print("----------------------------")
}
