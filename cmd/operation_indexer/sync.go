package main

import (
	"fmt"
	"sync"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func getContracts(es *elastic.Elastic, network string) (map[string]struct{}, map[string]struct{}, error) {
	addresses, err := es.GetContracts(map[string]interface{}{
		"network": network,
	})
	if err != nil {
		return nil, nil, err
	}
	res := make(map[string]struct{})
	spendable := make(map[string]struct{})
	for _, a := range addresses {
		res[a.Address] = struct{}{}
		if helpers.StringInArray(consts.SpendableTag, a.Tags) {
			spendable[a.Address] = struct{}{}
		}
	}

	return res, spendable, nil
}

func updateState(rpc *noderpc.NodeRPC, es *elastic.Elastic, currentLevel int64, s *models.State) error {
	if s.Level >= currentLevel {
		return nil
	}
	s.Level = currentLevel

	t, err := rpc.GetLevelTime(int(currentLevel))
	if err != nil {
		return err
	}
	s.Timestamp = t

	if _, err = es.UpdateDoc(elastic.DocStates, s.ID, *s); err != nil {
		return err
	}
	return nil
}

func saveOperations(ctx *Context, ops []models.Operation, s *models.State) error {
	if len(ops) == 0 {
		return nil
	}

	for j := range ops {
		ops[j].Timestamp = s.Timestamp
		if _, err := ctx.ES.AddDocumentWithID(ops[j], elastic.DocOperations, ops[j].ID); err != nil {
			return err
		}

		if err := ctx.MQ.Send(mq.ChannelNew, mq.QueueOperations, ops[j].ID); err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func syncNetwork(ctx *Context, network string, wg *sync.WaitGroup) {
	defer wg.Done()

	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "network", network)

	rpc, err := ctx.GetRPC(network)
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		helpers.LocalCatchErrorSentry(localSentry, err)
		return
	}

	indexer, err := ctx.GetIndexer(network)
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		helpers.LocalCatchErrorSentry(localSentry, err)
		return
	}

	cs, err := ctx.ES.CurrentState(network, models.StateContract)
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		helpers.LocalCatchErrorSentry(localSentry, err)
		return
	}
	logger.Info("[%s] Current contract indexer state: %d", network, cs.Level)

	// Get current DB state
	s, ok := ctx.States[network]
	if !ok {
		logger.Errorf("Unknown network: %s", network)
		helpers.LocalCatchErrorSentry(localSentry, fmt.Errorf("Unknown network: %s", network))
		return
	}
	logger.Info("[%s] Current state: %d", network, s.Level)

	if cs.Level > s.Level {
		addresses, spendable, err := getContracts(ctx.ES, network)
		if err != nil {
			logger.Errorf("[%s] %s", network, err.Error())
			helpers.LocalCatchErrorSentry(localSentry, err)
			return
		}

		levels, err := indexer.GetContractOperationBlocks(int(s.Level), int(cs.Level), addresses, spendable)
		if err != nil {
			logger.Errorf("[%s] %s", network, err.Error())
			helpers.LocalCatchErrorSentry(localSentry, err)
			return
		}

		if len(levels) == 0 {
			return
		}

		logger.Info("[%s] Found %d contracts", network, len(addresses))
		logger.Info("[%s] Found %d new levels", network, len(levels))

		for _, l := range levels {
			ops, err := getOperations(rpc, ctx.ES, l, network, addresses)
			if err != nil {
				logger.Errorf("[%s %d] %s", network, l, err.Error())
				helpers.LocalCatchErrorSentry(localSentry, fmt.Errorf("[%d] %s", l, err.Error()))
				return
			}

			logger.Info("[%s] %d/%d Found %d operations", network, l, cs.Level, len(ops))
			if err := saveOperations(ctx, ops, s); err != nil {
				logger.Errorf("[%s %d] %s", network, l, err.Error())
				helpers.LocalCatchErrorSentry(localSentry, fmt.Errorf("[%d] %s", l, err.Error()))
				return
			}

			if err := updateState(rpc, ctx.ES, l, s); err != nil {
				logger.Errorf("[%s %d] %s", network, l, err.Error())
				helpers.LocalCatchErrorSentry(localSentry, fmt.Errorf("[%d] %s", l, err.Error()))
				return
			}
		}
	}

	logger.Success("[%s] Synced", network)
}

func process(ctx *Context) error {
	var wg sync.WaitGroup
	for network := range ctx.Indexers {
		wg.Add(1)
		go syncNetwork(ctx, network, &wg)
	}
	wg.Wait()

	return nil
}
