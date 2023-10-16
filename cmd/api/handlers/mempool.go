package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/services/mempool"
	"github.com/gin-gonic/gin"
)

// GetMempool godoc
// @Summary Get contract mempool operations
// @Description Get contract mempool operations
// @Tags contract
// @ID get-contract-mempool
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/mempool [get]
func GetMempool() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		res, err := ctx.Mempool.Get(c.Request.Context(), req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, mempoolPostprocessing(c.Request.Context(), ctx, res))
	}
}

func mempoolPostprocessing(c context.Context, ctx *config.Context, res mempool.PendingOperations) []Operation {
	ret := make([]Operation, 0)
	if len(res.Originations)+len(res.Transactions) == 0 {
		return ret
	}

	for _, origination := range res.Originations {
		op := prepareMempoolOrigination(ctx, origination)
		if op != nil {
			ret = append(ret, *op)
		}
	}

	for _, tx := range res.Transactions {
		op := prepareMempoolTransaction(c, ctx, tx)
		if op != nil {
			ret = append(ret, *op)
		}
	}

	return ret
}

func prepareMempoolTransaction(c context.Context, ctx *config.Context, tx mempool.PendingTransaction) *Operation {
	status := tx.Status
	if status == consts.Applied {
		status = consts.Pending
	}
	if !helpers.StringInArray(tx.Kind, []string{consts.Transaction, consts.Origination, consts.OriginationNew}) {
		return nil
	}

	amount, err := tx.Amount.Int64()
	if err != nil {
		return nil
	}

	op := Operation{
		Hash:         tx.Hash,
		Network:      ctx.Network.String(),
		Timestamp:    time.Unix(tx.UpdatedAt, 0).UTC(),
		Kind:         tx.Kind,
		Source:       tx.Source,
		Fee:          tx.Fee,
		Counter:      tx.Counter,
		GasLimit:     tx.GasLimit,
		StorageLimit: tx.StorageLimit,
		Amount:       amount,
		Destination:  tx.Destination,
		Mempool:      true,
		Status:       status,
		RawMempool:   tx.Raw,
		Protocol:     tx.Protocol,
	}

	errs, err := tezerrors.ParseArray(tx.Errors)
	if err != nil {
		return nil
	}
	op.Errors = errs

	if bcd.IsContract(op.Destination) && op.Protocol != "" && op.Status == consts.Pending {
		if len(tx.Parameters) > 0 {
			_ = buildMempoolOperationParameters(c, ctx, tx.Parameters, &op)
		} else {
			op.Entrypoint = consts.DefaultEntrypoint
		}
	}

	return &op
}

func prepareMempoolOrigination(ctx *config.Context, origination mempool.PendingOrigination) *Operation {
	status := origination.Status
	if status == consts.Applied {
		status = consts.Pending
	}
	if !helpers.StringInArray(origination.Kind, []string{consts.Transaction, consts.Origination, consts.OriginationNew}) {
		return nil
	}

	op := Operation{
		Hash:         origination.Hash,
		Network:      ctx.Network.String(),
		Timestamp:    time.Unix(origination.UpdatedAt, 0).UTC(),
		Kind:         origination.Kind,
		Source:       origination.Source,
		Fee:          origination.Fee,
		Counter:      origination.Counter,
		GasLimit:     origination.GasLimit,
		StorageLimit: origination.StorageLimit,
		Mempool:      true,
		Status:       status,
		RawMempool:   origination.Raw,
		Protocol:     origination.Protocol,
	}

	errs, err := tezerrors.ParseArray(origination.Errors)
	if err != nil {
		return nil
	}
	op.Errors = errs
	return &op
}

func buildMempoolOperationParameters(c context.Context, ctx *config.Context, data []byte, op *Operation) error {
	proto, err := ctx.Protocols.Get(c, op.Protocol, -1)
	if err != nil {
		return err
	}
	parameter, err := getParameterType(c, ctx.Contracts, op.Destination, proto.SymLink)
	if err != nil {
		return err
	}
	params := types.NewParameters(data)
	op.Entrypoint = params.Entrypoint

	tree, err := parameter.FromParameters(params)
	if err != nil {
		return err
	}

	op.Parameters, err = tree.ToMiguel()
	if err != nil && !tezerrors.HasParametersError(op.Errors) {
		return err
	}
	return nil
}
