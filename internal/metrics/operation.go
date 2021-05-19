package metrics

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/getsentry/sentry-go"
)

// SetOperationStrings -
func (h *Handler) SetOperationStrings(op *operation.Operation) {
	ps, err := getStrings(op.Parameters)
	if err == nil {
		op.ParameterStrings = ps
	}
	ss, err := getStrings(op.DeffatedStorage)
	if err == nil {
		op.StorageStrings = ss
	}
}

// SendSentryNotifications -
func (h *Handler) SendSentryNotifications(op operation.Operation) error {
	if op.Status != "failed" {
		return nil
	}

	subscriptions, err := h.DB.GetSubscriptions(op.Destination, op.Network)
	if err != nil {
		return err
	}
	if len(subscriptions) == 0 {
		return nil
	}

	defer sentry.Flush(2 * time.Second)

	for _, subscription := range subscriptions {
		initSentry(op.Network.String(), subscription.SentryDSN)

		hub := sentry.CurrentHub().Clone()
		tags := map[string]string{
			"hash":    op.Hash,
			"source":  op.Source,
			"address": op.Destination,
			"kind":    op.Kind,
			"block":   fmt.Sprintf("%d", op.Level),
			"os.name": "tezos",
		}

		if op.Entrypoint != "" {
			tags["entrypoint"] = op.Entrypoint
		}

		exceptions := make([]sentry.Exception, 0)
		var message string
		for i := range op.Errors {
			if err := op.Errors[i].Format(); err != nil {
				return err
			}

			if i == 0 {
				message = op.Errors[i].GetTitle()
			}

			exceptions = append(exceptions, sentry.Exception{
				Value: op.Errors[i].String(),
				Type:  op.Errors[i].GetTitle(),
			})
		}

		hub.Client().Transport.SendEvent(&sentry.Event{
			Tags:        tags,
			Timestamp:   op.Timestamp.Unix(),
			Level:       sentry.LevelError,
			Environment: op.Network.String(),
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
		logger.Errorf("Sentry initialization failed: %v", err)
	}
}
