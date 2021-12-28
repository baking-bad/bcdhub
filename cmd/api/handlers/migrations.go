package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
)

// GetContractMigrations godoc
// @Summary Get contract migrations
// @Description Get contract migrations
// @Tags contract
// @ID get-contract-migrations
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept json
// @Produce json
// @Success 200 {array} Migration
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/migrations [get]
func (ctx *Context) GetContractMigrations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	contract, err := ctx.Contracts.Get(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	migrations, err := ctx.Migrations.Get(contract.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	result, err := prepareMigrations(ctx, req.NetworkID(), migrations)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, result)
}

func prepareMigrations(ctx *Context, network types.Network, data []migration.Migration) ([]Migration, error) {
	result := make([]Migration, len(data))
	for i := range data {
		proto, err := ctx.Cache.ProtocolByID(network, data[i].ProtocolID)
		if err != nil && !ctx.Storage.IsRecordNotFound(err) {
			return nil, err
		}
		prevProto, err := ctx.Cache.ProtocolByID(network, data[i].PrevProtocolID)
		if err != nil && !ctx.Storage.IsRecordNotFound(err) {
			return nil, err
		}
		result[i] = Migration{
			Level:        data[i].Level,
			Timestamp:    data[i].Timestamp,
			Hash:         data[i].Hash,
			Protocol:     proto.Hash,
			PrevProtocol: prevProto.Hash,
			Kind:         data[i].Kind.String(),
		}
	}
	return result, nil
}
