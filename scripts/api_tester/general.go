package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func testGeneral(ctx *config.Context) {
	if err := request("head"); err != nil {
		logger.Err(err)
	}
	if err := request("config"); err != nil {
		logger.Err(err)
	}
	if err := request("pick_random"); err != nil {
		logger.Err(err)
	}
	if err := request("stats"); err != nil {
		logger.Err(err)
	}

	for _, network := range ctx.Config.API.Networks {
		if err := request(fmt.Sprintf("stats/%s", network)); err != nil {
			logger.Err(err)
		}
		if err := request(fmt.Sprintf("tokens/%s", network)); err != nil {
			logger.Err(err)
		}
	}
}
