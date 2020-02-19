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
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
}

func (cfg config) print() {
	blue := color.New(color.FgBlue).SprintFunc()

	log.Print("-----------CONFIG-----------")
	log.Printf("Update every %s second", blue(cfg.UpdateTimer))
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	log.Print("----------------------------")
}
