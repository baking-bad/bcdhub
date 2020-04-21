package main

import (
	"bufio"
	"flag"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

func main() {
	esConnection := flag.String("search", "http://localhost:9200", "Elastic search connection string")
	mqConnection := flag.String("queue", "amqp://guest:guest@localhost:5672/", "Message queue connection string")
	appDirectory := flag.String("dir", "/etc/bcd", "Directory which used to save application data")
	level := flag.Int64("level", 0, "[mandatory] Level which rollback to")
	network := flag.String("network", "", "[mandatory] Network (mainnet, zeronet, carthagenet, babylonnet)")
	flag.Parse()

	if *level == 0 {
		logger.Errorf("Level flag is mandatory")
		flag.PrintDefaults()
		return
	}

	found := false
	for _, net := range []string{"mainnet", "zeronet", "carthagenet", "babylonnet"} {
		if net == *network {
			found = true
			break
		}
	}
	if !found {
		logger.Errorf("Invalid network value flag. One of: mainnet, zeronet, carthagenet, babylonnet ")
		flag.PrintDefaults()
		return
	}

	messageQueue, err := mq.New(*mqConnection, []string{"recalc"})
	if err != nil {
		panic(err)
	}

	es := elastic.WaitNew([]string{*esConnection})
	state, err := es.CurrentState(*network)
	if err != nil {
		panic(err)
	}

	logger.Warning("App directory: %s", *appDirectory)
	logger.Warning("Elastic search connection string: %s", *esConnection)
	logger.Warning("Rabbit mq connection string: %s", *mqConnection)
	logger.Warning("Do you want to rollback '%s' from %d to %d? (yes - continue. no - cancel)", state.Network, state.Level, *level)

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

	if err := meta.LoadProtocols(es, []string{*network}); err != nil {
		panic(err)
	}

	if err = rollback.Rollback(es, messageQueue, *appDirectory, state, *level); err != nil {
		panic(err)
	}
	logger.Success("Done")
}
