package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/aopoltorzhicky/bcdhub/cmd/api/handlers"
	"github.com/aopoltorzhicky/bcdhub/internal/db"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
)

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}

	db, err := db.Database(cfg.Db.URI, cfg.Db.Log)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := handlers.NewContext(db)
	r := gin.Default()
	v1 := r.Group("v1")
	{
		v1.GET("search", ctx.Search)
		projects := v1.Group("projects")
		{
			projects.GET(":id", ctx.GetProject)
		}
		contracts := v1.Group("contracts")
		{

			contract := contracts.Group(":id")
			{
				contract.GET("", ctx.GetContract)
				contract.GET(":field", ctx.GetContractField)
			}
		}
		contract := v1.Group("contract")
		{
			networkContract := contract.Group(":network")
			{
				address := networkContract.Group(":address")
				{
					address.GET("", ctx.GetContractByNetworkAndAddress)
					address.GET(":field", ctx.GetContractByNetworkAndAddressField)
				}
			}
		}
	}
	if err := r.Run(cfg.Address); err != nil {
		fmt.Println(err)
	}
}
