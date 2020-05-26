package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/schollz/progressbar/v3"
)

func createTasks(dbConn, esConn string, esTimeout int, userID uint, offset int64) error {
	es := elastic.WaitNew([]string{esConn}, esTimeout)

	fullDBConn, err := askDatabaseConnectionString(dbConn)
	if err != nil {
		return err
	}

	db, err := database.New(fullDBConn)
	if err != nil {
		return err
	}
	defer db.Close()

	diffTasks, err := es.GetDiffTasks(offset)
	if err != nil {
		return err
	}

	bar := progressbar.NewOptions(len(diffTasks), progressbar.OptionSetPredictTime(false))
	for _, diff := range diffTasks {
		bar.Add(1)
		if err := db.CreateOrUpdateAssessment(diff.Address1, diff.Network1, diff.Address2, diff.Network2, userID, 10); err != nil {
			fmt.Print("\033[2K\r")
			return err
		}
	}
	fmt.Print("\033[2K\r")

	return nil
}
