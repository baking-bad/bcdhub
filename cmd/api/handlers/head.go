package handlers

import (
	"net/http"
	"sort"

	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
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

	stats, err := ctx.Storage.GetStats(network)
	if ctx.handleError(c, err, 0) {
		return
	}

	body := make([]HeadResponse, len(blocks))
	for i := range blocks {
		body[i] = HeadResponse{
			Network:   blocks[i].Network.String(),
			Level:     blocks[i].Level,
			Timestamp: blocks[i].Timestamp.UTC(),
			Protocol:  blocks[i].Protocol.Hash,
		}
		networkStats, ok := stats[blocks[i].Network.String()]
		if !ok {
			continue
		}
		body[i].ContractCalls = int64(networkStats.CallsCount)
		body[i].FACount = int64(networkStats.FACount)
		body[i].UniqueContracts = int64(networkStats.UniqueContractsCount)
		body[i].Total = int64(networkStats.ContractsCount)
	}

	c.JSON(http.StatusOK, body)
}
