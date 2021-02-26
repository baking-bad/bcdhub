package main

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/providers"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (ctx *Context) verification(ct compilation.Task) error {
	task, err := ctx.verify(ct)
	if err != nil {
		if dbErr := ctx.DB.UpdateTaskStatus(ct.ID, compilation.StatusFailed); dbErr != nil {
			return dbErr
		}

		return err
	}

	user, err := ctx.DB.GetUser(task.UserID)
	if err != nil {
		return err
	}

	var sourcePath string
	for _, r := range task.Results {
		if r.Status == compilation.StatusSuccess {
			if r.AWSPath != "" {
				sourcePath = r.AWSPath
			} else {
				provider, err := providers.NewPublic(user.Provider)
				if err != nil {
					return err
				}
				basePath := provider.BaseFilePath(user.Login, task.Repo, task.Ref)
				sourcePath = basePath + r.Path
			}

			break
		}
	}

	verification := database.Verification{
		UserID:            task.UserID,
		CompilationTaskID: task.ID,
		Address:           task.Address,
		Network:           task.Network,
		SourcePath:        sourcePath,
	}

	if err := ctx.DB.CreateVerification(&verification); err != nil {
		return err
	}

	contract := contract.NewEmptyContract(task.Network, task.Address)
	contract.Verified = true
	contract.VerificationSource = sourcePath

	return ctx.Storage.UpdateFields(models.DocContracts, contract.GetID(), contract, "Verified", "VerificationSource")
}

func (ctx *Context) verify(ct compilation.Task) (*database.CompilationTask, error) {
	task, err := ctx.DB.GetCompilationTask(ct.ID)
	if err != nil {
		return nil, err
	}

	results := compile(ct)

	node, err := ctx.GetRPC(task.Network)
	if err != nil {
		return nil, err
	}

	code, err := node.GetCode(task.Address, 0)
	if err != nil {
		return nil, err
	}

	status, res := compareCode(code, results)
	if err != nil {
		return nil, err
	}

	logger.Info("id: %v | kind: %v | status: %s | address: %s | network: %s", ct.ID, ct.Kind, status, task.Address, task.Network)

	if err := ctx.DB.UpdateTaskResults(task, status, res); err != nil {
		return nil, err
	}

	if status != compilation.StatusSuccess {
		return nil, fmt.Errorf("verification for task_id %v failed", ct.ID)
	}

	return task, nil
}

func compareCode(original *ast.Script, results []database.CompilationTaskResult) (string, []database.CompilationTaskResult) {
	status := compilation.StatusFailed

	for i, r := range results {
		if r.Error != "" {
			finalizeResult(compilation.StatusError, nil, &results[i])
			continue
		}

		val, err := r.Script.Value()
		if err != nil {
			finalizeResult(compilation.StatusError, err, &results[i])
			continue
		}

		var s ast.Script
		if err := json.Unmarshal(val.([]byte), &s); err != nil {
			continue
		}

		if !s.Compare(original) {
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
