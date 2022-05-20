package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// GetConfig -
func GetConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)
		ctx := ctxs.Any()

		tzktEndpoints := make(map[string]string)
		for network, tzkt := range ctx.Config.TzKT {
			tzktEndpoints[network] = tzkt.BaseURI
			break
		}

		cfg := ConfigResponse{
			Networks:       ctx.Config.API.Networks,
			RPCEndpoints:   ctx.Config.API.Frontend.RPC,
			TzKTEndpoints:  tzktEndpoints,
			GaEnabled:      ctx.Config.API.Frontend.GaEnabled,
			MempoolEnabled: ctx.Config.API.Frontend.MempoolEnabled,
			SandboxMode:    ctx.Config.API.Frontend.SandboxMode,
		}

		if ctx.Config.API.SentryEnabled {
			cfg.SentryDSN = ctx.Config.Sentry.FrontURI
		}

		c.SecureJSON(http.StatusOK, cfg)
	}
}
