package handlers

import (
	"github.com/baking-bad/bcdhub/cmd/api/oauth"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

type tokenKey struct {
	Network  string
	Contract string
	TokenID  int64
}

// Context -
type Context struct {
	*config.Context
	OAUTH  oauth.Config
	Tokens map[tokenKey]TokenMetadata
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
		config.WithRabbit(cfg.RabbitMQ, cfg.API.ProjectName, cfg.API.Queues),
	)

	tokens, err := ctx.ES.GetTokenMetadata(elastic.GetTokenMetadataContext{
		TokenID: -1,
	})
	if err != nil {
		return nil, err
	}

	mapTokens := make(map[tokenKey]TokenMetadata)
	for i := range tokens {
		mapTokens[tokenKey{
			Network:  tokens[i].Network,
			Contract: tokens[i].Address,
			TokenID:  int64(tokens[i].TokenID),
		}] = TokenMetadata{
			Contract: tokens[i].Address,
			TokenID:  tokens[i].TokenID,
			Symbol:   tokens[i].Symbol,
			Name:     tokens[i].Name,
			Decimals: tokens[i].Decimals,
		}
	}

	return &Context{
		Context: ctx,
		OAUTH:   oauthCfg,
		Tokens:  mapTokens,
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

// FindToken -
func (ctx *Context) FindToken(network, address string, tokenID int64) (TokenMetadata, bool) {
	token, ok := ctx.Tokens[tokenKey{
		Network:  network,
		Contract: address,
		TokenID:  tokenID,
	}]
	return token, ok
}
