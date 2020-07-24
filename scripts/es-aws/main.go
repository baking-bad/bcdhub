package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/jessevdk/go-flags"
)

var handlers = map[string]func(elastic.IElastic, awsData) error{
	"create_repository":      createRepository,
	"snapshot":               snapshot,
	"restore":                restore,
	"set_policy":             setPolicy,
	"reload_secure_settings": reloadSecureSettings,
	"delete_indices":         deleteIndices,
}

func main() {
	handlerKeys := make([]string, 0)
	for k := range handlers {
		handlerKeys = append(handlerKeys, k)
	}
	keyString := strings.Join(handlerKeys, ",")

	var options struct {
		ConfigFiles []string `short:"f" default:"config.yml" description:"Config filename .yml"`
		Action      string   `short:"a" description:"Action"`
	}
	_, err := flags.Parse(&options)
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.LoadConfig(options.ConfigFiles...)
	if err != nil {
		logger.Fatal(err)
	}

	es := elastic.WaitNew([]string{cfg.Elastic.URI}, cfg.Elastic.Timeout)

	creds := awsData{
		BucketName: cfg.AWS.BucketName,
		Region:     cfg.AWS.Region,
	}

	handler, ok := handlers[options.Action]
	if !ok {
		logger.Errorf("Invalid action: %s. Allowed: %s", options.Action, keyString)
		return
	}

	if err := handler(es, creds); err != nil {
		logger.Errorf("%s: %s", options.Action, err)
		return
	}

	logger.Info("Done")
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
