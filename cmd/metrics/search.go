package main

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/search"
)

func saveSearchModels(searcher search.Searcher, items []models.Model) error {
	data := search.Prepare(items)
	return searcher.Save(data)
}
