package periodic

import (
	"context"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/teztnets"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Worker -
type Worker struct {
	network    types.Network
	rpc        *teztnets.RPC
	schedule   string
	cron       *cron.Cron
	currentUrl string
	handler    ChangedHandler
}

// ChangedHandler -
type ChangedHandler func(ctx context.Context, network, newUrl string) error

// New -
func New(cfg Config, network types.Network, handler ChangedHandler) (*Worker, error) {
	w := &Worker{
		cron: cron.New(
			cron.WithParser(cron.NewParser(
				cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
			)),
		),
		schedule: cfg.Schedule,
		network:  network,
		handler:  handler,
	}

	rpc, err := teztnets.New(cfg.InfoBaseURL)
	if err != nil {
		return w, err
	}
	w.rpc = rpc

	return w, nil
}

// Start -
func (w *Worker) Start(ctx context.Context) {
	if w.handler == nil {
		return
	}

	if _, err := w.checkNetwork(ctx); err != nil {
		log.Err(err).Str("network", w.network.String()).Msg("failed to receive periodic network info")
		return
	}

	if _, err := w.cron.AddFunc(
		w.schedule,
		w.handleScheduleEvent(ctx),
	); err != nil {
		log.Err(err).Str("network", w.network.String()).Msg("failed to run cron function")
		return
	}

	w.cron.Start()
}

// Close -
func (w *Worker) Close() error {
	w.cron.Stop()
	return nil
}

func (w *Worker) handleScheduleEvent(ctx context.Context) func() {
	return func() {
		log.Info().Str("network", w.network.String()).Msg("trying to receive new rpc url")

		changed, err := w.checkNetwork(ctx)
		if err != nil {
			log.Err(err).Str("network", w.network.String()).Msg("failed to receive periodic network info")
		}
		if changed {
			log.Info().Msg("rpc url changed")
			return
		}

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				changed, err := w.checkNetwork(ctx)
				if err != nil {
					log.Err(err).Str("network", w.network.String()).Msg("failed to receive periodic network info")
				}
				if changed {
					log.Info().Msg("rpc url changed")
					return
				}
			}
		}
	}
}

func (w *Worker) checkNetwork(ctx context.Context) (bool, error) {
	info, err := w.rpc.Teztnets(ctx)
	if err != nil {
		return false, err
	}

	for name, data := range info {
		parts := strings.Split(name, "-")
		if len(parts) == 0 {
			continue
		}

		if parts[0] != w.network.String() {
			continue
		}

		if w.currentUrl != data.RPCURL {
			if w.currentUrl != "" {
				if err := w.handler(ctx, w.network.String(), data.RPCURL); err != nil {
					log.Err(err).Str("network", w.network.String()).Msg("failed to apply new rpc url")
				}
			}
			w.currentUrl = data.RPCURL

			log.Info().Str("network", parts[0]).Str("url", w.currentUrl).Msg("new url was found")
			return true, nil
		}
	}

	return false, nil
}

// URL -
func (w *Worker) URL() string {
	return w.currentUrl
}
