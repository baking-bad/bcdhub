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
	Router  *gin.Engine
	Context *handlers.Context
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

	ctx, err := handlers.NewContext(cfg)
	if err != nil {
		logger.Err(err)
		helpers.CatchErrorSentry(err)
		return nil
	}

	api := &app{
		Context: ctx,
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
		if err := validations.Register(v, api.Context.Config.API); err != nil {
			panic(err)
		}
	}

	if api.Context.Config.API.CorsEnabled {
		r.Use(corsSettings())
	}

	if api.Context.Config.API.SentryEnabled {
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
		v1.GET("swagger.json", api.Context.GetSwaggerDoc)

		v1.GET("head", cache.CachePage(store, time.Second*10, api.Context.GetHead))
		v1.GET("opg/:hash", api.Context.GetOperation)
		v1.GET("opg/:hash/:content/storage_diff", api.Context.GetContentDiff)
		v1.GET("operation/:id/error_location", api.Context.GetOperationErrorLocation)
		v1.GET("pick_random", api.Context.GetRandomContract)
		v1.GET("search", api.Context.Search)
		v1.POST("fork", api.Context.ForkContract)
		v1.GET("config", api.Context.GetConfig)

		v1.POST("diff", api.Context.GetDiff)

		stats := v1.Group("stats")
		{
			stats.GET("", cache.CachePage(store, time.Second*30, api.Context.GetStats))
			networkStats := stats.Group(":network")
			{
				networkStats.GET("", cache.CachePage(store, time.Minute*10, api.Context.GetNetworkStats))
				networkStats.GET("series", cache.CachePage(store, time.Minute*10, api.Context.GetSeries))
				networkStats.GET("contracts", cache.CachePage(store, time.Minute*10, api.Context.GetContractsStats))
			}
		}

		slug := v1.Group("slug")
		{
			slug.GET(":slug", api.Context.GetBySlug)
		}

		bigmap := v1.Group("bigmap/:network/:ptr")
		{
			bigmap.GET("", cache.CachePage(store, time.Second*30, api.Context.GetBigMap))
			bigmap.GET("count", api.Context.GetBigMapDiffCount)
			bigmap.GET("history", api.Context.GetBigMapHistory)
			keys := bigmap.Group("keys")
			{
				keys.GET("", api.Context.GetBigMapKeys)
				keys.GET(":key_hash", api.Context.GetBigMapByKeyHash)
			}
		}

		contract := v1.Group("contract/:network/:address")
		{
			contract.GET("", api.Context.GetContract)
			contract.GET("code", api.Context.GetContractCode)
			contract.GET("operations", api.Context.GetContractOperations)
			contract.GET("migrations", api.Context.GetContractMigrations)
			contract.GET("transfers", api.Context.GetContractTransfers)

			tokens := contract.Group("tokens")
			{
				tokens.GET("", api.Context.GetContractTokens)
				tokens.GET("count", api.Context.GetContractTokensCount)
				tokens.GET("holders", api.Context.GetTokenHolders)
			}

			storage := contract.Group("storage")
			{
				storage.GET("", api.Context.GetContractStorage)
				storage.GET("raw", api.Context.GetContractStorageRaw)
				storage.GET("rich", api.Context.GetContractStorageRich)
				storage.GET("schema", api.Context.GetContractStorageSchema)
			}

			contract.GET("mempool", api.Context.GetMempool)
			contract.GET("same", api.Context.GetSameContracts)
			contract.GET("similar", api.Context.GetSimilarContracts)
			entrypoints := contract.Group("entrypoints")
			{
				entrypoints.GET("", api.Context.GetEntrypoints)
				entrypoints.GET("schema", api.Context.GetEntrypointSchema)
				entrypoints.POST("data", api.Context.GetEntrypointData)
				entrypoints.POST("trace", api.Context.RunCode)
				entrypoints.POST("run_operation", api.Context.RunOperation)
			}
			views := contract.Group("views")
			{
				views.GET("schema", api.Context.GetViewsSchema)
				views.POST("execute", api.Context.ExecuteView)
			}
		}

		domains := v1.Group("domains/:network")
		{
			domains.GET("", api.Context.TezosDomainsList)
			domains.GET("resolve", api.Context.ResolveDomain)
		}

		account := v1.Group("account/:network")
		{
			account.GET("", api.Context.GetBatchTokenBalances)
			acc := account.Group(":address")
			{
				acc.GET("", api.Context.GetInfo)
				acc.GET("metadata", api.Context.GetMetadata)
				acc.GET("token_balances", api.Context.GetAccountTokenBalances)
				acc.GET("count", api.Context.GetAccountTokensCountByContract)
				acc.GET("count_with_metadata", api.Context.GetAccountTokensCountByContractWithMetadata)
			}
		}

		fa12 := v1.Group("tokens/:network")
		{
			fa12.GET("", api.Context.GetFA)
			fa12.GET("series", api.Context.GetTokenVolumeSeries)
			fa12.GET("version/:faversion", api.Context.GetFAByVersion)
			fa12.GET("metadata", api.Context.GetTokenMetadata)
			transfers := fa12.Group("transfers")
			{
				transfers.GET(":address", api.Context.GetFA12OperationsForAddress)
			}
		}

		metadata := v1.Group("metadata")
		{
			metadata.POST("upload", api.Context.UploadMetadata)
			metadata.GET("list", api.Context.ListMetadata)
			metadata.DELETE("delete", api.Context.DeleteMetadata)
		}

		dapps := v1.Group("dapps")
		{
			dapps.GET("", api.Context.GetDAppList)
			dapps.GET(":slug", api.Context.GetDApp)
		}

		globalConstants := v1.Group("global_constants/:network/:address")
		{
			globalConstants.GET("", api.Context.GetGlobalConstant)
		}
	}
	api.Router = r
}

func (api *app) Close() {
	api.Context.Close()
}

func (api *app) Run() {
	if err := api.Router.Run(api.Context.Config.API.Bind); err != nil {
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
