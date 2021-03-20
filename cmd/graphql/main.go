package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/dosco/graphjin/core"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type request struct {
	Query     string          `json:"query"`
	Variables json.RawMessage `json:"variables,omitempty"`
}

func graphqlHandler(graphjinn *core.GraphJin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req request
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), core.UserIDKey, 0)

		res, err := graphjinn.GraphQL(ctx, req.Query, req.Variables, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		render.JSON(w, r, res)
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.Storage.Postgres)
	if err != nil {
		panic(err)
	}

	graphjin, err := core.NewGraphJin(&core.Config{
		Debug: true,
	}, db)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Post("/api", graphqlHandler(graphjin))

	http.ListenAndServe(":3000", r)
}
