package main

import (
	"fmt"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/docs"
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/validations"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type app struct {
	Router   *gin.Engine
	Contexts config.Contexts
	Config   config.Config
}

func newApp() *app {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic(err)
	}

	docs.SwaggerInfo.Host = cfg.API.SwaggerHost

	if cfg.API.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.API.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	api := &app{
		Contexts: config.NewContexts(cfg, cfg.API.Networks,
			config.WithStorage(cfg.Storage, cfg.API.ProjectName, int64(cfg.API.PageSize), cfg.API.Connections.Open, cfg.API.Connections.Idle),
			config.WithRPC(cfg.RPC, false),
			config.WithSearch(cfg.Storage),
			config.WithMempool(cfg.Services),
			config.WithLoadErrorDescriptions(),
			config.WithConfigCopy(cfg)),
		Config: cfg,
	}

	api.makeRouter()

	return api
}

func (api *app) makeRouter() {
	r := gin.New()
	store := persistence.NewInMemoryStore(time.Second * 30)

	r.MaxMultipartMemory = 4 << 20 // max upload size 4 MiB
	r.SecureJsonPrefix("")         // do not prepend while(1)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validations.Register(v, api.Config.API); err != nil {
			panic(err)
		}
	}

	if api.Config.API.CorsEnabled {
		r.Use(corsSettings())
	}

	if api.Config.API.SentryEnabled {
		r.Use(helpers.SentryMiddleware())
	}

	r.Use(gin.Recovery())

	if env := os.Getenv(config.EnvironmentVar); env == config.EnvironmentProd {
		r.Use(loggerFormat())
	} else {
		r.Use(gin.Logger())
	}

	v1 := r.Group("v1")
	{
		v1.GET("swagger.json", handlers.MainnetMiddleware(api.Contexts), handlers.GetSwaggerDoc())
		v1.GET("config", handlers.MainnetMiddleware(api.Contexts), handlers.GetConfig())

		v1.GET("head", handlers.ContextsMiddleware(api.Contexts), cache.CachePage(store, time.Second*10, handlers.GetHead()))
		v1.GET("head/:network", handlers.NetworkMiddleware(api.Contexts), cache.CachePage(store, time.Second*10, handlers.GetHeadByNetwork()))
		v1.GET("opg/:hash", handlers.ContextsMiddleware(api.Contexts), handlers.GetOperation())
		v1.GET("pick_random", handlers.ContextsMiddleware(api.Contexts), handlers.GetRandomContract())
		v1.GET("search", handlers.ContextsMiddleware(api.Contexts), handlers.Search())
		v1.POST("fork", handlers.ForkContract(api.Contexts))

		operation := v1.Group("operation/:network/:id")
		operation.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			operation.GET("error_location", handlers.GetOperationErrorLocation())
			operation.GET("diff", handlers.GetOperationDiff())
		}

		stats := v1.Group("stats")
		{
			stats.GET("", handlers.ContextsMiddleware(api.Contexts), cache.CachePage(store, time.Second*30, handlers.GetStats()))

			networkStats := stats.Group(":network")
			networkStats.Use(handlers.NetworkMiddleware(api.Contexts))
			{
				networkStats.GET("", cache.CachePage(store, time.Minute*10, handlers.GetNetworkStats()))
				networkStats.GET("series", cache.CachePage(store, time.Minute*10, handlers.GetSeries()))
				networkStats.GET("contracts", cache.CachePage(store, time.Minute*10, handlers.GetContractsStats()))
				networkStats.GET("recently_called_contracts", cache.CachePage(store, time.Second*10, handlers.RecentlyCalledContracts()))
			}
		}

		slug := v1.Group("slug")
		slug.Use(handlers.MainnetMiddleware(api.Contexts))
		{
			slug.GET(":slug", handlers.GetBySlug())
		}

		bigmap := v1.Group("bigmap/:network/:ptr")
		bigmap.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			bigmap.GET("", cache.CachePage(store, time.Second*30, handlers.GetBigMap()))
			bigmap.GET("count", handlers.GetBigMapDiffCount())
			bigmap.GET("history", handlers.GetBigMapHistory())
			keys := bigmap.Group("keys")
			{
				keys.GET("", handlers.GetBigMapKeys())
				keys.GET(":key_hash", handlers.GetBigMapByKeyHash())
			}
		}

		contract := v1.Group("contract/:network/:address")
		contract.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			contract.GET("", handlers.GetContract())
			contract.GET("code", handlers.GetContractCode())
			contract.GET("operations", handlers.GetContractOperations())
			contract.GET("migrations", handlers.GetContractMigrations())
			contract.GET("transfers", handlers.GetContractTransfers())

			tokens := contract.Group("tokens")
			{
				tokens.GET("", handlers.GetContractTokens())
				tokens.GET("count", handlers.GetContractTokensCount())
				tokens.GET("holders", handlers.GetTokenHolders())
			}

			storage := contract.Group("storage")
			{
				storage.GET("", handlers.GetContractStorage())
				storage.GET("raw", handlers.GetContractStorageRaw())
				storage.GET("rich", handlers.GetContractStorageRich())
				storage.GET("schema", handlers.GetContractStorageSchema())
			}

			contract.GET("mempool", handlers.GetMempool())
			contract.GET("same", handlers.GetSameContracts())
			entrypoints := contract.Group("entrypoints")
			{
				entrypoints.GET("", handlers.GetEntrypoints())
				entrypoints.GET("schema", handlers.GetEntrypointSchema())
				entrypoints.POST("data", handlers.GetEntrypointData())
				entrypoints.POST("trace", handlers.RunCode())
				entrypoints.POST("run_operation", handlers.RunOperation())
			}
			views := contract.Group("views")
			{
				views.GET("schema", handlers.GetViewsSchema())
				views.POST("execute", handlers.ExecuteView())
			}
		}

		account := v1.Group("account/:network")
		account.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			account.GET("", handlers.GetBatchTokenBalances())
			acc := account.Group(":address")
			{
				acc.GET("", handlers.GetInfo())
				acc.GET("metadata", handlers.GetMetadata())
				acc.GET("token_balances", handlers.GetAccountTokenBalances())
				acc.GET("count", handlers.GetAccountTokensCountByContract())
				acc.GET("count_with_metadata", handlers.GetAccountTokensCountByContractWithMetadata())
			}
		}

		fa12 := v1.Group("tokens/:network")
		fa12.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			fa12.GET("", handlers.GetFA())
			fa12.GET("series", handlers.GetTokenVolumeSeries())
			fa12.GET("version/:faversion", handlers.GetFAByVersion())
			fa12.GET("metadata", handlers.GetTokenMetadata())
			transfers := fa12.Group("transfers")
			{
				transfers.GET(":address", handlers.GetFA12OperationsForAddress())
			}
		}

		dapps := v1.Group("dapps")
		dapps.Use(handlers.MainnetMiddleware(api.Contexts))
		{
			dapps.GET("", handlers.GetDAppList())
			dappsBySlug := dapps.Group(":slug")
			{
				dappsBySlug.GET("", handlers.GetDApp())
				dex := dappsBySlug.Group("dex")
				{
					dex.GET("tokens", handlers.GetDexTokens())
					dex.GET("tezos_volume", cache.CachePage(store, time.Minute, handlers.GetDexTezosVolume()))
				}
			}
		}

		globalConstants := v1.Group("global_constants/:network/:address")
		globalConstants.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			globalConstants.GET("", handlers.GetGlobalConstant())
		}
	}
	api.Router = r
}

func (api *app) Close() {
	api.Contexts.Close()
}

func (api *app) Run() {
	if err := api.Router.Run(api.Config.API.Bind); err != nil {
		logger.Err(err)
		helpers.CatchErrorSentry(err)
		return
	}
}

// @title Better Call Dev API
// @description This is API description for Better Call Dev service.

// @contact.name Baking Bad Team
// @contact.url https://baking-bad.org/docs
// @contact.email hello@baking-bad.org

// @x-logo {"url": "https://better-call.dev/img/logo_og.png", "altText": "Better Call Dev logo", "href": "https://better-call.dev"}

// @query.collection.format multi
func main() {
	api := newApp()
	defer api.Close()

	api.Run()
}

func corsSettings() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH"},
		AllowHeaders:     []string{"X-Requested-With", "Authorization", "Origin", "Content-Length", "Content-Type", "Referer", "Cache-Control", "User-Agent"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func loggerFormat() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%15s | %3d | %13v | %-7s %s | %s\n%s",
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
