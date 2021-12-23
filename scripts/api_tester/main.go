package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/pkg/errors"
)

var (
	baseURL = "http://127.0.0.1:14000/v1"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
		return
	}

	ctx := config.NewContext(
		config.WithStorage(cfg.Storage, "api_tester", 0, cfg.Scripts.Connections.Open, cfg.Scripts.Connections.Idle),
		config.WithRPC(cfg.RPC),
		config.WithSearch(cfg.Storage),
		config.WithLoadErrorDescriptions(),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	testGeneral(ctx)
	testContracts(ctx)
	testBigMapDiff(ctx)
	testAccounts(ctx)
}

func request(uri string) error {
	url := fmt.Sprintf("%s/%s", baseURL, uri)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Errorf("NewRequest: %v", err)
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("invalid status code: [%d] %s", resp.StatusCode, url)
	}
	return nil
}
