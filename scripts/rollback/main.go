package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	network := os.Getenv("NETWORK")
	if network == "" {
		fmt.Println("Please, set NETWORK env variable")
		return
	}
	if _, ok := cfg.RPC[network]; !ok {
		logger.Fatal(fmt.Errorf("Unknown network %s", network))
	}

	levelStr := os.Getenv("LEVEL")
	if levelStr == "" {
		fmt.Println("Please, set LEVEL env variable")
		return
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		logger.Fatal(fmt.Errorf("Level has to be in range 0..HEAD, not %s", levelStr))
	}

	messageQueue, err := mq.New(cfg.RabbitMQ.URI, []string{"recalc"})
	if err != nil {
		panic(err)
	}

	es := elastic.WaitNew([]string{cfg.Elastic.URI})
	state, err := es.CurrentState(network)
	if err != nil {
		panic(err)
	}

	logger.Warning("Do you want to rollback '%s' from %d to %d? (yes - continue. no - cancel)", state.Network, state.Level, level)

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

	if err = rollback.Rollback(es, messageQueue, cfg.Share.Path, state, int64(level)); err != nil {
		panic(err)
	}
	logger.Success("Done")
}
