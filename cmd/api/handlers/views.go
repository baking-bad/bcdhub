package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var (
	errNoViews               = errors.New("There aren't views in the metadata")
	errInvalidImplementation = errors.New("Invalid implementation index")
	errEmptyImplementation   = errors.New("Empty implementation")
)

// GetViewsSchema godoc
// @Summary Get view schemas of contract metadata
// @Description Get view schemas of contract metadata
// @Tags contract
// @ID get-contract-tzip-views-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept json
// @Produce json
// @Success 200 {array} ViewSchema
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/views/schema [get]
func (ctx *Context) GetViewsSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	tzip, err := ctx.TZIP.Get(req.Network, req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	schemas := make([]ViewSchema, 0)

	if len(tzip.Views) == 0 {
		c.JSON(http.StatusOK, schemas)
		return
	}

	for _, view := range tzip.Views {
		for i, impl := range view.Implementations {
			if impl.MichelsonStorageView.Empty() {
				continue
			}

			schema := ViewSchema{
				Name:           view.Name,
				Description:    view.Description,
				Implementation: i,
			}

			params := gjson.ParseBytes(impl.MichelsonStorageView.Parameter)
			metadata, err := meta.ParseMetadata(params)
			if ctx.handleError(c, err, 0) {
				return
			}
			entrypoints, err := docstring.GetEntrypoints(metadata)
			if ctx.handleError(c, err, 0) {
				return
			}
			if len(entrypoints) != 1 {
				continue
			}
			schema.Type = entrypoints[0].Type
			schema.Schema, err = jsonschema.Create(entrypoints[0].BinPath, metadata)
			if ctx.handleError(c, err, 0) {
				return
			}

			schemas = append(schemas, schema)
		}
	}

	c.JSON(http.StatusOK, schemas)
}

// ExecuteView godoc
// @Summary Execute view of contracts metadata
// @Description Execute view of contracts metadata
// @Tags contract
// @ID contract-execute-view
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param body body executeViewRequest true "Request body"
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/views/execute [post]
func (ctx *Context) ExecuteView(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var execView executeViewRequest
	if err := c.BindJSON(&execView); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.Blocks.Last(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	tzipValue, err := ctx.TZIP.Get(req.Network, req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	if len(tzipValue.Views) == 0 {
		ctx.handleError(c, errNoViews, 0)
		return
	}

	var impl tzip.ViewImplementation
	for _, view := range tzipValue.Views {
		if view.Name != execView.Name {
			continue
		}
		idx := *execView.Implementation
		if len(view.Implementations) <= idx {
			ctx.handleError(c, errInvalidImplementation, 0)
			return
		}
		impl = view.Implementations[idx]
		break
	}
	if impl.MichelsonStorageView.Empty() {
		ctx.handleError(c, errEmptyImplementation, 0)
		return
	}

	params := gjson.ParseBytes(impl.MichelsonStorageView.Parameter)
	metadata, err := meta.ParseMetadata(params)
	if ctx.handleError(c, err, 0) {
		return
	}
	parameters, err := metadata.BuildEntrypointMicheline("0", execView.Data, false)
	if ctx.handleError(c, err, 0) {
		return
	}

	view := views.NewMichelsonStorageView(impl, execView.Name)
	response, err := views.ExecuteWithoutParsing(rpc, view, views.Context{
		Network:                  req.Network,
		Contract:                 req.Address,
		Source:                   execView.Source,
		Initiator:                execView.Sender,
		ChainID:                  state.ChainID,
		HardGasLimitPerOperation: execView.GasLimit,
		Amount:                   execView.Amount,
		Parameters:               parameters.Get("value").String(),
	})
	if ctx.handleError(c, err, 0) {
		return
	}

	storage := gjson.ParseBytes(impl.MichelsonStorageView.ReturnType)
	storageMetadata, err := meta.ParseMetadata(storage)
	if ctx.handleError(c, err, 0) {
		return
	}
	if response.Get("prim").String() == consts.None {
		c.JSON(http.StatusOK, nil)
		return
	}
	data, err := newmiguel.MichelineToMiguel(response.Get("args.0"), storageMetadata)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, data)
}
