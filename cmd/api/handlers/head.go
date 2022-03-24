package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetHead godoc
// @Summary Show indexer head
// @Description Get indexer head for each network
// @Tags head
// @ID get-indexer-head
// @Accept json
// @Produce json
// @Success 200 {array} HeadResponse
// @Failure 500 {object} Error
// @Router /v1/head [get]
func GetHead() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)

		ctx, err := ctxs.Get(types.Mainnet)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		stats, err := ctx.Statistics.NetworkStats(types.Empty)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		body := make([]HeadResponse, 0)
		for network, ctx := range ctxs {
			block, err := ctx.Blocks.Last()
			if err != nil {
				if ctx.Storage.IsRecordNotFound(err) {
					continue
				}
				handleError(c, ctx.Storage, err, 0)
				return
			}

			head, err := getHead(ctx, network, block, stats)
			if err != nil {
				logger.Warning().Str("network", network.String()).Err(err).Msg("head API")
				continue
			}

			body = append(body, head)
		}

		c.SecureJSON(http.StatusOK, body)
	}
}

// GetHeadByNetwork godoc
// @Summary Show indexer head for the network
// @Description Get indexer head for the network
// @Tags head
// @ID get-indexer-head-by-network
// @Param network path string true "Network"
// @Accept json
// @Produce json
// @Success 200 {object} HeadResponse
// @Failure 500 {object} Error
// @Router /v1/head/{network} [get]
func GetHeadByNetwork() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		block, err := ctx.Blocks.Last()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		stats, err := ctx.Statistics.NetworkStats(ctx.Network)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		head, err := getHead(ctx, ctx.Network, block, stats)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, head)
	}
}

func getHead(ctx *config.Context, network types.Network, block block.Block, stats map[string]*models.NetworkStats) (HeadResponse, error) {
	var found bool
	for j := range ctx.Config.API.Networks {
		if network.String() == ctx.Config.API.Networks[j] {
			found = true
			break
		}
	}

	if !found {
		return HeadResponse{}, errors.New("unknown network")
	}

	head := HeadResponse{
		Network:   network.String(),
		Level:     block.Level,
		Timestamp: block.Timestamp.UTC(),
		Protocol:  block.Protocol.Hash,
		Synced:    !block.Timestamp.UTC().Add(2 * time.Minute).Before(time.Now().UTC()),
	}
	networkStats, ok := stats[network.String()]
	if !ok {
		return head, errors.Errorf("can't get stats for %s", network.String())
	}
	head.ContractCalls = int64(networkStats.CallsCount)
	head.FACount = int64(networkStats.FACount)
	head.UniqueContracts = int64(networkStats.UniqueContractsCount)
	head.Total = int64(networkStats.ContractsCount)
	return head, nil
}
