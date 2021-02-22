package handlers

import (
	"github.com/baking-bad/bcdhub/cmd/api/oauth"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/karlseguin/ccache"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Context -
type Context struct {
	*config.Context
	OAUTH oauth.Config
	Cache *ccache.Cache
}

// NewContext -
func NewContext(cfg config.Config) (*Context, error) {
	var oauthCfg oauth.Config
	if cfg.API.OAuthEnabled {
		var err error
		oauthCfg, err = oauth.New(cfg)
		if err != nil {
			return nil, err
		}
	}

	ctx := config.NewContext(
		config.WithStorage(cfg.Storage),
		config.WithRPC(cfg.RPC),
		config.WithDatabase(cfg.DB),
		config.WithShare(cfg.SharePath),
		config.WithTzKTServices(cfg.TzKT),
		config.WithLoadErrorDescriptions(),
		config.WithConfigCopy(cfg),
		config.WithRabbit(cfg.RabbitMQ, cfg.API.ProjectName, cfg.API.MQ),
		config.WithPinata(cfg.API.Pinata),
		config.WithTzipSchema("data/tzip-16-schema.json"),
	)

	return &Context{
		Context: ctx,
		OAUTH:   oauthCfg,
		Cache:   ccache.New(ccache.Configure().MaxSize(10)),
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
