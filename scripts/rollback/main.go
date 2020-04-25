package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/jessevdk/go-flags"
)

func main() {
	var options struct {
		ConfigFiles []string `short:"f" default:"config.yml" description:"Config filename .yml"`
		Level       int64    `short:"l" description:"Level which rollback to"`
		Network     string   `short:"n" description:"Target network"`
	}
	_, err := flags.Parse(&options)
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.LoadConfig(options.ConfigFiles...)
	if err != nil {
		logger.Fatal(err)
	}

	if _, ok := cfg.RPC[options.Network]; !ok {
		logger.Fatal(fmt.Errorf("Unknown network %s", options.Network))
	}

	messageQueue, err := mq.New(cfg.RabbitMQ.URI, []string{"recalc"})
	if err != nil {
		panic(err)
	}

	es := elastic.WaitNew([]string{cfg.Elastic.URI})
	state, err := es.CurrentState(options.Network)
	if err != nil {
		panic(err)
	}

	logger.Warning("App directory: %s", cfg.Share.Path)
	logger.Warning("Elastic search connection string: %s", cfg.Elastic.URI)
	logger.Warning("Rabbit mq connection string: %s", cfg.RabbitMQ.URI)
	logger.Warning("Do you want to rollback '%s' from %d to %d? (yes - continue. no - cancel)", state.Network, state.Level, options.Level)

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	text = strings.Replace(text, "\n", "", -1)

	if text != "yes" {
		logger.Success("Cancelled")
		return
	}

	if err := meta.LoadProtocols(es, []string{options.Network}); err != nil {
		panic(err)
	}

	if err = rollback.Rollback(es, messageQueue, cfg.Share.Path, state, options.Level); err != nil {
		panic(err)
	}
	logger.Success("Done")
}
