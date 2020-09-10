package config

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// ContextOption -
type ContextOption func(ctx *Context)

// WithRPC -
func WithRPC(rpcConfig map[string]RPCConfig) ContextOption {
	return func(ctx *Context) {
		if len(rpcConfig) == 0 {
			panic("RPC config is invalid")
		}
		rpc := make(map[string]noderpc.INode)
		for network, rpcProvider := range rpcConfig {
			rpc[network] = noderpc.NewPool(
				[]string{rpcProvider.URI},
				noderpc.WithTimeout(time.Second*time.Duration(rpcProvider.Timeout)),
			)
		}
		ctx.RPC = rpc
	}
}

// WithElasticSearch -
func WithElasticSearch(esConfig ElasticSearchConfig) ContextOption {
	return func(ctx *Context) {
		ctx.ES = elastic.WaitNew([]string{esConfig.URI}, esConfig.Timeout)
	}
}

// WithDatabase -
func WithDatabase(dbConfig DatabaseConfig) ContextOption {
	return func(ctx *Context) {
		db, err := database.New(dbConfig.ConnString)
		if err != nil {
			log.Panicf("Database connection error: %s", err)
		}
		ctx.DB = db
	}
}

// WithRabbitReceiver -
func WithRabbitReceiver(rabbitConfig RabbitConfig, service string, queues []string) ContextOption {
	return func(ctx *Context) {
		messageQueue, err := mq.NewReceiver(rabbitConfig.URI, queues, service)
		if err != nil {
			log.Panicf("Rabbit MQ connection error: %s", err)
		}
		ctx.MQReceiver = messageQueue
	}
}

// WithRabbitPublisher -
func WithRabbitPublisher(rabbitConfig RabbitConfig, service string) ContextOption {
	return func(ctx *Context) {
		messageQueue, err := mq.NewPublisher(rabbitConfig.URI)
		if err != nil {
			log.Panicf("Rabbit MQ connection error: %s", err)
		}
		ctx.MQPublisher = messageQueue
	}
}

// WithShare -
func WithShare(path string) ContextOption {
	return func(ctx *Context) {
		if path == "" {
			panic("Empty share path in config")
		}
		ctx.SharePath = path
	}
}

// WithConfigCopy -
func WithConfigCopy(cfg Config) ContextOption {
	return func(ctx *Context) {
		ctx.Config = cfg
	}
}

// WithTzKTServices -
func WithTzKTServices(tzktConfig map[string]TzKTConfig) ContextOption {
	return func(ctx *Context) {
		if len(tzktConfig) == 0 {
			panic("Please, set TzKT link in config")
		}
		svc := make(map[string]*tzkt.ServicesTzKT)
		for network, tzktProvider := range tzktConfig {
			svc[network] = tzkt.NewServicesTzKT(network, tzktProvider.ServicesURI, time.Second*time.Duration(tzktProvider.Timeout))
		}
		ctx.TzKTServices = svc
	}
}

// WithLoadErrorDescriptions -
func WithLoadErrorDescriptions(filePath string) ContextOption {
	return func(ctx *Context) {
		if err := cerrors.LoadErrorDescriptions(filePath); err != nil {
			panic(err)
		}
	}
}

// WithContractsInterfaces -
func WithContractsInterfaces() ContextOption {
	return func(ctx *Context) {
		result, err := kinds.Load()
		if err != nil {
			panic(err)
		}
		ctx.Interfaces = result
	}
}

// WithAliases -
func WithAliases(network string) ContextOption {
	return func(ctx *Context) {
		if ctx.DB == nil {
			panic("[WithAliases] Empty database connection")
		}
		aliases, err := ctx.DB.GetAliasesMap(network)
		if err != nil {
			panic(err)
		}
		ctx.Aliases = aliases
	}
}

// WithAWS -
func WithAWS(cfg AWSConfig) ContextOption {
	return func(ctx *Context) {
		client, err := aws.New(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Region, cfg.BucketName)
		if err != nil {
			log.Panicf("aws client init error: %s", err)
		}
		ctx.AWS = client
	}
}
