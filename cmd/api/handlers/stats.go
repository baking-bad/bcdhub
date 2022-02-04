package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
// @Router /v1/stats [get]
func (ctx *Context) GetStats(c *gin.Context) {
	stats := make([]block.Block, 0)
	for _, net := range ctx.Config.API.Networks {
		block, err := ctx.Blocks.Last(types.NewNetwork(net))
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				continue
			}
			ctx.handleError(c, err, 0)
			return
		}
		stats = append(stats, block)
	}

	blocks := make([]Block, 0)
	for i := range stats {
		var block Block
		block.FromModel(stats[i])
		blocks = append(blocks, block)
	}

	c.SecureJSON(http.StatusOK, blocks)
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
// @Router /v1/stats/{network} [get]
func (ctx *Context) GetNetworkStats(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var stats NetworkStats
	counts, err := ctx.Statistics.NetworkCountStats(req.NetworkID())
	if ctx.handleError(c, err, 0) {
		return
	}
	stats.ContractsCount = counts[models.DocContracts]
	stats.OperationsCount = counts[models.DocOperations]

	protocols, err := ctx.Protocols.GetByNetworkWithSort(req.NetworkID(), "start_level", "desc")
	if ctx.handleError(c, err, 0) {
		return
	}
	ps := make([]Protocol, len(protocols))
	for i := range protocols {
		ps[i].FromModel(protocols[i])
	}
	stats.Protocols = ps

	head, err := ctx.Statistics.NetworkStats(req.NetworkID())
	if ctx.handleError(c, err, 0) {
		return
	}

	if networkHead, ok := head[req.Network]; ok {
		stats.ContractCalls = networkHead.CallsCount
		stats.UniqueContracts = networkHead.UniqueContractsCount
		stats.FACount = networkHead.FACount
	}

	c.SecureJSON(http.StatusOK, stats)
}

// GetSeries godoc
// @Summary Get network series
// @Description Get count series data for network
// @Tags statistics
// @ID get-network-series
// @Param network path string true "Network"
// @Param name query string true "One of names" Enums(contract, operation, paid_storage_size_diff, consumed_gas, users, volume)
// @Param period query string true "One of periods"  Enums(year, month, week, day, hour)
// @Param address query string false "Comma-separated contract addresses"
// @Accept  json
// @Produce  json
// @Success 200 {object} Series
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/stats/{network}/series [get]
func (ctx *Context) GetSeries(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var reqArgs getSeriesRequest
	if err := c.BindQuery(&reqArgs); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var addresses []string
	if reqArgs.Address != "" {
		addresses = strings.Split(reqArgs.Address, ",")
	}

	options, err := ctx.getHistogramOptions(reqArgs.Name, req.NetworkID(), addresses...)
	if ctx.handleError(c, err, 0) {
		return
	}

	series, err := ctx.Statistics.Histogram(reqArgs.Period, options...)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, series)
}

func (ctx *Context) getHistogramOptions(name string, network types.Network, addresses ...string) ([]models.HistogramOption, error) {
	filters := []models.HistogramFilter{
		{
			Field: "network",
			Value: network,
			Kind:  models.HistogramFilterKindMatch,
		},
	}

	switch name {
	case "contract":
		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocContracts),
			models.WithHistogramFilters(filters),
		}, nil
	case "operation":
		filters = append(filters, models.HistogramFilter{
			Field: "entrypoint",
			Value: "",
			Kind:  models.HistogramFilterKindExists,
		})

		filters = append(filters, models.HistogramFilter{
			Field: "status",
			Value: types.OperationStatusApplied.String(),
			Kind:  models.HistogramFilterKindMatch,
		})

		if len(addresses) > 0 {
			filters = append(filters, models.HistogramFilter{
				Kind:  models.HistogramFilterKindAddresses,
				Value: addresses,
				Field: "destination",
			})
		}

		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocOperations),
			models.WithHistogramFilters(filters),
		}, nil
	case "paid_storage_size_diff":
		if len(addresses) > 0 {
			filters = append(filters, models.HistogramFilter{
				Kind:  models.HistogramFilterKindAddresses,
				Value: addresses,
				Field: "destination",
			})
		}

		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocOperations),
			models.WithHistogramFunction("sum", "paid_storage_diff"),
			models.WithHistogramFilters(filters),
		}, nil
	case "consumed_gas":
		if len(addresses) > 0 {
			filters = append(filters, models.HistogramFilter{
				Kind:  models.HistogramFilterKindAddresses,
				Value: addresses,
				Field: "destination",
			})
		}

		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocOperations),
			models.WithHistogramFunction("sum", "consumed_gas"),
			models.WithHistogramFilters(filters),
		}, nil
	case "users":
		if len(addresses) > 0 {
			filters = append(filters, models.HistogramFilter{
				Kind:  models.HistogramFilterKindAddresses,
				Value: addresses,
				Field: "destination",
			})
		}

		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocOperations),
			models.WithHistogramFunction("cardinality", "initiator.keyword"),
			models.WithHistogramFilters(filters),
		}, nil
	case "volume":
		if len(addresses) > 0 {
			filters = append(filters, models.HistogramFilter{
				Kind:  models.HistogramFilterKindAddresses,
				Value: addresses,
				Field: "destination",
			})
		}

		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocOperations),
			models.WithHistogramFunction("sum", "amount"),
			models.WithHistogramFilters(filters),
		}, nil
	case "token_volume":
		return []models.HistogramOption{
			models.WithHistogramIndex(models.DocTransfers),
			models.WithHistogramFunction("sum", "amount"),
			models.WithHistogramFilters(filters),
		}, nil
	default:
		return nil, errors.Errorf("Unknown series name: %s", name)
	}
}

// GetContractsStats godoc
// @Summary Show contracts stats
// @Description Show total volume, unique users and transactions count for period
// @Tags contract
// @ID get-stats-contracts
// @Param network path string true "Network"
// @Param contracts query string true "Comma-separated KT addresses" minlength(36)
// @Param period query string true "One of periods"  Enums(all, year, month, week, day)
// @Accept  json
// @Produce  json
// @Success 200 {object} operation.DAppStats
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/stats/{network}/contracts [get]
func (ctx *Context) GetContractsStats(c *gin.Context) {
	var req getByNetwork
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqStats GetTokenStatsRequest
	if err := c.BindQuery(&reqStats); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	addresses := reqStats.Addresses()
	if len(addresses) == 0 {
		ctx.handleError(c, errors.Errorf("Empty address list"), http.StatusBadRequest)
		return
	}
	stats, err := ctx.Operations.GetDAppStats(req.NetworkID(), addresses, reqStats.Period)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, stats)
}
