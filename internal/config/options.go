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
		ctx.ES = elastic.WaitNew(esConfig.URI, esConfig.Timeout)
	}
}

// WithDatabase -
func WithDatabase(dbConfig DatabaseConfig) ContextOption {
	return func(ctx *Context) {
		ctx.DB = database.WaitNew(dbConfig.ConnString, dbConfig.Timeout)
	}
}

// WithRabbit -
func WithRabbit(rabbitConfig RabbitConfig, service string, mqConfig MQConfig) ContextOption {
	return func(ctx *Context) {
		mqueues := make([]mq.Queue, 0)
		for name, params := range mqConfig.Queues {
			mqueues = append(mqueues, mq.Queue{
				Name:       name,
				AutoDelete: params.AutoDeleted,
				Durable:    !params.NonDurable,
			})
		}

		ctx.MQ = mq.WaitNew(rabbitConfig.URI, service, mqConfig.NeedPublisher, rabbitConfig.Timeout, mqueues...)
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
			return
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
		if ctx.ES == nil {
			panic("[WithAliases] Empty database connection")
		}
		aliases, err := ctx.ES.GetAliasesMap(network)
		if err != nil {
			if elastic.IsRecordNotFound(err) {
				return
			}
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
