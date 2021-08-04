package handlers

import (
	"github.com/baking-bad/bcdhub/internal/config"
	jsoniter "github.com/json-iterator/go"
	"github.com/microcosm-cc/bluemonday"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Context -
type Context struct {
	*config.Context
	Sanitizer *bluemonday.Policy
}

// NewContext -
func NewContext(cfg config.Config) (*Context, error) {
	ctx := config.NewContext(
		config.WithStorage(cfg.Storage, cfg.API.ProjectName, int64(cfg.API.PageSize)),
		config.WithRPC(cfg.RPC),
		config.WithSearch(cfg.Storage),
		config.WithShare(cfg.SharePath),
		config.WithTzKTServices(cfg.TzKT),
		config.WithLoadErrorDescriptions(),
		config.WithConfigCopy(cfg),
		config.WithPinata(cfg.API.Pinata),
		config.WithTzipSchema("data/tzip-16-schema.json"),
	)

	return &Context{
		Context:   ctx,
		Sanitizer: bluemonday.UGCPolicy(),
	}, nil
}
