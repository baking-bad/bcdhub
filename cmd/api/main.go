package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/validations"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/periodic"
	"github.com/baking-bad/bcdhub/internal/profiler"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/grafana/pyroscope-go"
)

type app struct {
	Router   *gin.Engine
	Contexts config.Contexts
	Config   config.Config

	cancel   context.CancelFunc
	worker   *periodic.GeneralWorker
	profiler *pyroscope.Profiler
}

func newApp() *app {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		panic(err)
	}

	if cfg.API.SentryEnabled {
		helpers.InitSentry(cfg.Sentry.Debug, cfg.Sentry.Environment, cfg.Sentry.URI)
		helpers.SetTagSentry("project", cfg.API.ProjectName)
		defer helpers.CatchPanicSentry()
	}

	app := new(app)
	app.Config = cfg

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	if cfg.Profiler != nil {
		profiler, err := profiler.New(cfg.Profiler.Server, "api")
		if err != nil {
			panic(err)
		}
		app.profiler = profiler
	}

	if cfg.API.Periodic != nil {
		worker, err := periodic.NewGeneralWorker(*cfg.API.Periodic, app.handleUrlChanged)
		if err != nil {
			panic(err)
		}
		app.worker = worker

		ctx, cancel := context.WithCancel(context.Background())
		app.cancel = cancel
		app.worker.Start(ctx)

		for len(app.worker.URLs()) == 0 {
			time.Sleep(time.Second)
		}
	}

	app.Contexts = config.NewContexts(cfg, cfg.API.Networks,
		config.WithStorage(cfg.Storage, cfg.API.ProjectName, int64(cfg.API.PageSize), cfg.API.Connections.Open, cfg.API.Connections.Idle, false),
		config.WithRPC(cfg.RPC),
		config.WithMempool(cfg.Services),
		config.WithLoadErrorDescriptions(),
		config.WithConfigCopy(cfg))

	app.makeRouter()

	return app
}

