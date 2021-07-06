package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func testBigMapDiff(ctx *config.Context) {
	for ptr := 0; ptr < 100; ptr++ {
		for _, network := range ctx.Config.API.Networks {
			prefix := fmt.Sprintf("bigmap/%s/%d", network, ptr)
			if err := request(prefix); err != nil {
				logger.Err(err)
			}
			if err := request(fmt.Sprintf("%s/count", prefix)); err != nil {
				logger.Err(err)
			}
			if err := request(fmt.Sprintf("%s/history", prefix)); err != nil {
				logger.Err(err)
			}
			if err := request(fmt.Sprintf("%s/keys", prefix)); err != nil {
				logger.Err(err)
			}
		}
	}
}
