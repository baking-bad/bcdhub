package main

import (
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/jsonload"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/tidwall/gjson"
)

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		logger.Fatal(err)
	}
	cfg.print()

	helpers.InitSentry(cfg.Sentry.DSN, cfg.Sentry.Debug)
	helpers.SetTagSentry("project", cfg.Sentry.Project)
	defer helpers.CatchPanicSentry()

	ctx, err := newContext(cfg)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}
	defer ctx.close()

	if err := ctx.createIndexes(); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
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
		helpers.CatchErrorSentry(err)
	}

	// Update state by ticker
	ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	for range ticker.C {
		if err = process(ctx); err != nil {
			logger.Error(err)
			helpers.CatchErrorSentry(err)
		}
	}
}
