package main

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
)

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	ctx, err := newContext(cfg)
	if err != nil {
		panic(err)
	}

	// Initial syncronization
	if err = process(ctx); err != nil {
		logger.Error(err)
	}

	// Update state by ticker
	ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	for range ticker.C {
		if err = process(ctx); err != nil {
			logger.Error(err)
		}
	}
}
