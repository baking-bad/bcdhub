package main

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/compiler/compilers"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
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

	contract, err := ctx.ES.GetContract(map[string]interface{}{
		"address": task.Address,
		"network": task.Network,
	})
	if err != nil {
		return err
	}

	return ctx.MQPublisher.Send(mq.ChannelNew, &contract, contract.GetID())
}

func tryToCompile(task compilation.Task) []database.CompilationTaskResult {
	var result []database.CompilationTaskResult

	for _, filepath := range task.Files {
		taskResult := database.CompilationTaskResult{
			CompilationTaskID: task.ID,
			Path:              strings.TrimPrefix(filepath, task.Dir),
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

func compareCode(original gjson.Result, results []database.CompilationTaskResult) (string, []database.CompilationTaskResult) {
	status := compilation.StatusFailed

	for i, r := range results {
		if r.Error != "" {
			finalizeResult(compilation.StatusError, nil, &results[i])
			continue
		}

		script, err := r.Script.Value()
		if err != nil {
			finalizeResult(compilation.StatusError, err, &results[i])
			continue
		}

		eq, err := helpers.AreEqualJSON(original.Raw, string(script.([]byte)))
		if err != nil {
			finalizeResult(compilation.StatusError, err, &results[i])
			continue
		}

		if !eq {
			finalizeResult(compilation.StatusMismatch, nil, &results[i])
			continue
		}

		status = compilation.StatusSuccess
		results[i].Status = compilation.StatusSuccess
	}

	return status, results
}

func finalizeResult(status string, err error, result *database.CompilationTaskResult) {
	result.Status = status
	result.Script = new(postgres.Jsonb)

	if err != nil {
		result.Error = err.Error()
	}
}
