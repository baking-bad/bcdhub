package handlers

import (
	"github.com/baking-bad/bcdhub/cmd/api/oauth"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// Context -
type Context struct {
	*config.Context
	OAUTH oauth.Config
}

// NewContext -
func NewContext(cfg config.Config) (*Context, error) {
	var oauthCfg oauth.Config
	if cfg.API.OAuth.Enabled {
		var err error
		oauthCfg, err = oauth.New(cfg)
		if err != nil {
			return nil, err
		}
	}

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithRPC(cfg.RPC),
		config.WithDatabase(cfg.DB),
		config.WithShare(cfg.Share.Path),
		config.WithTzKTServices(cfg.TzKT),
		config.WithLoadErrorDescriptions("data/errors.json"),
		config.WithConfigCopy(cfg),
		config.WithContractsInterfaces(),
	)
	return &Context{
		Context: ctx,
		OAUTH:   oauthCfg,
	}, nil
}

// CurrentUserID - return userID (uint) from gin context
func CurrentUserID(c *gin.Context) uint {
	if val, ok := c.Get("userID"); ok && val != nil {
		if userID, valid := val.(uint); valid {
			return userID
		}
	}

	return 0
}
