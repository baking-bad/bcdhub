package main

import (
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var states = map[string]models.State{}

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	es, err := elastic.New([]string{cfg.Search.URI})
	if err != nil {
		panic(err)
	}

	if err := es.CreateIndexIfNotExists(elastic.DocContracts); err != nil {
		panic(err)
	}

	RPCs := createRPCs(cfg)

	// contract, err := RPCs["mainnet"].GetContract("KT1REHQ183LzfoVoqiDR87mCrt7CLUH1MbcV")
	// if err != nil {
	// 	panic(err)
	// }

	// script, err := contractparser.New(contract)
	// if err != nil {
	// 	panic(err)
	// }
	// if err := script.Parse(); err != nil {
	// 	panic(err)
	// }

	// b, err := json.MarshalIndent(script.Code.Parameter.Metadata, "", " ")
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println(string(b))
	// log.Println(script.Code.Parameter.Hash)

	indexers, err := createIndexers(es, cfg)
	if err != nil {
		panic(err)
	}

	// Initial syncronization
	if err = sync(RPCs, indexers, es); err != nil {
		log.Println(err)
	}

	// // Update state by ticker
	// ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	// for range ticker.C {
	// 	if err = sync(RPCs, indexers, es); err != nil {
	// 		log.Println(err)
	// 	}
	// }
}
