package indexer

import (
	"log"

	"github.com/fatih/color"
)

// Config -
type Config struct {
	Search struct {
		URI string `json:"uri"`
	} `json:"search"`
	Mq struct {
		URI    string   `json:"uri"`
		Queues []string `json:"queues"`
	} `json:"mq"`
	Sentry struct {
		Project string `json:"project"`
		DSN     string `json:"dsn"`
		Debug   bool   `json:"debug"`
	} `json:"sentry"`
	Indexers map[string]EntityConfig `json:"indexers"`
}

// EntityConfig -
type EntityConfig struct {
	RPC             RPCConfig              `json:"rpc"`
	Boost           bool                   `json:"boost,omitempty"`
	ExternalIndexer *ExternalIndexerConfig `json:"external_indexer,omitempty"`
	UpdateTimer     int64                  `json:"update_timer"`
}

// ExternalIndexerConfig -
type ExternalIndexerConfig struct {
	Type    string `json:"type"`
	Link    string `json:"link"`
	Timeout int64  `json:"timeout"`
}

// RPCConfig -
type RPCConfig struct {
	URLs    []string `json:"urls"`
	Timeout int64    `json:"timeout"`
}

// Print -
func (cfg Config) Print() {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	log.Print("-----------CONFIG-----------")
	log.Printf("Elastic URI: %s", blue(cfg.Search.URI))
	for network, config := range cfg.Indexers {
		log.Printf("[%s]", green(network))
		log.Printf("Nodes: %v", config.RPC.URLs)
		if config.ExternalIndexer != nil {
			log.Printf("External indexer %s: %s", blue(config.ExternalIndexer.Type), config.ExternalIndexer.Link)
		}
		log.Printf("Updates every %s seconds", blue(config.UpdateTimer))
	}
	log.Print("----------------------------")
}
