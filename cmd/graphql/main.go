package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/dosco/graphjin/core"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
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

func initUser() error {
	connection := os.ExpandEnv("host=$DB_HOSTNAME port=5432 user=$POSTGRES_USER dbname=indexer password=$POSTGRES_PASSWORD sslmode=disable")

	data, err := ioutil.ReadFile("init.sql")
	if err != nil {
		return err
	}

	expanded := os.ExpandEnv(string(data))
	expanded = strings.ReplaceAll(expanded, "{dlr}", "$") // dirty hack for escaping dollar sign

	root, err := sql.Open("pgx", connection)
	if err != nil {
		return err
	}
	defer root.Close()

	_, err = root.Exec(expanded)
	return err
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if err := initUser(); err != nil {
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

	if err := http.ListenAndServe(":3000", r); err != nil {
		panic(err)
	}
}
