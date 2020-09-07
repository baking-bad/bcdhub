package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/getsentry/sentry-go"
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
