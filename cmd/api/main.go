package main

import (
	"fmt"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/tidwall/gjson"

	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers"
	"github.com/aopoltorzhicky/bcdhub/cmd/api/oauth"
	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/gin-gonic/gin"
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
		panic(err)
	}

	es, err := elastic.New([]string{cfg.Search.URI})
	if err != nil {
		panic(err)
	}

	db, err := database.New(cfg.DB.URI)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rpc := createRPC(cfg.RPCs)

	oauth, err := oauth.New()
	if err != nil {
		panic(err)
	}

	ctx := handlers.NewContext(es, rpc, cfg.Dir, db, oauth)

	r := gin.Default()

	r.Use(cors.Default())
	v1 := r.Group("v1")
	{
		v1.GET("search", ctx.Search)
		contract := v1.Group("contract")
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
				}
			}
		}

		v1.GET("pick_random", ctx.GetRandomContract)
		v1.GET("diff", ctx.GetDiff)
		v1.GET("projects", ctx.GetProjects)
		v1.POST("vote", ctx.Vote)

		project := v1.Group("project")
		{
			address := project.Group(":address")
			{
				address.GET("", ctx.GetProjectContracts)
			}
		}

		oauth := v1.Group("oauth")
		{
			oauth.GET("welcome", ctx.GetOauthWelcome)
			oauth.GET("github/login", ctx.GithubOauthLogin)
			oauth.GET("github/callback", ctx.GithubOauthCallback)
			oauth.GET("gitlab/login", ctx.GitlabOauthLogin)
			oauth.GET("gitlab/callback", ctx.GitlabOauthCallback)
		}

		authorized := v1.Group("/")
		authorized.Use(ctx.AuthJWTRequired())
		{
			authorized.GET("profile", ctx.GetUserProfile)

			subscriptions := authorized.Group("subscriptions")
			{
				subscriptions.GET("", ctx.ListSubscriptions)
				subscriptions.POST("", ctx.CreateSubscription)
				subscriptions.DELETE("", ctx.DeleteSubscription)
			}
		}
	}
	if err := r.Run(cfg.Address); err != nil {
		fmt.Println(err)
	}
}

func corsSettings() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{"*"}
	return cors.New(cfg)
}

func createRPC(data map[string]string) map[string]*noderpc.NodeRPC {
	res := make(map[string]*noderpc.NodeRPC)
	for k, v := range data {
		res[k] = noderpc.NewNodeRPC(v)
	}
	return res
}
