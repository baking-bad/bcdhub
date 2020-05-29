package handlers

import (
	"fmt"
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

	languages, err := ctx.ES.GetLanguagesForNetwork(req.Network)
	if handleError(c, err, 0) {
		return
	}
	stats.Languages = languages

	c.JSON(http.StatusOK, stats)
}

// GetSeries godoc
// @Summary Get network series
// @Description Get count series data for network
// @Tags statistics
// @ID get-network-series
// @Param network path string true "Network"
// @Param name query string true "One of names" Enums(contract, operation, paid_storage_size_diff, consumed_gas)
// @Param period query string true "One of periods"  Enums(year, month, week, day)
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

	params, err := getSeriesIndexAndField(reqArgs.Name)
	if handleError(c, err, 0) {
		return
	}

	series, err := ctx.ES.GetDateHistogram(req.Network, params.Index, params.Function, params.Field, reqArgs.Period)
	if handleError(c, err, 0) {
		return
	}
	var response Series
	response = series

	c.JSON(http.StatusOK, response)
}

type seriesParams struct {
	Index    string
	Function string
	Field    string
}

func getSeriesIndexAndField(name string) (seriesParams, error) {
	switch name {
	case "contract":
		return seriesParams{
			Index: "contract",
		}, nil
	case "operation":
		return seriesParams{
			Index: "operation",
		}, nil
	case "paid_storage_size_diff":
		return seriesParams{
			Index:    "operation",
			Function: "sum",
			Field:    "result.paid_storage_size_diff",
		}, nil
	case "consumed_gas":
		return seriesParams{
			Index:    "operation",
			Function: "sum",
			Field:    "result.consumed_gas",
		}, nil
	default:
		return seriesParams{}, fmt.Errorf("Unknown series name: %s", name)
	}
}
