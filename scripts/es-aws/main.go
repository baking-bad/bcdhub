package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

var handlers = map[string]func(*elastic.Elastic, awsData) error{
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

	esConnection := flag.String("c", "http://localhost:9200", "Elastic search connection string")
	action := flag.String("a", "", fmt.Sprintf("[mandatory] Action: %s", keyString))
	flag.Parse()

	sAction := *action
	if sAction == "" {
		logger.Errorf("'action' key is mandatory")
		return
	}

	es := elastic.WaitNew([]string{*esConnection})

	var creds awsData
	if err := creds.FromEnv(); err != nil {
		logger.Error(err)
		return
	}

	handler, ok := handlers[sAction]
	if !ok {
		logger.Errorf("Invalid action: %s. Allowed: %s", sAction, keyString)
		return
	}

	if err := handler(es, creds); err != nil {
		logger.Errorf("%s: %s", sAction, err)
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
