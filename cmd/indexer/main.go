package main

import (
	"log"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/aopoltorzhicky/bcdhub/internal/db"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
)

func main() {
	var cfg config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.print()

	db, err := db.Database(cfg.Db.URI, cfg.Db.Log)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		panic(err)
	}
	defer messageQueue.Close()

	RPCs := createRPCs(cfg)
	indexers := createIndexers(cfg)

	// Initial syncronization
	if err = sync(RPCs, indexers, db, messageQueue); err != nil {
		log.Println(err)
	}

	// Update state by ticker
	ticker := time.NewTicker(time.Duration(cfg.UpdateTimer) * time.Second)
	for range ticker.C {
		if err = sync(RPCs, indexers, db, messageQueue); err != nil {
			log.Println(err)
		}
	}
}
