package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/views"
	"github.com/gin-gonic/gin"
)

var (
	errNoViews             = errors.New("there aren't views in the metadata")
	errEmptyImplementation = errors.New("empty implementation")
)

// GetViewsSchema godoc
// @Summary Get view schemas of contract metadata
// @Description Get view schemas of contract metadata
// @Tags contract
// @ID get-contract-tzip-views-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param kind query string false "Views kind" Enums(on-chain, off-chain)
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

		var args getViewsArgs
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		views := make([]ViewSchema, 0)

		if args.Kind == EmptyView || args.Kind == OffchainView {
			offChain, err := getOffChainViewsSchema(ctx.ContractMetadata, req.Address)
			if err != nil {
				if !ctx.Storage.IsRecordNotFound(err) {
					handleError(c, ctx.Storage, err, 0)
					return
				}
			}
			views = append(views, offChain...)
		}

		if args.Kind == EmptyView || args.Kind == OnchainView {
			onChain, err := getOnChainViewsSchema(ctx.Contracts, ctx.Blocks, req.Address)
			if err != nil {
				if !ctx.Storage.IsRecordNotFound(err) {
					handleError(c, ctx.Storage, err, 0)
					return
				}
			}
			views = append(views, onChain...)
		}

		c.SecureJSON(http.StatusOK, views)
	}
}

// OffChainView godoc
// @Summary Get JSON schema for off-chain view
// @Description Get JSON schema for off-chain view
// @Tags contract
// @ID get-off-chain-view
// @Param body body json.RawMessage true "Micheline. Limit: 1MB"
// @Accept json
// @Produce json
// @Success 200 {object} ViewSchema
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/off_chain_view [post]
func OffChainView() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var view contract_metadata.View
		if err := json.NewDecoder(io.LimitReader(c.Request.Body, 1024*1024)).Decode(&view); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		c.SecureJSON(http.StatusOK, getOffChainViewSchema(view))
	}
}

func getOffChainViewSchema(view contract_metadata.View) *ViewSchema {
	var schema ViewSchema
	for i, impl := range view.Implementations {
		if impl.MichelsonStorageView.Empty() {
			continue
		}

		schema.Name = view.Name
		schema.Description = view.Description
		schema.Implementation = i
		schema.Kind = OffchainView

		tree, err := getOffChainViewTree(impl)
		if err != nil {
			schema.Error = err.Error()
			return &schema
		}
		entrypoints, err := tree.GetEntrypointsDocs()
		if err != nil {
			schema.Error = err.Error()
			return &schema
		}
		if len(entrypoints) != 1 {
			continue
		}
		schema.Type = entrypoints[0].Type
		schema.Schema, err = tree.ToJSONSchema()
		if err != nil {
			schema.Error = err.Error()
		}
		return &schema
	}

	return nil
}

func getOffChainViewsSchema(contractMetadata contract_metadata.Repository, address string) ([]ViewSchema, error) {
	tzip, err := contractMetadata.Get(address)
	if err != nil {
		return nil, err
	}

	schemas := make([]ViewSchema, 0)

	for _, view := range tzip.Views {
		if schema := getOffChainViewSchema(view); schema != nil {
			schemas = append(schemas, *schema)
		}
	}

	return schemas, nil
}

func getOnChainViewsSchema(contracts contract.Repository, blocks block.Repository, address string) ([]ViewSchema, error) {
	block, err := blocks.Last()
	if err != nil {
		return nil, err
	}
	rawViews, err := contracts.ScriptPart(address, block.Protocol.SymLink, consts.VIEWS)
	if err != nil {
		return nil, err
	}

	if len(rawViews) == 0 {
		return nil, nil
	}

	var views []views.OnChain
	if err := json.Unmarshal(rawViews, &views); err != nil {
		return nil, err
	}

	schemas := make([]ViewSchema, 0)
	for _, view := range views {
		schema := ViewSchema{
			Name: view.ViewName(),
			Kind: OnchainView,
		}

		parameterTree, err := ast.NewTypedAstFromBytes(view.Parameter)
		if err != nil {
			schema.Error = err.Error()
			schemas = append(schemas, schema)
			continue
		}
		entrypoints, err := parameterTree.GetEntrypointsDocs()
		if err != nil {
			schema.Error = err.Error()
			schemas = append(schemas, schema)
			continue
		}
		if len(entrypoints) != 1 {
			continue
		}
		schema.Type = entrypoints[0].Type
		schema.Schema, err = parameterTree.ToJSONSchema()
		if err != nil {
			schema.Error = err.Error()
			schemas = append(schemas, schema)
			continue
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
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

		view, parameters, err := getViewForExecute(ctx.ContractMetadata, ctx.Contracts, ctx.Blocks, req.Address, execView)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		state, err := ctx.Blocks.Last()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		timeoutContext, cancel := context.WithTimeout(c, 10*time.Second)
		defer cancel()

		response, err := view.Execute(timeoutContext, ctx.RPC, views.Args{
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

		storage, err := ast.NewTypedAstFromBytes(view.Return())
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

func getViewForExecute(contractMetadata contract_metadata.Repository, contracts contract.Repository, blocks block.Repository, address string, req executeViewRequest) (views.View, []byte, error) {
	switch req.Kind {
	case OnchainView:
		block, err := blocks.Last()
		if err != nil {
			return nil, nil, err
		}
		rawViews, err := contracts.ScriptPart(address, block.Protocol.SymLink, consts.VIEWS)
		if err != nil {
			return nil, nil, err
		}
		var onChain []views.OnChain
		if err := json.Unmarshal(rawViews, &onChain); err != nil {
			return nil, nil, err
		}

		if len(onChain) == 0 {
			return nil, nil, nil
		}

		for i := range onChain {
			if onChain[i].ViewName() != req.Name {
				continue
			}

			parameterTree, err := ast.NewTypedAstFromBytes(onChain[i].Parameter)
			if err != nil {
				return nil, nil, err
			}
			if err := parameterTree.FromJSONSchema(req.Data); err != nil {
				return nil, nil, err
			}
			parameters, err := parameterTree.ToParameters("")
			if err != nil {
				return nil, nil, err
			}
			return &onChain[i], parameters, nil
		}

		return nil, nil, errNoViews

	case OffchainView:
		if req.View == nil {
			return nil, nil, errors.New("empty off-chain-view")
		}
		if req.View.MichelsonStorageView.Empty() {
			return nil, nil, errEmptyImplementation
		}

		tree, err := getOffChainViewTree(*req.View)
		if err != nil {
			return nil, nil, err
		}
		if err := tree.FromJSONSchema(req.Data); err != nil {
			return nil, nil, err
		}
		parameters, err := tree.ToParameters("")
		if err != nil {
			return nil, nil, err
		}

		return views.NewMichelsonStorageView(*req.View, req.Name), parameters, nil
	default:
		return nil, nil, errors.New("invalid view kind")
	}
}

func getOffChainViewTree(impl contract_metadata.ViewImplementation) (*ast.TypedAst, error) {
	if !impl.MichelsonStorageView.IsParameterEmpty() {
		return ast.NewTypedAstFromBytes(impl.MichelsonStorageView.Parameter)
	}
	return ast.NewTypedAstFromString(`{"prim":"unit"}`)
}
