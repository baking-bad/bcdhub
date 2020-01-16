package main

import (
	"log"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

var states = map[string]*models.State{}
var statesContract = map[string]*models.State{}

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	es, err := elastic.New([]string{cfg.Search.URI})
	if err != nil {
		panic(err)
	}

	RPCs := createRPCs(cfg)
	indexers, err := createIndexers(es, cfg)
	if err != nil {
		panic(err)
	}

	// Initial syncronization
	if err = sync(RPCs, indexers, es); err != nil {
		panic(err)
	}

	// Update state by ticker
	ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	for range ticker.C {
		if err = sync(RPCs, indexers, es); err != nil {
			log.Println(err)
		}
	}

	// res, err := getOperations(RPCs["mainnet"], es, 665514, "mainnet", map[string]struct{}{"KT1NQfJvo9v8hXmEgqos8NP7sS8V4qaEfvRF": struct{}{}})
	// if err != nil {
	// 	panic(err)
	// }

	// b, err := json.MarshalIndent(res, "", " ")
	// if err != nil {
	// 	panic(err)
	// }
	// log.Print(string(b))
}
