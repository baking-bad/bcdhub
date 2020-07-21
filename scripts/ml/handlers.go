package main

import (
	"fmt"
	"math/rand"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/schollz/progressbar/v3"
)

func createTasks(dbConn, esConn string, esTimeout int, userID uint, offset, size int) error {
	es := elastic.WaitNew([]string{esConn}, esTimeout)

	db, err := database.New(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()

	allTasks, err := es.GetDiffTasks()
	if err != nil {
		return err
	}

	rand.Seed(42)
	rand.Shuffle(len(allTasks), func(i, j int) { allTasks[i], allTasks[j] = allTasks[j], allTasks[i] })

	fmt.Printf("Total %d pairs, picking %d:%d\n", len(allTasks), offset, offset+size)

	tasks := allTasks[offset : offset+size]

	bar := progressbar.NewOptions(len(tasks), progressbar.OptionSetPredictTime(false))
	for _, diff := range tasks {
		bar.Add(1) //nolint
		a := database.Assessments{
			Address1:   diff.Address1,
			Network1:   diff.Network1,
			Address2:   diff.Address2,
			Network2:   diff.Network2,
			UserID:     userID,
			Assessment: database.AssessmentUndefined,
		}
		if err := db.CreateAssessment(&a); err != nil {
			fmt.Print("\033[2K\r")
			return err
		}
	}
	fmt.Print("\033[2K\r")

	return nil
}
