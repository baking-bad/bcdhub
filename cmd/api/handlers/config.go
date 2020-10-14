package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetConfig -
func (ctx *Context) GetConfig(c *gin.Context) {
	rpcEndpoints := make(map[string]string)
	tzktEndpoints := make(map[string]string)

	for network, rpc := range ctx.Config.RPC {
		rpcEndpoints[network] = rpc.URI
	}

	for network, tzkt := range ctx.Config.TzKT {
		tzktEndpoints[network] = tzkt.BaseURI
	}

	cfg := ConfigResponse{
		Networks:      ctx.Config.API.Networks,
		OauthEnabled:  ctx.Config.API.OAuthEnabled,
		RPCEndpoints:  rpcEndpoints,
		TzKTEndpoints: tzktEndpoints,
	}

	c.JSON(http.StatusOK, cfg)
}
