package migrations

import (
	"log"

	"github.com/fatih/color"
)

// Config -
type Config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	TzKT    map[string]string   `json:"tzkt"`
	TzStats map[string]string   `json:"tzstats"`
	NodeRPC map[string][]string `json:"nodes"`
	Indexer string              `json:"indexer"`
	Mq      struct {
		URI    string   `json:"uri"`
		Queues []string `json:"queues"`
	} `json:"mq"`
	DB struct {
		URI string `json:"uri"`
	} `json:"db"`
}

// Print -
func (cfg Config) Print() {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	log.Print("-----------CONFIG-----------")
	for network, hosts := range cfg.NodeRPC {
		log.Printf("Nodes %s: %v", green(network), hosts)
	}
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	log.Printf("RabbitMQ URI: %s", blue(cfg.Mq.URI))
	log.Printf("Postgres URI: %s", blue(cfg.DB.URI))
	log.Print("----------------------------")
}
