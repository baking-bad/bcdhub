package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	"github.com/tidwall/gjson"
)

// SetOperationAliases -
func (h *Handler) SetOperationAliases(aliases map[string]string, op *models.Operation) bool {
	var changed bool

	if srcAlias, ok := aliases[op.Source]; ok {
		op.SourceAlias = srcAlias
		changed = true
	}

	if dstAlias, ok := aliases[op.Destination]; ok {
		op.DestinationAlias = dstAlias
		changed = true
	}

	if dlgtAlias, ok := aliases[op.Delegate]; ok {
		op.DelegateAlias = dlgtAlias
		changed = true
	}

	return changed
}

// SetOperationBurned -
func (h *Handler) SetOperationBurned(op *models.Operation) {
	if op.Status != consts.Applied {
		return
	}

	if op.Result == nil {
		return
	}

	var burned int64

	if op.Result.PaidStorageSizeDiff != 0 {
		burned += op.Result.PaidStorageSizeDiff * 1000
	}

	if op.Result.AllocatedDestinationContract {
		burned += 257000
	}

	op.Burned = burned
}

// SetOperationStrings -
func (h *Handler) SetOperationStrings(op *models.Operation) {
	params := gjson.Parse(op.Parameters)
	op.ParameterStrings = stringer.Get(params)
	storage := gjson.Parse(op.DeffatedStorage)
	op.StorageStrings = stringer.Get(storage)
}

// SendSentryNotifications -
func (h *Handler) SendSentryNotifications(operation models.Operation) error {
	if operation.Status != "failed" {
		return nil
	}

	subscriptions, err := h.DB.GetSubscriptions(operation.Destination, operation.Network)
	if err != nil {
		return err
	}
	if len(subscriptions) == 0 {
		return nil
	}

	defer sentry.Flush(2 * time.Second)

	for _, subscription := range subscriptions {
		initSentry(operation.Network, subscription.SentryDSN)

		hub := sentry.CurrentHub().Clone()
		tags := map[string]string{
			"hash":    operation.Hash,
			"source":  operation.Source,
			"address": operation.Destination,
			"kind":    operation.Kind,
			"block":   fmt.Sprintf("%d", operation.Level),
			"os.name": "tezos",
		}

		if operation.Entrypoint != "" {
			tags["entrypoint"] = operation.Entrypoint
		}

		exceptions := make([]sentry.Exception, 0)
		var message string
		for i := range operation.Errors {
			if err := operation.Errors[i].Format(); err != nil {
				return err
			}

			if i == 0 {
				message = operation.Errors[i].GetTitle()
			}

			exceptions = append(exceptions, sentry.Exception{
				Value: operation.Errors[i].String(),
				Type:  operation.Errors[i].GetTitle(),
			})
		}

		hub.Client().Transport.SendEvent(&sentry.Event{
			Tags:        tags,
			Timestamp:   operation.Timestamp.Unix(),
			Level:       sentry.LevelError,
			Environment: operation.Network,
			Message:     message,
			Exception:   exceptions,
			Sdk: sentry.SdkInfo{
				Name:    "BCD tezos SDK",
				Version: "0.1",
			},
		})
	}
	return nil
}

func initSentry(environment, dsn string) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Debug:            false,
		AttachStacktrace: false,
	}); err != nil {
		log.Printf("Sentry initialization failed: %v\n", err)
	}
}

// SetOperationDeployment -
func (h *Handler) SetOperationDeployment(op *models.Operation) error {
	d, err := h.DB.GetDeploymentBy(op.Hash)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return err
	}

	d.Address = op.Destination
	d.Network = op.Network

	if err := h.DB.UpdateDeployment(d); err != nil {
		return err
	}

	task, err := h.DB.GetCompilationTask(d.CompilationTaskID)
	if err != nil {
		return err
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
		CompilationTaskID: d.CompilationTaskID,
		Address:           op.Destination,
		Network:           op.Network,
		SourcePath:        sourcePath,
	}

	if err := h.DB.CreateVerification(&verification); err != nil {
		return err
	}

	by := map[string]interface{}{
		"address": op.Destination,
		"network": op.Network,
	}

	contract, err := h.ES.GetContract(by)
	if err != nil {
		return err
	}

	if !contract.Verified {
		if err := h.SetContractVerification(&contract); err != nil {
			return err
		}

		return h.ES.UpdateFields(elastic.DocContracts, contract.ID, contract, "Verified", "VerificationSource")
	}

	return nil
}
