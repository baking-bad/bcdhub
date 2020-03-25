package main

import (
	"github.com/baking-bad/bcdhub/cmd/migration/migrations"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/jsonload"
)

func main() {
	var cfg migrations.Config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		panic(err)
	}
	cfg.Print()

	db, err := database.New(cfg.DB.URI)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, err := migrations.NewContext(cfg, db)
	if err != nil {
		panic(err)
	}

	// migration := migrations.SetTimestampMigration{}
	// migration := migrations.SetLanguageMigration{}
	migration := migrations.SetAliasMigration{}
	if err := migration.Do(ctx); err != nil {
		panic(err)
	}
}
