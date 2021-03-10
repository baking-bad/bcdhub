package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

var (
	baseURL = "http://localhost:14000/v1"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := config.NewContext(
		config.WithStorage(cfg.Storage),
		config.WithRPC(cfg.RPC),
		config.WithShare(cfg.SharePath),
		config.WithTzKTServices(cfg.TzKT),
		config.WithLoadErrorDescriptions(),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	baseURL = fmt.Sprintf("http://%s/v1", ctx.Config.API.Bind)

	testContracts(ctx)
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
		return fmt.Errorf("Invalid status code: [%d] %s", resp.StatusCode, url)
	}
	return nil
}

func testContracts(ctx *config.Context) {
	threads := 6
	offset := int64(41138)

	for _, network := range ctx.Config.API.Networks {
		if network != "edo2net" {
			continue
		}
		logger.Info("testing %s contract endpoints...", network)

		contracts, err := ctx.Contracts.GetMany(map[string]interface{}{
			"network": network,
		})
		if err != nil {
			logger.Errorf("testContracts: %s", err.Error())
			return
		}

		total := len(contracts)
		contracts = contracts[offset:]

		tasks := make(chan contract.Contract, len(contracts))
		for i := range contracts {
			tasks <- contracts[i]
		}

		counter := offset
		var wg sync.WaitGroup
		stop := make(chan struct{}, threads)

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go testContract(tasks, stop, &counter, &wg)
		}

		wg.Add(1)
		go stopThread(threads, total, &counter, stop, &wg)

		wg.Wait()
	}
}

func stopThread(threads, total int, counter *int64, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	if counter != nil {
		bar := progressbar.NewOptions(total, progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())

		for int64(total) > *counter {
			ctr := int(*counter)
			if err := bar.Set(ctr); err != nil {
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	for i := 0; i < threads; i++ {
		stop <- struct{}{}
	}
}

func testContract(tasks chan contract.Contract, stop chan struct{}, counter *int64, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stop:
			return
		case contract := <-tasks:
			prefix := fmt.Sprintf("contract/%s/%s", contract.Network, contract.Address)

			if err := request(prefix); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/code", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/operations", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/migrations", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/transfers", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/tokens", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/storage", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/same", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/similar", prefix)); err != nil {
				logger.Error(err)
			}
			if err := request(fmt.Sprintf("%s/entrypoints", prefix)); err != nil {
				logger.Error(err)
			}
			for i := range contract.Entrypoints {
				if err := request(fmt.Sprintf("%s/entrypoints/schema?entrypoint=%s", prefix, contract.Entrypoints[i])); err != nil {
					logger.Error(err)
				}
			}
			atomic.AddInt64(counter, 1)
		}

	}
}
