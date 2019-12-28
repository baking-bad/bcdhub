package db

import (
	"github.com/jinzhu/gorm"

	"github.com/aopoltorzhicky/bcdhub/internal/db/account"
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/db/project"
	"github.com/aopoltorzhicky/bcdhub/internal/db/relation"
	"github.com/aopoltorzhicky/bcdhub/internal/db/state"
	"github.com/aopoltorzhicky/bcdhub/internal/db/tags"
)

// Database - Open database connection
func Database(connectionString string, log bool) (*gorm.DB, error) {
	//open a db connection
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	db.LogMode(log)

	db.AutoMigrate(&contract.Contract{}, &state.State{}, &project.Project{}, &tags.Tag{}, &relation.Relation{}, &account.Account{})

	db.Exec("SET pg_trgm.similarity_threshold = 0.8")
	db = db.Set("gorm:auto_preload", true)

	return db, nil
}
