package main

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/verifier/compilation"
	"github.com/baking-bad/bcdhub/internal/verifier/compilers"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/tidwall/gjson"
)

func (ctx *Context) verification(ct compilation.Task) error {
	if err := ctx.verify(ct); err != nil {
		if dbErr := ctx.DB.UpdateTaskStatus(ct.ID, compilation.StatusFailed); dbErr != nil {
			return dbErr
		}

		return err
	}

	return nil
}

func (ctx *Context) verify(ct compilation.Task) error {
	task, err := ctx.DB.GetCompilationTask(ct.ID)
	if err != nil {
		return err
	}

	results := tryToCompile(ct)

	node, err := ctx.GetRPC(task.Network)
	if err != nil {
		return err
	}

	code, err := node.GetCode(task.Address, 0)
	if err != nil {
		return err
	}

	status, res := compareCode(code, results)
	if err != nil {
		return err
	}

	logger.Info("id: %v | kind: %v | status: %s | address: %s | network: %s", ct.ID, ct.Kind, status, task.Address, task.Network)

	if err := ctx.DB.UpdateTaskResults(task, status, res); err != nil {
		return err
	}

	return nil
}

func tryToCompile(task compilation.Task) []database.CompilationTaskResult {
	var result []database.CompilationTaskResult

	for _, filepath := range task.Files {
		var compilationErr string
		jsonb := new(postgres.Jsonb)

		data, err := compilers.BuildFromFile(filepath)
		if err != nil {
			result = append(result, database.CompilationTaskResult{
				CompilationTaskID: task.ID,
				Path:              strings.TrimPrefix(filepath, task.Dir),
				Error:             err.Error(),
			})
			continue
		}

		if err := jsonb.Scan([]byte(data.Script)); err != nil {
			compilationErr = err.Error()
		}

		result = append(result, database.CompilationTaskResult{
			CompilationTaskID: task.ID,
			Path:              strings.TrimPrefix(filepath, task.Dir),
			Script:            jsonb,
			Language:          data.Language,
			Error:             compilationErr,
		})
	}

	return result
}

func compareCode(original gjson.Result, results []database.CompilationTaskResult) (string, []database.CompilationTaskResult) {
	status := compilation.StatusFailed

	for i, r := range results {
		if r.Error != "" {
			results[i].Status = compilation.StatusError
			results[i].Script = new(postgres.Jsonb)
			continue
		}

		script, err := r.Script.Value()
		if err != nil {
			results[i].Status = compilation.StatusError
			results[i].Script = new(postgres.Jsonb)
			results[i].Error = err.Error()
			continue
		}

		eq, err := helpers.AreEqualJSON(original.Raw, string(script.([]byte)))
		if err != nil {
			results[i].Status = compilation.StatusError
			results[i].Script = new(postgres.Jsonb)
			results[i].Error = err.Error()
			continue
		}

		if eq {
			status = compilation.StatusSuccess
			results[i].Status = compilation.StatusSuccess
			continue
		}

		results[i].Status = compilation.StatusMismatch
		results[i].Script = new(postgres.Jsonb)
	}

	return status, results
}
