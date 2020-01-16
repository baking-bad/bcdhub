package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

	ctx := handlers.NewContext(es)
	r := gin.Default()

	r.Use(corsSettings())
	v1 := r.Group("v1")
	{
		v1.GET("search", ctx.Search)
		contract := v1.Group("contract")
		{
			networkContract := contract.Group(":network")
			{
				address := networkContract.Group(":address")
				{
					address.GET("", ctx.GetContract)
					address.GET(":field", ctx.GetContractField)
				}
			}
		}
		project := v1.Group("project")
		{
			address := project.Group(":address")
			{
				address.GET("", ctx.GetProjectContracts)
			}
		}
	}
	if err := r.Run(cfg.Address); err != nil {
		fmt.Println(err)
	}
}

func corsSettings() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
