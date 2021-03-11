package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/mq"
)

// Context -
type Context struct {
	*config.Context
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if cfg.Compiler.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.Compiler.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	context := &Context{
		config.NewContext(
			config.WithRPC(cfg.RPC),
			config.WithDatabase(cfg.DB),
			config.WithRabbit(cfg.RabbitMQ, cfg.Compiler.ProjectName, cfg.Compiler.MQ),
			config.WithStorage(cfg.Storage),
			config.WithAWS(cfg.Compiler.AWS),
		),
	}

	defer context.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	protocol, err := context.Protocols.GetProtocol(consts.Mainnet, "", -1)
	if err != nil {
		logger.Fatal(err)
	}

	tickerTime := protocol.Constants.TimeBetweenBlocks
	if tickerTime == 0 {
		tickerTime = 30
	}
	ticker := time.NewTicker(time.Second * time.Duration(tickerTime))

	msgs, err := context.MQ.Consume(mq.QueueCompilations)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Connected to %s queue", mq.QueueCompilations)

	for {
		select {
		case <-signals:
			logger.Info("Stopped compiler")
			return
		case <-ticker.C:
			if err := context.setDeployment(); err != nil {
				logger.Error(err)
			}
		case msg := <-msgs:
			if err := context.handleMessage(msg); err != nil {
				logger.Error(err)
			}
		}
	}

}

func (ctx *Context) setDeployment() error {
	deployments, err := ctx.DB.GetDeploymentsByAddressNetwork("", "")
	if err != nil {
		return err
	}

	for i, d := range deployments {
		ops, err := ctx.Operations.Get(
			map[string]interface{}{"hash": d.OperationHash},
			0,
			true,
		)

		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				continue
			}

			return fmt.Errorf("GetOperations %s error %w", d.OperationHash, err)
		}

		if len(ops) == 0 {
			continue
		}

		if err := ctx.processDeployment(&deployments[i], &ops[0]); err != nil {
			return fmt.Errorf("deployment ID %d operationHash %s processDeployment error %w", d.ID, d.OperationHash, err)
		}
	}

	return nil
}

func (ctx *Context) processDeployment(deployment *database.Deployment, operation *operation.Operation) error {
	deployment.Address = operation.Destination
	deployment.Network = operation.Network

	if err := ctx.DB.UpdateDeployment(deployment); err != nil {
		return fmt.Errorf("UpdateDeployment error %w", err)
	}

	task, err := ctx.DB.GetCompilationTask(deployment.CompilationTaskID)
	if err != nil {
		return fmt.Errorf("task ID %d GetCompilationTask error %w", deployment.CompilationTaskID, err)
	}

	var sourcePath string

	for _, r := range task.Results {
		if r.Status == compilation.StatusSuccess {
			sourcePath = r.AWSPath
			break
		}
	}

	verification := database.Verification{
		UserID:            task.UserID,
		CompilationTaskID: deployment.CompilationTaskID,
		Address:           operation.Destination,
		Network:           operation.Network,
		SourcePath:        sourcePath,
	}

	if err := ctx.DB.CreateVerification(&verification); err != nil {
		return fmt.Errorf("CreateVerification error %w", err)
	}

	contract := contract.NewEmptyContract(task.Network, task.Address)
	contract.Verified = true
	contract.VerificationSource = sourcePath

	return ctx.Storage.UpdateFields(models.DocContracts, contract.GetID(), contract, "Verified", "VerificationSource")
}

func (ctx *Context) handleMessage(data mq.Data) error {
	if err := ctx.parseData(data); err != nil {
		return err
	}

	return data.Ack(false)
}

func (ctx *Context) parseData(data mq.Data) error {
	if data.GetKey() != mq.QueueCompilations {
		logger.Warning("[parseData] Unknown data routing key %s", data.GetKey())
		return data.Ack(false)
	}

	var ct compilation.Task
	if err := json.Unmarshal(data.GetBody(), &ct); err != nil {
		return fmt.Errorf("[parseData] Unmarshal message body error: %s", err)
	}

	defer os.RemoveAll(ct.Dir) // clean up

	switch ct.Kind {
	case compilation.KindVerification:
		return ctx.verification(ct)
	case compilation.KindDeployment:
		return ctx.deployment(ct)
	}

	return fmt.Errorf("[parseData] Unknown compilation task kind %s", ct.Kind)
}
