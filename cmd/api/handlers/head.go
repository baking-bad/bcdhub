package handlers

import (
	"net/http"
	"sort"
	"time"

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
func (ctx *Context) GetHead(c *gin.Context) {

	blocks, err := ctx.Blocks.LastByNetworks()
	if ctx.handleError(c, err, 0) {
		return
	}

	var network types.Network
	if len(blocks) == 1 {
		network = blocks[0].Network
	} else {
		sort.Sort(block.ByNetwork(blocks))
	}

	stats, err := ctx.Statistics.NetworkStats(network)
	if ctx.handleError(c, err, 0) {
		return
	}

	body := make([]HeadResponse, 0)
	for i := range blocks {

		head, err := ctx.getHead(blocks[i], stats)
		if err != nil {
			logger.Warning().Str("network", blocks[i].Network.String()).Err(err).Msg("head API")
			continue
		}

		body = append(body, head)
	}

	c.SecureJSON(http.StatusOK, body)
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
func (ctx *Context) GetHeadByNetwork(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	block, err := ctx.Blocks.Last(req.NetworkID())
	if ctx.handleError(c, err, 0) {
		return
	}

	stats, err := ctx.Statistics.NetworkStats(req.NetworkID())
	if ctx.handleError(c, err, 0) {
		return
	}

	head, err := ctx.getHead(block, stats)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, head)
}

func (ctx *Context) getHead(block block.Block, stats map[string]*models.NetworkStats) (HeadResponse, error) {
	var found bool
	for j := range ctx.Config.API.Networks {
		if block.Network.String() == ctx.Config.API.Networks[j] {
			found = true
			break
		}
	}

	if !found {
		return HeadResponse{}, errors.New("unknown network")
	}

	head := HeadResponse{
		Network:   block.Network.String(),
		Level:     block.Level,
		Timestamp: block.Timestamp.UTC(),
		Protocol:  block.Protocol.Hash,
		Synced:    !block.Timestamp.UTC().Add(2 * time.Minute).Before(time.Now().UTC()),
	}
	networkStats, ok := stats[block.Network.String()]
	if !ok {
		return head, errors.Errorf("can't get stats for %s", block.Network)
	}
	head.ContractCalls = int64(networkStats.CallsCount)
	head.FACount = int64(networkStats.FACount)
	head.UniqueContracts = int64(networkStats.UniqueContractsCount)
	head.Total = int64(networkStats.ContractsCount)
	return head, nil
}
