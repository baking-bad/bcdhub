package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/database"
)

func (ctx *Context) deployment(ct compilation.Task) error {
	if err := ctx.deploy(ct); err != nil {
		if dbErr := ctx.DB.UpdateTaskStatus(ct.ID, compilation.StatusFailed); dbErr != nil {
			return dbErr
		}

		return err
	}

	return nil
}

func (ctx *Context) deploy(ct compilation.Task) error {
	task, err := ctx.DB.GetCompilationTask(ct.ID)
	if err != nil {
		return err
	}

	results := compile(ct)
	if len(results) == 0 {
		return fmt.Errorf("no files in compilation results %v", ct)
	}

	status, res := prepareResults(results)

	for i := range res {
		if err := ctx.uploadSource(&res[i], ct.Dir); err != nil {
			return err
		}
	}

	if err := ctx.DB.UpdateTaskResults(task, status, res); err != nil {
		return err
	}

	return nil
}

func prepareResults(results []database.CompilationTaskResult) (string, []database.CompilationTaskResult) {
	status := compilation.StatusFailed

	for i, r := range results {
		if r.Error != "" {
			finalizeResult(compilation.StatusError, nil, &results[i])
			continue
		}

		status = compilation.StatusSuccess
		results[i].Status = compilation.StatusSuccess
	}

	return status, results
}

func (ctx *Context) uploadSource(result *database.CompilationTaskResult, dir string) error {
	if result.Status != compilation.StatusSuccess {
		return nil
	}

	srcPath := filepath.Join(dir, result.Path)
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	awsFilename := filepath.Join(filepath.Base(dir), result.Path)

	res, err := ctx.AWS.Upload(file, awsFilename)
	if err != nil {
		return err
	}

	result.AWSPath = res.Location

	return nil
}
