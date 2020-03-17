package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/tidwall/gjson"
	"gopkg.in/go-playground/validator.v9"

	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers"
	"github.com/aopoltorzhicky/bcdhub/cmd/api/oauth"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/cerrors"
	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func main() {
	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		logger.Fatal(err)
	}

	helpers.InitSentry(cfg.Sentry.DSN, cfg.Sentry.Debug)
	helpers.SetTagSentry("project", cfg.Sentry.Project)
	defer helpers.CatchPanicSentry()

	es := elastic.WaitNew([]string{cfg.Search.URI})
	cerrors.LoadErrorDescriptions("data/errors.json")

	db, err := database.New(cfg.DB.URI)
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}
	defer db.Close()

	rpc := createRPC(cfg)

	oauth, err := oauth.New()
	if err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
		return
	}

	ctx := handlers.NewContext(es, rpc, cfg.Dir, db, oauth)

	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("address", handlers.AddressValidator)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("opg", handlers.OPGValidator)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("network", handlers.NetworkValidator)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("status", handlers.StatusValidator)
	}

	r.Use(corsSettings())
	r.Use(helpers.SentryMiddleware())
	v1 := r.Group("v1")
	{
		v1.GET("search", ctx.Search)
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
					address.GET("entrypoints", ctx.GetEntrypoints)
					address.GET("storage", ctx.GetContractStorage)
					address.GET("migration", ctx.GetMigrationDiff)
					address.GET("rating", ctx.GetContractRating)
					address.GET("mempool", ctx.GetMempool)
					address.GET("same", ctx.GetSameContracts)
					address.GET("similar", ctx.GetSimilarContracts)
					bigmap := address.Group("bigmap")
					{
						bigmap.GET(":ptr", ctx.GetBigMap)
						bigmap.GET(":ptr/:key_hash", ctx.GetBigMapByKeyHash)
					}
				}
			}
		}

		v1.GET("pick_random", ctx.GetRandomContract)
		v1.GET("diff", ctx.GetDiff)
		v1.GET("opg/:hash", ctx.GetOperation)
		v1.GET("projects", ctx.GetProjects)

		// PRIVATE
		// TODO - remove in prod
		v1.POST("vote", ctx.Vote)
		v1.POST("set_alias", ctx.SetAlias)

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
					subscriptions.GET("recommended", ctx.Recommendations)
					subscriptions.POST("", ctx.CreateSubscription)
					subscriptions.DELETE("", ctx.DeleteSubscription)
				}
			}
		}
	}
	if err := r.Run(cfg.Address); err != nil {
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

func createRPC(cfg config) map[string]noderpc.Pool {
	rpc := make(map[string]noderpc.Pool)
	for network, hosts := range cfg.NodeRPC {
		rpc[network] = noderpc.NewPool(hosts, time.Second*30)
	}
	return rpc
}
