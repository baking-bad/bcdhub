package main

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/compiler/compilers"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func compile(task compilation.Task) []database.CompilationTaskResult {
	result := make([]database.CompilationTaskResult, 0)

	for _, filepath := range task.Files {
		path := strings.TrimPrefix(filepath, task.Dir)

		if task.Kind == compilation.KindVerification {
			pathParts := strings.SplitAfterN(path, "/", 3)
			if len(pathParts) != 3 {
				continue
			}

			path = pathParts[2]
		}

		taskResult := database.CompilationTaskResult{
			CompilationTaskID: task.ID,
			Path:              path,
		}

		data, err := compilers.BuildFromFile(filepath)

		if err != nil {
			taskResult.Error = err.Error()
		} else {
			jsonb := new(postgres.Jsonb)

			if err := jsonb.Scan([]byte(data.Script)); err != nil {
				taskResult.Error = err.Error()
			}

			taskResult.Script = jsonb
			taskResult.Language = data.Language
		}

		result = append(result, taskResult)
	}

	return result
}
