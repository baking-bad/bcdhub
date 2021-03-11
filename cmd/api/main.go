package main

import (
	"fmt"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/docs"
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/seed"
	"github.com/baking-bad/bcdhub/cmd/api/validations"
	"github.com/baking-bad/bcdhub/cmd/api/ws"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
)

type app struct {
	Router  *gin.Engine
	Hub     *ws.Hub
	Context *handlers.Context
}

func newApp() *app {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	docs.SwaggerInfo.Host = cfg.API.SwaggerHost

	if cfg.API.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.API.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctx, err := handlers.NewContext(cfg)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return nil
	}

	if cfg.API.SeedEnabled {
		if err := seed.Run(ctx, cfg.API.Seed); err != nil {
			logger.Fatal(err)
		}
	}

	api := &app{
		Hub:     ws.DefaultHub(ctx),
		Context: ctx,
	}

	api.makeRouter()

	return api
}

func (api *app) makeRouter() {
	r := gin.New()

	r.MaxMultipartMemory = 4 << 20 // max upload size 4 MiB

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validations.Register(v, api.Context.Config.API.Networks); err != nil {
			logger.Fatal(err)
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
		v1.GET("ws", func(c *gin.Context) { ws.Handler(c, api.Hub) })

		v1.GET("opg/:hash", api.Context.GetOperation)
		v1.GET("operation/:id/error_location", api.Context.GetOperationErrorLocation)
		v1.GET("pick_random", api.Context.GetRandomContract)
		v1.GET("search", api.Context.Search)
		v1.POST("fork", api.Context.ForkContract)
		v1.GET("config", api.Context.GetConfig)

		v1.POST("diff", api.Context.GetDiff)

		stats := v1.Group("stats")
		{
			stats.GET("", api.Context.GetStats)
			networkStats := stats.Group(":network")
			{
				networkStats.GET("", api.Context.GetNetworkStats)
				networkStats.GET("series", api.Context.GetSeries)
				networkStats.GET("contracts", api.Context.GetContractsStats)
			}
		}

		slug := v1.Group("slug")
		{
			slug.GET(":slug", api.Context.GetBySlug)
		}

		bigmap := v1.Group("bigmap/:network/:ptr")
		{
			bigmap.GET("", api.Context.GetBigMap)
			bigmap.GET("count", api.Context.GetBigMapDiffCount)
			bigmap.GET("history", api.Context.GetBigMapHistory)
			keys := bigmap.Group("keys")
			{
				keys.GET("", api.Context.GetBigMapKeys)
				keys.GET(":key_hash", api.Context.GetBigMapByKeyHash)
			}
		}

		contract := v1.Group("contract/:network/:address")
		contract.Use(api.Context.IsAuthenticated())
		{
			contract.GET("", api.Context.GetContract)
			contract.GET("code", api.Context.GetContractCode)
			contract.GET("operations", api.Context.GetContractOperations)
			contract.GET("migrations", api.Context.GetContractMigrations)
			contract.GET("transfers", api.Context.GetContractTransfers)

			tokens := contract.Group("tokens")
			{
				tokens.GET("", api.Context.GetContractTokens)
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

		account := v1.Group("account/:network/:address")
		{
			account.GET("", api.Context.GetInfo)
			account.GET("metadata", api.Context.GetMetadata)
		}

		fa12 := v1.Group("tokens/:network")
		{
			fa12.GET("", api.Context.GetFA)
			fa12.GET("series", api.Context.GetTokenVolumeSeries)
			fa12.GET("version/:faversion", api.Context.GetFAByVersion)
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

		oauth := v1.Group("oauth/:provider")
		{
			oauth.GET("login", api.Context.OauthLogin)
			oauth.GET("callback", api.Context.OauthCallback)
		}

		authorized := v1.Group("/")
		authorized.Use(api.Context.AuthJWTRequired())
		{
			profile := authorized.Group("profile")
			{
				profile.GET("", api.Context.GetUserProfile)
				profile.POST("/mark_all_read", api.Context.UserMarkAllRead)
				subscriptions := profile.Group("subscriptions")
				{
					subscriptions.GET("", api.Context.ListSubscriptions)
					subscriptions.POST("", api.Context.CreateSubscription)
					subscriptions.DELETE("", api.Context.DeleteSubscription)
					subscriptions.GET("events", api.Context.GetEvents)
					subscriptions.GET("mempool", api.Context.GetMempoolEvents)
				}
				vote := profile.Group("vote")
				{
					vote.POST("", api.Context.Vote)
					vote.GET("tasks", api.Context.GetTasks)
					vote.GET("generate", api.Context.GenerateTasks)
				}
				profile.GET("accounts", api.Context.ListPublicAccounts)
				profile.GET("repos", api.Context.ListPublicRepos)
				profile.GET("refs", api.Context.ListPublicRefs)

				compilations := profile.Group("compilations")
				{
					compilations.GET("", api.Context.ListCompilationTasks)

					compilations.GET("verification", api.Context.ListVerifications)
					compilations.POST("verification", api.Context.CreateVerification)

					compilations.GET("deployment", api.Context.ListDeployments)
					compilations.POST("deployment", api.Context.CreateDeployment)
					compilations.PATCH("deployment", api.Context.FinalizeDeployment)
				}
			}
		}

		dapps := v1.Group("dapps")
		{
			dapps.GET("", api.Context.GetDAppList)
			dapps.GET(":slug", api.Context.GetDApp)
		}
	}
	api.Router = r
}

func (api *app) Close() {
	api.Context.Close()
	api.Hub.Stop()
}

func (api *app) Run() {
	api.Hub.Run()
	if err := api.Router.Run(api.Context.Config.API.Bind); err != nil {
		logger.Error(err)
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