func (api *app) makeRouter() {
	r := gin.New()
	store := persistence.NewInMemoryStore(time.Second * 30)

	r.MaxMultipartMemory = 4 << 20 // max upload size 4 MiB
	r.SecureJsonPrefix("")         // do not prepend while(1)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validations.Register(v, api.Config.API); err != nil {
			panic(err)
		}
	}

	if api.Config.API.CorsEnabled {
		r.Use(corsSettings())
	}

	if api.Config.API.SentryEnabled {
		r.Use(helpers.SentryMiddleware())
	}

	r.Use(gin.Recovery())

	if env := os.Getenv(config.EnvironmentVar); env == config.EnvironmentProd {
		r.Use(loggerFormat())
	} else {
		r.Use(gin.Logger())
	}

	v1 := r.Group("v1")
	{
		v1.GET("config", handlers.ContextsMiddleware(api.Contexts), handlers.GetConfig())

		v1.GET("head", handlers.ContextsMiddleware(api.Contexts), cache.CachePage(store, time.Second*10, handlers.GetHead()))
		v1.GET("head/:network", handlers.NetworkMiddleware(api.Contexts), cache.CachePage(store, time.Second*10, handlers.GetHeadByNetwork()))
		opg := v1.Group("opg/:hash")
		{
			opg.GET("", handlers.ContextsMiddleware(api.Contexts), handlers.GetOperation())
			opg.GET(":counter", handlers.ContextsMiddleware(api.Contexts), handlers.GetByHashAndCounter())
		}
		v1.GET("implicit/:network/:counter", handlers.NetworkMiddleware(api.Contexts), handlers.GetImplicitOperation())
		v1.POST("off_chain_view", handlers.MainnetMiddleware(api.Contexts), handlers.OffChainView())
		v1.POST("michelson", handlers.ContextsMiddleware(api.Contexts), handlers.CodeFromMichelson())
		v1.POST("fork", handlers.ForkContract(api.Contexts))

		operation := v1.Group("operation/:network/:id")
		operation.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			operation.GET("error_location", handlers.GetOperationErrorLocation())
			operation.GET("diff", handlers.GetOperationDiff())
			operation.GET("ticket_updates", handlers.GetTicketUpdatesForOperation())
		}

		stats := v1.Group("stats")
		{
			stats.GET("", handlers.ContextsMiddleware(api.Contexts), cache.CachePage(store, time.Second*30, handlers.GetStats()))

			networkStats := stats.Group(":network")
			networkStats.Use(handlers.NetworkMiddleware(api.Contexts))
			{
				networkStats.GET("recently_called_contracts", cache.CachePage(store, time.Second*10, handlers.RecentlyCalledContracts()))
				networkStats.GET("contracts_count", cache.CachePage(store, time.Second*10, handlers.ContractsCount()))
			}
		}

		helpers := v1.Group("helpers")
		{
			helpers.GET("contracts/:network", handlers.NetworkMiddleware(api.Contexts), cache.CachePage(store, time.Hour, handlers.ContractsHelpers()))
		}

		bigmap := v1.Group("bigmap/:network/:ptr")
		bigmap.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			bigmap.GET("", cache.CachePage(store, time.Second*30, handlers.GetBigMap()))
			bigmap.GET("count", handlers.GetBigMapDiffCount())
			bigmap.GET("history", handlers.GetBigMapHistory())
			keys := bigmap.Group("keys")
			{
				keys.GET("", handlers.GetBigMapKeys())
				keys.GET(":key_hash", handlers.GetBigMapByKeyHash())
				keys.GET(":key_hash/state", handlers.GetCurrentBigMapKeyHash())
			}
		}

		contract := v1.Group("contract/:network/:address")
		contract.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			contract.GET("", handlers.ContextsMiddleware(api.Contexts), handlers.GetContract())
			contract.GET("code", handlers.GetContractCode())
			contract.GET("operations", handlers.GetContractOperations())
			contract.GET("opg", handlers.GetOperationGroups())
			contract.GET("migrations", handlers.GetContractMigrations())
			contract.GET("global_constants", handlers.GetContractGlobalConstants())
			contract.GET("ticket_updates", handlers.GetContractTicketUpdates())
			contract.GET("events", handlers.ListEvents())

			storage := contract.Group("storage")
			{
				storage.GET("", handlers.GetContractStorage())
				storage.GET("raw", handlers.GetContractStorageRaw())
				storage.GET("rich", handlers.GetContractStorageRich())
				storage.GET("schema", handlers.GetContractStorageSchema())
			}

			contract.GET("mempool", handlers.GetMempool())
			contract.GET("same", handlers.ContextsMiddleware(api.Contexts), handlers.GetSameContracts())
			entrypoints := contract.Group("entrypoints")
			{
				entrypoints.GET("", handlers.GetEntrypoints())
				entrypoints.GET("schema", handlers.GetEntrypointSchema())
				entrypoints.POST("data", handlers.GetEntrypointData())
				entrypoints.POST("trace", handlers.RunCode())
				entrypoints.POST("run_operation", handlers.RunOperation())
			}
			views := contract.Group("views")
			{
				views.GET("schema", handlers.GetViewsSchema())
				views.POST("execute", handlers.ExecuteView())
			}
		}

		account := v1.Group("account/:network")
		account.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			acc := account.Group(":address")
			{
				acc.GET("", handlers.GetInfo())
			}
		}

		globalConstants := v1.Group("global_constants/:network")
		globalConstants.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			globalConstants.GET("", handlers.ListGlobalConstants())
			globalConstant := globalConstants.Group(":address")
			{
				globalConstant.GET("", handlers.GetGlobalConstant())
				globalConstant.GET("contracts", handlers.GetGlobalConstantContracts())
			}
		}

		smartRollups := v1.Group("smart_rollups/:network")
		smartRollups.Use(handlers.NetworkMiddleware(api.Contexts))
		{
			smartRollups.GET("", handlers.ListSmartRollups())
			smartRollups.GET(":address", handlers.GetSmartRollup())
		}
	}
	api.Router = r
}

// Close -
func (api *app) Close() error {
	api.cancel()

	if api.profiler != nil {
		if err := api.profiler.Stop(); err != nil {
			return err
		}
	}

	if err := api.Contexts.Close(); err != nil {
		return err
	}
	return api.worker.Close()
}

// Run -
func (api *app) Run() {
	if err := api.Router.Run(api.Config.API.Bind); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.Err(err)
		helpers.CatchErrorSentry(err)
	}
}

// @title Better Call Dev API
// @description This is API description for Better Call Dev service.

// @contact.name Baking Bad Team
// @contact.url https://baking-bad.org/docs
// @contact.email hello@baking-bad.org

// @x-logo {"url": "https://better-call.dev/img/logo_og.png", "altText": "Better Call Dev logo", "href": "https://better-call.dev"}

// @query.collection.format multi
func main() {
	api := newApp()
	defer api.Close()

	api.Run()
}

func corsSettings() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH"},
		AllowHeaders:     []string{"X-Requested-With", "Authorization", "Origin", "Content-Length", "Content-Type", "Referer", "Cache-Control", "User-Agent"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func loggerFormat() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%15s | %3d | %13v | %-7s %s | %s\n%s",
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func (api *app) handleUrlChanged(ctx context.Context, network, url string) error {
	if value, ok := api.Config.RPC[network]; ok {
		value.URI = url
		api.Config.RPC[network] = value
	}

	if _, ok := api.Config.API.Frontend.RPC[network]; ok {
		api.Config.API.Frontend.RPC[network] = url
	}

	return nil
}
