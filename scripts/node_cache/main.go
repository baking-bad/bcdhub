package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

func main() {
	var network string
	flag.StringVar(&network, "n", "mainnet", "network name")

	var startLevel int
	flag.IntVar(&startLevel, "l", 1, "start level")

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workers := make(chan struct{}, 50)
	var wg sync.WaitGroup

	dir := path.Join(cfg.SharePath, "node_cache", network)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Err(err)
			return
		}
	}

	rpcConfig, ok := cfg.RPC[network]
	if !ok {
		logger.Error().Msgf("unknwon RPC: %s", network)
		return
	}
	bar := progressbar.NewOptions(1969278, progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for level := int64(1); level < 1969278; level++ {
		if err := bar.Add(1); err != nil {
			logger.Err(err)
			return
		}

		sLevel := strconv.FormatInt(level, 10)
		levelDir := path.Join(dir, sLevel)

		if _, err := os.Stat(levelDir); os.IsNotExist(err) {
			if err := os.MkdirAll(levelDir, os.ModePerm); err != nil {
				logger.Err(err)
				return
			}
		} else {
			continue
		}

		workers <- struct{}{}
		wg.Add(1)
		go cache(ctx, rpcConfig, levelDir, network, sLevel, workers, &wg)
	}
}

func cache(ctx context.Context, rpcConfig config.RPCConfig, dir, network, level string, workers chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		<-workers
	}()

	getHeader(ctx, rpcConfig, dir, network, level)
	getOperations(ctx, rpcConfig, dir, network, level)
}

func getHeader(ctx context.Context, rpcConfig config.RPCConfig, dir, network, level string) {
	urlHeader := rpcConfig.URI + path.Join("/chains/main/blocks", level, "header")
	fileHeader := filepath.Join(dir, "header.json")
	ctxHeader, cancelHeader := context.WithTimeout(ctx, time.Second*time.Duration(rpcConfig.Timeout))
	defer cancelHeader()
	if err := getAndSave(ctxHeader, urlHeader, fileHeader); err != nil {
		logger.Error().Err(err).Str("network", network).Msg("get header")
		return
	}
}

func getOperations(ctx context.Context, rpcConfig config.RPCConfig, dir, network, level string) {
	urlOperations := rpcConfig.URI + path.Join("/chains/main/blocks", level, "operations/3")
	fileOperations := filepath.Join(dir, "operations.json")
	ctxOperations, cancelOperations := context.WithTimeout(ctx, time.Second*time.Duration(rpcConfig.Timeout))
	defer cancelOperations()
	if err := getAndSave(ctxOperations, urlOperations, fileOperations); err != nil {
		logger.Error().Err(err).Str("network", network).Msg("get operations")
		return
	}
}

func getAndSave(ctx context.Context, url, filename string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Errorf("makeGetRequest.NewRequest: %v", err)
	}

	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return errors.Errorf("http.Do: %v", err)
	}
	defer res.Body.Close()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.ReadFrom(res.Body)
	if err != nil {
		return err
	}

	return f.Sync()
}
