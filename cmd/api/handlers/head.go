package handlers

import (
	"net/http"
	"time"

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
	item, err := ctx.Cache.Fetch("head", time.Second*30, ctx.getHead)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, item.Value())
}

func (ctx *Context) getHead() (interface{}, error) {
	blocks, err := ctx.Blocks.LastByNetworks()
	if err != nil {
		return nil, err
	}

	var network string
	if len(blocks) == 1 {
		network = blocks[0].Network
	}
	callCounts, err := ctx.Storage.GetCallsCountByNetwork(network)
	if err != nil {
		return nil, err
	}
	contractStats, err := ctx.Storage.GetContractStatsByNetwork(network)
	if err != nil {
		return nil, err
	}
	faCount, err := ctx.Storage.GetFACountByNetwork(network)
	if err != nil {
		return nil, err
	}
	body := make([]HeadResponse, len(blocks))
	for i := range blocks {
		body[i] = HeadResponse{
			Network:   blocks[i].Network,
			Level:     blocks[i].Level,
			Timestamp: blocks[i].Timestamp,
			Protocol:  blocks[i].Protocol,
		}
		calls, ok := callCounts[blocks[i].Network]
		if ok {
			body[i].ContractCalls = calls
		}
		fa, ok := faCount[blocks[i].Network]
		if ok {
			body[i].FACount = fa
		}
		stats, ok := contractStats[blocks[i].Network]
		if ok {
			body[i].Total = stats.Total
			body[i].TotalBalance = stats.Balance
			body[i].UniqueContracts = stats.SameCount
		}
	}

	return body, nil
}
