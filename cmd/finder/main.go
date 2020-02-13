package main

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

const (
	allNetworks = "all"
)

var currentState models.State

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

	s, err := es.CurrentState(allNetworks, models.StateFinder)
	if err != nil {
		panic(err)
	}
	currentState = s

	RPCs := createRPCs(cfg)

	// Initial syncronization
	if err = sync(RPCs, es); err != nil {
		logger.Error(err)
	}

	// Update state by ticker
	ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	for range ticker.C {
		if err = sync(RPCs, es); err != nil {
			logger.Error(err)
		}
	}
}
