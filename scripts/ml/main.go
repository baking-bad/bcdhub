package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func getEnvInt(name string, required bool, defaultValue int) (int, error) {
	valStr := os.Getenv(name)
	var val int
	var err error
	if valStr == "" {
		if required {
			return val, fmt.Errorf("Please, set %s env variable", name)
		}
		fmt.Printf("%s env variable is not set, will use default\n", name)
		val = defaultValue
	} else {
		val, err = strconv.Atoi(valStr)
		if err != nil {
			return val, err
		}
	}
	if val < defaultValue {
		return val, fmt.Errorf("%s should be >= %d", name, defaultValue)
	}
	fmt.Printf("%s = %d\n", name, val)
	return val, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	userID, err := getEnvInt("USER_ID", true, 1)
	if err != nil {
		logger.Fatal(err)
	}

	offset, err := getEnvInt("OFFSET", false, 0)
	if err != nil {
		logger.Fatal(err)
	}

	size, err := getEnvInt("SIZE", false, 10)
	if err != nil {
		logger.Fatal(err)
	}

	if err := createTasks(cfg.DB.ConnString, cfg.Elastic.URI, uint(userID), offset, size); err != nil {
		logger.Error(err)
	}

	logger.Info("Done")
}
