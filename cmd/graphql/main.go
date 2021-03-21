package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/dosco/graphjin/core"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jinzhu/gorm"
)

type request struct {
	Query     string          `json:"query"`
	Variables json.RawMessage `json:"variables,omitempty"`
}

func (ctx *apiContext) graphqlHandler(c *gin.Context) {
	var req request
	if err := c.BindJSON(&req); handleError(c, err) {
		return
	}

	res, err := ctx.graphjin.GraphQL(context.Background(), req.Query, req.Variables, nil)
	if handleError(c, err) {
		return
	}

	c.JSON(http.StatusOK, res)
}

func handleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})

	return true
}

type apiContext struct {
	graphjin *core.GraphJin
}

func initUser(connection string) error {
	data, err := ioutil.ReadFile("init.sql")
	if err != nil {
		return err
	}

	expanded := os.ExpandEnv(string(data))
	root, err := gorm.Open("postgres", connection)
	if err != nil {
		return err
	}
	defer root.Close()

	return root.Raw(expanded).Error
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if err := initUser(cfg.Storage.Postgres); err != nil {
		logger.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.GraphQL.DB)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	var production bool
	if env := os.Getenv(config.EnvironmentVar); env == config.EnvironmentProd {
		production = true
	}

	graphjin, err := core.NewGraphJin(&core.Config{
		Production: production,
		Debug:      true,
	}, db)
	if err != nil {
		logger.Fatal(err)
	}

	ctx := apiContext{
		graphjin,
	}
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.POST("/api", ctx.graphqlHandler)

	http.ListenAndServe(":3000", r)
}
