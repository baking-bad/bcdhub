package periodic

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
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
	mx         sync.RWMutex
	handler    ChangedHandler
	isRunning  atomic.Bool
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
		if !w.isRunning.CompareAndSwap(false, true) {
			log.Debug().Str("network", w.network.String()).Msg("periodic worker is already running")
			return
		}
		defer w.isRunning.Store(false)

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

		current := w.URL()
		if current == data.RPCURL {
			continue
		}

		// the handler tears the indexer down and re-initialises it, so it must not run
		// while the mutex that URL() needs is held
		if current != "" {
			if err := w.handler(ctx, w.network.String(), data.RPCURL); err != nil {
				// leave currentUrl untouched so that the next poll retries the switch
				// instead of silently dropping the network
				log.Err(err).Str("network", w.network.String()).Msg("failed to apply new rpc url")
				return false, nil
			}
		}

		w.mx.Lock()
		w.currentUrl = data.RPCURL
		w.mx.Unlock()

		log.Info().Str("network", parts[0]).Str("url", data.RPCURL).Msg("new url was found")
		return true, nil
	}

	return false, nil
}

// URL -
func (w *Worker) URL() string {
	w.mx.RLock()
	defer w.mx.RUnlock()
	return w.currentUrl
}
