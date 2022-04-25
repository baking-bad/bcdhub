package handlers

import (
	"errors"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/views"
	"github.com/gin-gonic/gin"
)

var (
	errNoViews               = errors.New("there aren't views in the metadata")
	errInvalidImplementation = errors.New("invalid implementation index")
	errEmptyImplementation   = errors.New("empty implementation")
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/views/schema [get]
func GetViewsSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		tzip, err := ctx.ContractMetadata.Get(req.Address)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				c.SecureJSON(http.StatusOK, []ViewSchema{})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		schemas := make([]ViewSchema, 0)

		if len(tzip.Views) == 0 {
			c.SecureJSON(http.StatusOK, schemas)
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

				tree, err := getViewTree(impl)
				if err != nil {
					schema.Error = err.Error()
					schemas = append(schemas, schema)
					continue
				}
				entrypoints, err := tree.GetEntrypointsDocs()
				if err != nil {
					schema.Error = err.Error()
					schemas = append(schemas, schema)
					continue
				}
				if len(entrypoints) != 1 {
					continue
				}
				schema.Type = entrypoints[0].Type
				schema.Schema, err = tree.ToJSONSchema()
				if err != nil {
					schema.Error = err.Error()
					schemas = append(schemas, schema)
					continue
				}

				schemas = append(schemas, schema)
			}
		}

		c.SecureJSON(http.StatusOK, schemas)
	}
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
// @Success 200 {array} ast.MiguelNode
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/views/execute [post]
func ExecuteView() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)
		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var execView executeViewRequest
		if err := c.BindJSON(&execView); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		state, err := ctx.Blocks.Last()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		tzipValue, err := ctx.ContractMetadata.Get(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if len(tzipValue.Views) == 0 {
			handleError(c, ctx.Storage, errNoViews, 0)
			return
		}

		var impl contract_metadata.ViewImplementation
		for _, view := range tzipValue.Views {
			if view.Name != execView.Name {
				continue
			}
			idx := *execView.Implementation
			if len(view.Implementations) <= idx {
				handleError(c, ctx.Storage, errInvalidImplementation, 0)
				return
			}
			impl = view.Implementations[idx]
			break
		}
		if impl.MichelsonStorageView.Empty() {
			handleError(c, ctx.Storage, errEmptyImplementation, 0)
			return
		}

		tree, err := getViewTree(impl)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		if err := tree.FromJSONSchema(execView.Data); handleError(c, ctx.Storage, err, 0) {
			return
		}
		parameters, err := tree.ToParameters("")
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		view := views.NewMichelsonStorageView(impl, execView.Name)
		response, err := views.ExecuteWithoutParsing(c, ctx.RPC, view, views.Args{
			Contract:                 req.Address,
			Source:                   execView.Source,
			Initiator:                execView.Sender,
			ChainID:                  state.Protocol.ChainID,
			HardGasLimitPerOperation: execView.GasLimit,
			Amount:                   execView.Amount,
			Protocol:                 state.Protocol.Hash,
			Parameters:               string(parameters),
		})
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		storage, err := ast.NewTypedAstFromBytes(impl.MichelsonStorageView.ReturnType)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		var responseTree ast.UntypedAST
		if err := json.Unmarshal(response, &responseTree); handleError(c, ctx.Storage, err, 0) {
			return
		}

		if responseTree[0].Prim == consts.None {
			c.SecureJSON(http.StatusOK, nil)
			return
		}

		settleData := []*base.Node{responseTree[0].Args[0]}
		if err := storage.Settle(settleData); handleError(c, ctx.Storage, err, 0) {
			return
		}

		miguel, err := storage.ToMiguel()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, miguel)
	}
}

func getViewTree(impl contract_metadata.ViewImplementation) (*ast.TypedAst, error) {
	if !impl.MichelsonStorageView.IsParameterEmpty() {
		return ast.NewTypedAstFromBytes(impl.MichelsonStorageView.Parameter)
	}
	return ast.NewTypedAstFromString(`{"prim":"unit"}`)
}
