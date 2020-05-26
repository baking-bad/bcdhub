package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func main() {
	createTaskCommand := flag.NewFlagSet("create_tasks", flag.ExitOnError)
	userID := createTaskCommand.Uint("user", 0, "Tasks are generated for user with id")
	offset := createTaskCommand.Int64("offset", -1, "Offset for contracts to generate tasks")
	esConnection := createTaskCommand.String("es", "http://localhost:9200", "Elastic search connection string")
	esTimeout := createTaskCommand.Int("esTimeout", 30, "Elastic search connect timeout")
	dbConnection := createTaskCommand.String("db", "host=127.0.0.1 port=5432 sslmode=disable", "Postgres connection string")
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Println("usage: ml <command> [<args>]")
		fmt.Println("Commands: ")
		fmt.Println(" create_tasks | Create tasks of markup data for user")
		return
	}

	switch os.Args[1] {
	case "create_tasks":
		createTaskCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if createTaskCommand.Parsed() {
		if *userID == 0 {
			fmt.Println("Please supply the user id with -user option.")
			return
		}
		if err := createTasks(*dbConnection, *esConnection, *esTimeout, *userID, *offset); err != nil {
			logger.Error(err)
		}
		return
	}

	logger.Info("Done")
}
