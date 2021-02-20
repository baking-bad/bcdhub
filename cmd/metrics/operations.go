package main

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func getOperation(ids []string) error {
	operations := make([]operation.Operation, 0)
	if err := ctx.Storage.GetByIDs(&operations, ids...); err != nil {
		return errors.Errorf("[getOperation] Find operation error for IDs %v: %s", ids, err)
	}

	updated := make([]models.Model, 0)
	for i := range operations {
		if err := parseOperation(operations[i]); err != nil {
			return errors.Errorf("[getOperation] Compute error message: %s", err)
		}

		updated = append(updated, &operations[i])
	}

	if err := ctx.Storage.BulkUpdate(updated); err != nil {
		return err
	}

	logger.Info("%d operations are processed", len(operations))

	return getOperationsContracts(operations)
}

func parseOperation(operation operation.Operation) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.Schema, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	aliases, err := getAliases(operation.Network)
	if err != nil {
		return err
	}

	if _, err := h.SetOperationAliases(&operation, aliases); err != nil {
		return err
	}
	h.SetOperationStrings(&operation)

	if bcd.IsContract(operation.Destination) || operation.IsOrigination() {
		if err := h.SendSentryNotifications(operation); err != nil {
			return err
		}
	}
	return nil
}

type stats struct {
	Count      int64
	LastAction time.Time
}

func (s *stats) update(ts time.Time) {
	s.Count++
	s.LastAction = ts
}

func (s *stats) isZero() bool {
	return s.Count == 0 && s.LastAction.IsZero()
}

func getOperationsContracts(operations []operation.Operation) error {
	addresses := make([]contract.Address, 0)
	addressesMap := make(map[contract.Address]*stats)
	for i := range operations {
		if bcd.IsContract(operations[i].Destination) {
			dest := contract.Address{
				Address: operations[i].Destination,
				Network: operations[i].Network,
			}
			if _, ok := addressesMap[dest]; !ok {
				addressesMap[dest] = new(stats)
				addresses = append(addresses, dest)
			}
			addressesMap[dest].update(operations[i].Timestamp)
		}
		if bcd.IsContract(operations[i].Source) {
			src := contract.Address{
				Address: operations[i].Source,
				Network: operations[i].Network,
			}
			if _, ok := addressesMap[src]; !ok {
				addressesMap[src] = new(stats)
				addresses = append(addresses, src)
			}
			addressesMap[src].update(operations[i].Timestamp)
		}
	}

	contracts, err := ctx.Contracts.GetByAddresses(addresses)
	if err != nil {
		return err
	}

	updated := make([]contract.Contract, 0)
	for i := range contracts {
		addr := contract.Address{
			Address: contracts[i].Address,
			Network: contracts[i].Network,
		}
		if s, ok := addressesMap[addr]; ok {
			if !s.isZero() {
				contracts[i].TxCount += s.Count
				contracts[i].LastAction = s.LastAction
				updated = append(updated, contracts[i])
			}
		}
	}
	return ctx.Contracts.UpdateField(updated, "TxCount", "LastAction")
}
