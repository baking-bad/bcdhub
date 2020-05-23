package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

// GetStats godoc
// @Summary Show indexer stats
// @Description get indexer states for all networks
// @Tags statistics
// @ID get-stats
// @Accept  json
// @Produce  json
// @Success 200 {array} Block
// @Failure 500 {object} Error
// @Router /stats [get]
func (ctx *Context) GetStats(c *gin.Context) {
	stats, err := ctx.ES.GetAllStates()
	if handleError(c, err, 0) {
		return
	}
	blocks := make([]Block, len(stats))
	for i := range stats {
		blocks[i].FromModel(stats[i])
	}

	c.JSON(http.StatusOK, stats)
}

// GetNetworkStats godoc
// @Summary Network statistics
// @Description Get detailed statistics for network
// @Tags statistics
// @ID get-network-stats
// @Param network path string true "Network"
// @Accept  json
// @Produce  json
// @Success 200 {object} NetworkStats
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /stats/{network} [get]
func (ctx *Context) GetNetworkStats(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var stats NetworkStats
	counts, err := ctx.ES.GetItemsCountForNetwork(req.Network)
	if handleError(c, err, 0) {
		return
	}
	stats.ContractsCount = counts.Contracts
	stats.OperationsCount = counts.Operations

	var protocols []models.Protocol
	if err := ctx.ES.GetByNetworkWithSort(req.Network, "start_level", "desc", &protocols); handleError(c, err, 0) {
		return
	}
	ps := make([]Protocol, len(protocols))
	for i := range protocols {
		ps[i].FromModel(protocols[i])
	}
	stats.Protocols = ps

	c.JSON(http.StatusOK, stats)
}

// GetSeries godoc
// @Summary Get network series
// @Description Get count series data for network
// @Tags statistics
// @ID get-network-series
// @Param network path string true "Network"
// @Param index query string true "One of index name (contract, operation)" Enums(contract, operation)
// @Param period query string true "One of period (year, month, week, day)"  Enums(year, month, week, day)
// @Accept  json
// @Produce  json
// @Success 200 {object} Series
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /stats/{network}/series [get]
func (ctx *Context) GetSeries(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var reqArgs getSeriesRequest
	if err := c.BindQuery(&reqArgs); handleError(c, err, http.StatusBadRequest) {
		return
	}

	series, err := ctx.ES.GetDateHistogram(req.Network, reqArgs.Index, reqArgs.Period)
	if handleError(c, err, 0) {
		return
	}
	var response Series
	response = series

	c.JSON(http.StatusOK, response)
}

// GetLanguagesStatsForNetwork godoc
// @Summary Get languages stats for network
// @Description Get languages stats for network
// @Tags statistics
// @ID get-network-languages
// @Param network path string true "Network"
// @Accept  json
// @Produce  json
// @Success 200 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /stats/{network}/languages [get]
func (ctx *Context) GetLanguagesStatsForNetwork(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	stats, err := ctx.ES.GetLanguagesForNetwork(req.Network)
	if handleError(c, err, 0) {
		return
	}
	delete(stats, "michelson")

	c.JSON(http.StatusOK, stats)
}
