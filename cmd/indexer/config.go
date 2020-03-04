package main

import (
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/fatih/color"
)

type config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	TzKT    map[string]string `json:"tzkt"`
	TzStats map[string]string `json:"tzstats"`
	Indexer string            `json:"indexer"`
	NodeRPC []struct {
		Host    string `json:"host"`
		Network string `json:"network"`
	} `json:"nodes"`
	UpdateTimer    int64  `json:"update_timer"`
	FilesDirectory string `json:"files_directory"`
	Mq             struct {
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

	logger.Info("Indexer: %s", blue(cfg.Indexer))
	for _, node := range cfg.NodeRPC {
		log.Printf("Node: [%s] %s", green(node.Network), node.Host)
	}
	logger.Info("Update every %s second", blue(cfg.UpdateTimer))
	logger.Info("Elastic URI: %s", blue(cfg.Search.URI))
	logger.Info("Files directory: %s", blue(cfg.FilesDirectory))
}
