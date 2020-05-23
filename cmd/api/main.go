package main

import (
	"strings"

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

	docs.SwaggerInfo.Host = cfg.API.Bind

	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	if cfg.API.Sentry.Enabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.API.Sentry.Project)
		defer helpers.CatchPanicSentry()
	}

	ctx, err := handlers.NewContext(cfg)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}
	defer ctx.Close()

	if cfg.API.Seed.Enabled {
		if err := seed.Run(ctx, cfg.Seed); err != nil {
			logger.Fatal(err)
		}
	}

	r := gin.Default()

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

	r.Use(corsSettings())

	if cfg.API.Sentry.Enabled {
		r.Use(helpers.SentryMiddleware())
	}

	v1 := r.Group("v1")
	{
		v1.GET("opg/:hash", ctx.GetOperation)
		v1.GET("operation/:id/error_location", ctx.GetOperationErrorLocation)
		v1.GET("pick_random", ctx.GetRandomContract)
		v1.GET("projects", ctx.GetProjects)
		v1.GET("search", ctx.Search)

		v1.POST("diff", ctx.GetDiff)

		stats := v1.Group("stats")
		{
			stats.GET("", ctx.GetStats)
			networkStats := stats.Group(":network")
			{
				networkStats.GET("", ctx.GetNetworkStats)
				networkStats.GET("series", ctx.GetSeries)
				networkStats.GET("languages", ctx.GetLanguagesStatsForNetwork)
			}
		}

		slug := v1.Group("slug")
		{
			slug.GET(":slug", ctx.GetBySlug)
		}

		contract := v1.Group("contract")
		contract.Use(ctx.IsAuthenticated())
		{
			network := contract.Group(":network")
			{
				address := network.Group(":address")
				{
					address.GET("", ctx.GetContract)
					address.GET("code", ctx.GetContractCode)
					address.GET("operations", ctx.GetContractOperations)
					address.GET("migrations", ctx.GetContractMigrations)
					address.GET("storage", ctx.GetContractStorage)
					address.GET("raw_storage", ctx.GetContractStorageRaw)
					address.GET("rich_storage", ctx.GetContractStorageRich)
					address.GET("rating", ctx.GetContractRating)
					address.GET("mempool", ctx.GetMempool)
					address.GET("same", ctx.GetSameContracts)
					address.GET("similar", ctx.GetSimilarContracts)
					bigmap := address.Group("bigmap")
					{
						bigmap.GET(":ptr", ctx.GetBigMap)
						bigmap.GET(":ptr/:key_hash", ctx.GetBigMapByKeyHash)
					}
					entrypoints := address.Group("entrypoints")
					{
						entrypoints.GET("", ctx.GetEntrypoints)
						entrypoints.POST("data", ctx.GetEntrypointData)
						entrypoints.POST("trace", ctx.RunCode)
					}
				}
			}
		}

		fa12 := v1.Group("tokens")
		{
			network := fa12.Group(":network")
			{
				network.GET("", ctx.GetFA)
				address := network.Group(":address")
				{
					address.GET("transfers", ctx.GetFA12OperationsForAddress)
				}
			}
		}

		oauth := v1.Group("oauth")
		{
			provider := oauth.Group(":provider")
			{
				provider.GET("login", ctx.OauthLogin)
				provider.GET("callback", ctx.OauthCallback)
			}
		}

		authorized := v1.Group("/")
		authorized.Use(ctx.AuthJWTRequired())
		{
			profile := authorized.Group("profile")
			{
				profile.GET("", ctx.GetUserProfile)
				subscriptions := profile.Group("subscriptions")
				{
					subscriptions.GET("", ctx.ListSubscriptions)
					subscriptions.POST("", ctx.CreateSubscription)
					subscriptions.DELETE("", ctx.DeleteSubscription)
					subscriptions.GET("timeline", ctx.GetTimeline)
				}
				vote := profile.Group("vote")
				{
					vote.POST("", ctx.Vote)
					vote.GET("task", ctx.GetNextDiffTask)
				}
			}
		}
	}

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	hub := ws.DefaultHub(cfg.Elastic.URI, cfg.RabbitMQ.URI, cfg.RabbitMQ.Queues)
	hub.Run()
	defer hub.Stop()
	r.GET("ws", func(c *gin.Context) { ws.Handler(c, hub) })

	if err := r.Run(cfg.API.Bind); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

}

func corsSettings() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"*"}
	cfg.AllowCredentials = true
	cfg.AddAllowHeaders("X-Requested-With", "Authorization")
	return cors.New(cfg)
}
