package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/jessevdk/go-flags"
)

var ctx *config.Context
var creds awsData

type args struct {
	ConfigFiles func(string) `short:"f" long:"files" default:"config.yml" description:"Config filename.yml"`
}

func main() {
	var options args
	options.ConfigFiles = func(value string) {
		filenames := strings.Split(value, " ")
		cfg, err := config.LoadConfig(filenames...)
		if err != nil {
			logger.Fatal(err)
		}

		creds = awsData{
			BucketName: cfg.Scripts.AWS.BucketName,
			Region:     cfg.Scripts.AWS.Region,
		}

		ctx = config.NewContext(
			config.WithElasticSearch(cfg.Elastic),
			config.WithRabbit(cfg.RabbitMQ, "", cfg.Scripts.MQ),
			config.WithConfigCopy(cfg),
			config.WithRPC(cfg.RPC),
			config.WithShare(cfg.SharePath),
		)
	}

	parser := flags.NewParser(&options, flags.Default)

	if _, err := parser.AddCommand("rollback",
		"Rollback state",
		"Rollback network state to certain level",
		&rollbackCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("remove",
		"Remove network data",
		"Remove full network data from BCD",
		&removeCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("create_repository",
		"Create repository",
		"Create repository",
		&createRepoCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("snapshot",
		"Create snapshot",
		"Create snapshot",
		&snapshotCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("restore",
		"Restore snapshot",
		"Restore snapshot",
		&restoreCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("set_policy",
		"Set policy",
		"Set elastic snapshot policy",
		&setPolicyCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("reload_secure_settings",
		"Reload secure settings",
		"Reload secure settings",
		&reloadSecureSettingsCmd); err != nil {
		logger.Fatal(err)
	}

	if _, err := parser.AddCommand("delete_indices",
		"Delete indices",
		"Delete indices",
		&deleteIndicesCmd); err != nil {
		logger.Fatal(err)
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
	return strings.Replace(text, "\n", "", -1) == "yes"
}

func askQuestion(question string) (string, error) {
	logger.Warning(question)

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Replace(text, "\n", "", -1), nil
}
