package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/jessevdk/go-flags"
)

var ctxs config.Contexts
var creds awsData

type awsData struct {
	BucketName string
	Region     string
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
		return
	}

	creds = awsData{
		BucketName: cfg.Scripts.AWS.BucketName,
		Region:     cfg.Scripts.AWS.Region,
	}

	ctxs = config.NewContexts(cfg, cfg.Scripts.Networks,
		config.WithStorage(cfg.Storage, "bcdctl", 0, cfg.Scripts.Connections.Open, cfg.Scripts.Connections.Idle),
		config.WithConfigCopy(cfg),
		config.WithRPC(cfg.RPC, false),
		config.WithSearch(cfg.Storage),
	)
	defer ctxs.Close()

	parser := flags.NewParser(nil, flags.Default)

	if _, err := parser.AddCommand("rollback",
		"Rollback state",
		"Rollback network state to certain level",
		&rollbackCmd); err != nil {
		logger.Err(err)
		return
	}

	if _, err := parser.AddCommand("list_services",
		"Lists metrics services",
		"Lists metrics services",
		&listServicesCmd); err != nil {
		logger.Err(err)
		return
	}

	if _, err := parser.Parse(); err != nil {
		panic(err)
	}

}

func yes() bool {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.ReplaceAll(text, "\n", "") == "yes"
}
