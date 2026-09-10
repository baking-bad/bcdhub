package periodic

import (
	"context"
	"maps"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baking-bad/bcdhub/internal/teztnets"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// GeneralWorker -
type GeneralWorker struct {
	rpc       *teztnets.RPC
	schedule  string
	cron      *cron.Cron
	handler   ChangedHandler
	urls      map[string]string
	mx        sync.RWMutex
	isRunning atomic.Bool
}

// NewGeneralWorker -
func NewGeneralWorker(cfg Config, handler ChangedHandler) (*GeneralWorker, error) {
	w := &GeneralWorker{
		cron: cron.New(
			cron.WithParser(cron.NewParser(
				cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
			)),
		),
		schedule: cfg.Schedule,
		handler:  handler,
		urls:     make(map[string]string),
	}

	rpc, err := teztnets.New(cfg.InfoBaseURL)
	if err != nil {
		return w, err
	}
	w.rpc = rpc

	return w, nil
}

// Start -
func (w *GeneralWorker) Start(ctx context.Context) {
	if w.handler == nil {
		return
	}

	if _, err := w.checkNetwork(ctx); err != nil {
		log.Err(err).Msg("failed to receive periodic network info")
		return
	}

	if _, err := w.cron.AddFunc(
		w.schedule,
		w.handleScheduleEvent(ctx),
	); err != nil {
		log.Err(err).Msg("failed to run cron function")
		return
	}

	w.cron.Start()
}

// Close -
func (w *GeneralWorker) Close() error {
	w.cron.Stop()
	return nil
}

func (w *GeneralWorker) handleScheduleEvent(ctx context.Context) func() {
	return func() {
		if !w.isRunning.CompareAndSwap(false, true) {
			log.Debug().Msg("periodic worker is already running")
			return
		}
		defer w.isRunning.Store(false)

		log.Info().Msg("trying to receive new rpc url")

		changed, err := w.checkNetwork(ctx)
		if err != nil {
			log.Err(err).Msg("failed to receive periodic network info")
		}
		if changed {
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
					log.Err(err).Msg("failed to receive periodic network info")
				}
				if changed {
					return
				}
			}
		}
	}
}

func (w *GeneralWorker) checkNetwork(ctx context.Context) (bool, error) {
	info, err := w.rpc.Teztnets(ctx)
	if err != nil {
		return false, err
	}

	var changed bool

	for name, data := range info {
		parts := strings.Split(name, "-")
		if len(parts) == 0 {
			continue
		}

		network := parts[0]

		w.mx.RLock()
		current := w.urls[network]
		w.mx.RUnlock()

		if current == data.RPCURL {
			continue
		}

		// the handler re-points the service at the new node, so it must not run while
		// the mutex that URLs() needs is held
		if err := w.handler(ctx, network, data.RPCURL); err != nil {
			// leave the stored url untouched so that the next poll retries the switch
			// instead of silently dropping the network
			log.Err(err).Str("network", network).Msg("failed to apply new rpc url")
			continue
		}

		w.mx.Lock()
		w.urls[network] = data.RPCURL
		w.mx.Unlock()

		log.Info().Str("network", network).Str("url", data.RPCURL).Msg("new url was found")
		changed = true
	}

	return changed, nil
}

// URLs -
func (w *GeneralWorker) URLs() map[string]string {
	w.mx.RLock()
	defer w.mx.RUnlock()
	urls := make(map[string]string, len(w.urls))
	maps.Copy(urls, w.urls)
	return urls
}
