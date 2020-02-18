package main

import (
	"strings"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/tidwall/gjson"
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
	defer ctx.Close()

	if err := ctx.ES.CreateIndexIfNotExists(elastic.DocContracts); err != nil {
		panic(err)
	}

	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

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
