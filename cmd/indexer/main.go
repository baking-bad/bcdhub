package main

import (
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var states = map[string]*models.State{}

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	// es, err := elastic.New([]string{cfg.Search.URI})
	// if err != nil {
	// 	panic(err)
	// }

	// if err := es.CreateIndexIfNotExists(elastic.DocContracts); err != nil {
	// 	panic(err)
	// }

	RPCs := createRPCs(cfg)
	// indexers, err := createIndexers(es, cfg)
	// if err != nil {
	// 	panic(err)
	// }

	// // Initial syncronization
	// if err = sync(RPCs, indexers, es); err != nil {
	// 	logger.Error(err)
	// }

	// // Update state by ticker
	// ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	// for range ticker.C {
	// 	if err = sync(RPCs, indexers, es); err != nil {
	// 		logger.Error(err)
	// 	}
	// }

	scr, err := RPCs["mainnet"].GetScriptJSON("KT1REHQ183LzfoVoqiDR87mCrt7CLUH1MbcV", 0)
	if err != nil {
		panic(err)
	}

	// code := scr.Get("code.#(prim==code)")
	s, err := macros.FindMacros(scr.Get("code"))
	if err != nil {
		panic(err)
	}
	log.Println(s)
}
