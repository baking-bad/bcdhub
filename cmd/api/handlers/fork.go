package handlers

import (
	stdJSON "encoding/json"
	"io"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/translator"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// ForkContract -
func ForkContract(ctxs config.Contexts) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req forkRequest
		if err := c.BindJSON(&req); handleError(c, ctxs.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		response, err := buildStorageDataFromForkRequest(ctxs, req)
		if err != nil {
			handleError(c, ctxs.Any().Storage, err, 0)
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

func buildStorageDataFromForkRequest(ctxs config.Contexts, req forkRequest) (*ForkResponse, error) {
	var err error
	var scriptData []byte

	if req.Script != "" {
		scriptData = []byte(req.Script)
	} else {
		ctx, err := ctxs.Get(req.NetworkID())
		if err != nil {
			return nil, err
		}
		symLink, err := getCurrentSymLink(ctx.Blocks)
		if err != nil {
			return nil, err
		}
		scriptData, err = getScriptBytes(ctx.Contracts, req.Address, symLink)
		if err != nil {
			return nil, err
		}
	}
	script, err := ast.NewScript(scriptData)
	if err != nil {
		return nil, err
	}

	storageType, err := script.StorageType()
	if err != nil {
		return nil, err
	}

	if storageType.Nodes[0].IsPrim(consts.PAIR) {
		req.Storage = map[string]interface{}{
			storageType.Nodes[0].GetName(): req.Storage,
		}
	}
	if err = storageType.FromJSONSchema(req.Storage); err != nil {
		return nil, err
	}

	storage, err := storageType.ToParameters("")
	if err != nil {
		return nil, err
	}

	return &ForkResponse{
		Script:  scriptData,
		Storage: storage,
	}, nil
}

// CodeFromMichelson -
func CodeFromMichelson() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("contexts").(config.Contexts)

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1024*1024))
		if handleError(c, ctx.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		t, err := translator.NewConverter()
		if handleError(c, ctx.Any().Storage, err, http.StatusBadRequest) {
			return
		}
		micheline, err := t.FromString(string(body))
		if handleError(c, ctx.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		response := CodeFromMichelsonResponse{
			Script:  stdJSON.RawMessage(micheline),
			Storage: CodeFromMichelsonStorage{},
		}

		script, err := ast.NewScript(response.Script)
		if handleError(c, ctx.Any().Storage, err, http.StatusInternalServerError) {
			return
		}

		storageType, err := script.StorageType()
		if handleError(c, ctx.Any().Storage, err, http.StatusInternalServerError) {
			return
		}

		schema, err := storageType.ToJSONSchema()
		if handleError(c, ctx.Any().Storage, err, http.StatusInternalServerError) {
			return
		}
		response.Storage.Schema = schema

		docs, err := storageType.GetEntrypointsDocs()
		if handleError(c, ctx.Any().Storage, err, 0) {
			return
		}
		if len(docs) > 0 {
			response.Storage.Type = docs[0].Type
		}

		response.Storage.DefaultModel = make(ast.JSONModel)
		storageType.GetJSONModel(response.Storage.DefaultModel)

		c.SecureJSON(http.StatusOK, response)
	}
}
