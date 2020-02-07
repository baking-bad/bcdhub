package main

import (
	"fmt"

	"github.com/gin-contrib/cors"

	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers"
	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
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

	oauth, err := createOauth()
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
			oauth.GET("login", ctx.GetOauthLogin)
			oauth.GET("callback", ctx.GetOauthCallback)
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

func createOauth() (*oauth2.Config, error) {
	// TO-DO: uncomment in prod
	// var githubClientID, githubClientSecret string

	// if id := os.Getenv("OAUTH_CLIENT_ID"); id == "" {
	// 	return nil, fmt.Errorf("emtpty OAUTH_CLIENT_ID env variable")
	// } else {
	// 	githubClientID = id
	// }

	// if secret := os.Getenv("OAUTH_CLIENT_SECRET"); secret == "" {
	// 	return nil, fmt.Errorf("emtpty OAUTH_CLIENT_SECRET env variable")
	// } else {
	// 	githubClientSecret = secret
	// }

	// TO-DO: delete in prod
	githubClientID = "d35966939d838f410dd9"
	githubClientSecret = "287ae6a529f479afadd19e4e2386b33f5889f58c"

	// TO-DO: move redirect URL to config
	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:14000/v1/oauth/callback",
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}

	return githubOauthConfig, nil
}
