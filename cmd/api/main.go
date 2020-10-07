package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/tidwall/gjson"
	"gopkg.in/go-playground/validator.v9"

	"github.com/baking-bad/bcdhub/cmd/api/docs"
	_ "github.com/baking-bad/bcdhub/cmd/api/docs"
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/seed"
	"github.com/baking-bad/bcdhub/cmd/api/ws"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// @title Better Call Dev API
// @version 1.0
// @description This is API description for Better Call Dev service.

// @BasePath /v1
// @query.collection.format multi

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	docs.SwaggerInfo.Host = cfg.API.SwaggerHost

	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	if cfg.API.Sentry.Enabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.API.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	ctx, err := handlers.NewContext(cfg)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}
	defer ctx.Close()

	if err := ctx.LoadAliases(); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

	if cfg.API.Seed.Enabled {
		if err := seed.Run(ctx, cfg.Seed); err != nil {
			logger.Fatal(err)
		}
	}

	r := gin.Default()
	r.MaxMultipartMemory = 4 << 20 // max upload size 4 MiB

	initValidators(cfg)

	r.Use(corsSettings())

	if cfg.API.Sentry.Enabled {
		r.Use(helpers.SentryMiddleware())
	}

	hub := ws.DefaultHub(cfg.Elastic.URI, cfg.Elastic.Timeout, ctx.MQ)
	hub.Run()
	defer hub.Stop()

	v1 := r.Group("v1")
	{
		v1.GET("docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		v1.GET("ws", func(c *gin.Context) { ws.Handler(c, hub) })

		v1.GET("opg/:hash", ctx.GetOperation)
		v1.GET("operation/:id/error_location", ctx.GetOperationErrorLocation)
		v1.GET("pick_random", ctx.GetRandomContract)
		v1.GET("search", ctx.Search)
		v1.POST("fork", ctx.ForkContract)
		v1.GET("config", ctx.GetConfig)

		v1.POST("diff", ctx.GetDiff)

		stats := v1.Group("stats")
		{
			stats.GET("", ctx.GetStats)
			networkStats := stats.Group(":network")
			{
				networkStats.GET("", ctx.GetNetworkStats)
				networkStats.GET("series", ctx.GetSeries)
				networkStats.GET("contracts", ctx.GetContractsStats)
			}
		}

		slug := v1.Group("slug")
		{
			slug.GET(":slug", ctx.GetBySlug)
		}

		bigmap := v1.Group("bigmap/:network/:ptr")
		{
			bigmap.GET("", ctx.GetBigMap)
			bigmap.GET("history", ctx.GetBigMapHistory)
			keys := bigmap.Group("keys")
			{
				keys.GET("", ctx.GetBigMapKeys)
				keys.GET(":key_hash", ctx.GetBigMapByKeyHash)
			}
		}

		contract := v1.Group("contract/:network/:address")
		contract.Use(ctx.IsAuthenticated())
		{
			contract.GET("", ctx.GetContract)
			contract.GET("code", ctx.GetContractCode)
			contract.GET("operations", ctx.GetContractOperations)
			contract.GET("migrations", ctx.GetContractMigrations)
			contract.GET("transfers", ctx.GetContractTransfers)
			tokens := contract.Group("tokens")
			{
				tokens.GET("", ctx.GetContractTokens)
			}

			storage := contract.Group("storage")
			{
				storage.GET("", ctx.GetContractStorage)
				storage.GET("raw", ctx.GetContractStorageRaw)
				storage.GET("rich", ctx.GetContractStorageRich)
				storage.GET("schema", ctx.GetContractStorageSchema)
			}

			contract.GET("mempool", ctx.GetMempool)
			contract.GET("same", ctx.GetSameContracts)
			contract.GET("similar", ctx.GetSimilarContracts)
			entrypoints := contract.Group("entrypoints")
			{
				entrypoints.GET("", ctx.GetEntrypoints)
				entrypoints.GET("schema", ctx.GetEntrypointSchema)
				entrypoints.POST("data", ctx.GetEntrypointData)
				entrypoints.POST("trace", ctx.RunCode)
				entrypoints.POST("run_operation", ctx.RunOperation)
			}
		}

		account := v1.Group("account/:network/:address")
		{
			account.GET("", ctx.GetInfo)
		}

		fa12 := v1.Group("tokens/:network")
		{
			fa12.GET("", ctx.GetFA)
			fa12.GET("series", ctx.GetTokenVolumeSeries)
			fa12.GET("version/:faversion", ctx.GetFAByVersion)
			transfers := fa12.Group("transfers")
			{
				transfers.GET(":address", ctx.GetFA12OperationsForAddress)
			}
		}

		oauth := v1.Group("oauth/:provider")
		{
			oauth.GET("login", ctx.OauthLogin)
			oauth.GET("callback", ctx.OauthCallback)
		}

		authorized := v1.Group("/")
		authorized.Use(ctx.AuthJWTRequired())
		{
			profile := authorized.Group("profile")
			{
				profile.GET("", ctx.GetUserProfile)
				profile.POST("/mark_all_read", ctx.UserMarkAllRead)
				subscriptions := profile.Group("subscriptions")
				{
					subscriptions.GET("", ctx.ListSubscriptions)
					subscriptions.POST("", ctx.CreateSubscription)
					subscriptions.DELETE("", ctx.DeleteSubscription)
					subscriptions.GET("events", ctx.GetEvents)
					subscriptions.GET("mempool", ctx.GetMempoolEvents)
				}
				vote := profile.Group("vote")
				{
					vote.POST("", ctx.Vote)
					vote.GET("tasks", ctx.GetTasks)
					vote.GET("generate", ctx.GenerateTasks)
				}
				profile.GET("repos", ctx.ListPublicRepos)
				profile.GET("refs", ctx.ListPublicRefs)

				compilations := profile.Group("compilations")
				{
					compilations.GET("", ctx.ListCompilationTasks)

					compilations.GET("verification", ctx.ListVerifications)
					compilations.POST("verification", ctx.CreateVerification)

					compilations.GET("deployment", ctx.ListDeployments)
					compilations.POST("deployment", ctx.CreateDeployment)
					compilations.PATCH("deployment", ctx.FinalizeDeployment)
				}
			}
		}

		dapps := v1.Group("dapps")
		{
			dapps.GET("", ctx.GetDAppList)
			dapps.GET(":slug", ctx.GetDApp)
		}
	}

	if err := r.Run(cfg.API.Bind); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

}

func corsSettings() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH"},
		AllowHeaders:     []string{"X-Requested-With", "Authorization", "Origin", "Content-Length", "Content-Type", "Referer", "Cache-Control"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func initValidators(cfg config.Config) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("address", handlers.AddressValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("opg", handlers.OPGValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("network", handlers.MakeNetworkValidator(cfg.API.Networks)); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("status", handlers.StatusValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("faversion", handlers.FAVersionValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("fill_type", handlers.FillTypeValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("compilation_kind", handlers.CompilationKindValidator); err != nil {
			logger.Fatal(err)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("search", handlers.SearchStringValidator); err != nil {
			logger.Fatal(err)
		}
	}
}
