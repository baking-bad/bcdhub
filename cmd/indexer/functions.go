package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/db/account"
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/db/state"
	"github.com/aopoltorzhicky/bcdhub/internal/index"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func createRPCs(cfg config) map[string]*noderpc.NodeRPC {
	rpc := make(map[string]*noderpc.NodeRPC)
	for i := range cfg.NodeRPC {
		nodeCfg := cfg.NodeRPC[i]
		rpc[nodeCfg.Network] = noderpc.NewNodeRPC(nodeCfg.Host, nodeCfg.Network)
		rpc[nodeCfg.Network].SetTimeout(time.Second * 30)
	}
	return rpc
}

func createIndexers(cfg config) map[string]index.Indexer {
	idx := make(map[string]index.Indexer)
	if cfg.Indexer == "tzkt" {
		if cfg.TzKT.MainNet != "" {
			idx["mainnet"] = index.NewTzKT(cfg.TzKT.MainNet, time.Duration(cfg.TzKT.Timeout)*time.Second)
			log.Println("Create mainnet indexer")
		}
		if cfg.TzKT.ZeroNet != "" {
			idx["zeronet"] = index.NewTzKT(cfg.TzKT.ZeroNet, time.Duration(cfg.TzKT.Timeout)*time.Second)
			log.Println("Create zeronet indexer")
		}
		if cfg.TzKT.BabylonNet != "" {
			idx["babylonnet"] = index.NewTzKT(cfg.TzKT.BabylonNet, time.Duration(cfg.TzKT.Timeout)*time.Second)
			log.Println("Create babylonnet indexer")
		}
		if cfg.TzKT.CarthagenNet != "" {
			idx["carthagenet"] = index.NewTzKT(cfg.TzKT.CarthagenNet, time.Duration(cfg.TzKT.Timeout)*time.Second)
			log.Println("Create carthagenet indexer")
		}
		return idx
	} else if cfg.Indexer == "tzstats" {
		if cfg.TzStats.MainNet != "" {
			idx["mainnet"] = index.NewTzStats(cfg.TzStats.MainNet)
			log.Println("Create mainnet indexer")
		}
		if cfg.TzStats.ZeroNet != "" {
			idx["zeronet"] = index.NewTzStats(cfg.TzStats.ZeroNet)
			log.Println("Create zeronet indexer")
		}
		if cfg.TzStats.BabylonNet != "" {
			idx["babylonnet"] = index.NewTzStats(cfg.TzStats.BabylonNet)
			log.Println("Create babylonnet indexer")
		}
		if cfg.TzStats.CarthagenNet != "" {
			idx["carthagenet"] = index.NewTzStats(cfg.TzStats.CarthagenNet)
			log.Println("Create carthagenet indexer")
		}
		return idx
	}
	panic(fmt.Sprintf("Unknown indexer type: %s", cfg.Indexer))
}

func createContract(c index.Contract, rpc *noderpc.NodeRPC, db *gorm.DB, network string) (n contract.Contract, err error) {
	n.Level = c.Level
	n.Timestamp = c.Timestamp
	n.Balance = c.Balance

	if c.Address != "" {
		if err = db.FirstOrCreate(&n.Address, account.Account{
			Network: network,
			Address: c.Address,
		}).Error; err != nil {
			return
		}
	}
	if c.Manager != "" {
		if err = db.FirstOrCreate(&n.Manager, account.Account{
			Network: network,
			Address: c.Manager,
		}).Error; err != nil {
			return
		}
	}
	if c.Delegate != "" {
		if err = db.FirstOrCreate(&n.Delegate, account.Account{
			Network: network,
			Address: c.Delegate,
		}).Error; err != nil {
			return
		}
	}

	n.Network = network

	scriptMap, err := rpc.GetContractScript(c.Address)
	if err != nil {
		return n, err
	}
	b, err := json.Marshal(scriptMap)
	if err != nil {
		return n, err
	}
	n.Script = postgres.Jsonb{b}
	err = contract.Add(db, &n)
	return
}

func syncIndexer(rpc *noderpc.NodeRPC, indexer index.Indexer, db *gorm.DB, messageQueue *mq.MQ, network string) error {
	log.Printf("-----------%s-----------", strings.ToUpper(network))
	head, err := rpc.GetHead()
	if err != nil {
		return err
	}
	log.Printf("Current node state: %d", head.Level)

	// Get current DB state
	s, err := state.Current(db, network)
	if err != nil {
		return err
	}

	log.Printf("Current state: %d", s.Level)

	if head.Level > s.Level {
		contracts, err := indexer.GetContracts(s.Level)
		if err != nil {
			return err
		}
		log.Printf("New contracts: %d", len(contracts))

		for _, c := range contracts {
			n, err := createContract(c, rpc, db, network)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Printf("[%s] Contract created", n.Address.Address)
			if s.Level < n.Level {
				s.Level = n.Level
				s.Timestamp = n.Timestamp
				if err = db.Save(&s).Error; err != nil {
					fmt.Println(err)
				}
			}

			if err := messageQueue.SendStruct(mq.ChannelContract, n.ID); err != nil {
				fmt.Println(err)
			}
		}
		if head.Level > s.Level {
			s.Level = head.Level
			s.Timestamp = head.Timestamp
			if err = db.Save(&s).Error; err != nil {
				log.Println(err)
				return err
			}
		}
		log.Print("Synced")
	}
	return nil
}

func sync(rpcs map[string]*noderpc.NodeRPC, indexers map[string]index.Indexer, db *gorm.DB, messageQueue *mq.MQ) error {
	for network, indexer := range indexers {
		rpc, ok := rpcs[network]
		if !ok {
			log.Printf("Unknown RPC network: %s", network)
			continue
		}

		if err := syncIndexer(rpc, indexer, db, messageQueue, network); err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}
