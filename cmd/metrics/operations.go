package main

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func getOperation(ids []string) error {
	operations := make([]models.Operation, 0)
	if err := ctx.ES.GetByIDs(&operations, ids...); err != nil {
		return errors.Errorf("[getOperation] Find operation error for IDs %v: %s", ids, err)
	}

	h := metrics.New(ctx.ES, ctx.DB)
	updated := make([]elastic.Model, 0)
	for i := range operations {
		if err := parseOperation(h, operations[i]); err != nil {
			return errors.Errorf("[getOperation] Compute error message: %s", err)
		}

		updated = append(updated, &operations[i])
	}

	if err := ctx.ES.BulkUpdate(updated); err != nil {
		return err
	}

	logger.Info("%d operations are processed", len(operations))

	return getOperationsContracts(h, operations)
}

func parseOperation(h *metrics.Handler, operation models.Operation) error {
	if _, err := h.SetOperationAliases(&operation); err != nil {
		return err
	}
	h.SetOperationStrings(&operation)

	if helpers.IsContract(operation.Destination) || operation.IsOrigination() {
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

func getOperationsContracts(h *metrics.Handler, operations []models.Operation) error {
	addresses := make([]elastic.Address, 0)
	addressesMap := make(map[elastic.Address]*stats)
	for i := range operations {
		if helpers.IsContract(operations[i].Destination) {
			dest := elastic.Address{
				Address: operations[i].Destination,
				Network: operations[i].Network,
			}
			if _, ok := addressesMap[dest]; !ok {
				addressesMap[dest] = new(stats)
				addresses = append(addresses, dest)
			}
			addressesMap[dest].update(operations[i].Timestamp)
		}
		if helpers.IsContract(operations[i].Source) {
			src := elastic.Address{
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

	contracts, err := ctx.ES.GetContractsByAddresses(addresses)
	if err != nil {
		return err
	}

	updated := make([]models.Contract, 0)
	contractsMap := make(map[elastic.Address]models.Contract)
	for i := range contracts {
		addr := elastic.Address{
			Address: contracts[i].Address,
			Network: contracts[i].Network,
		}
		if s, ok := addressesMap[addr]; ok {
			if !s.isZero() {
				contracts[i].TxCount += s.Count
				contracts[i].LastAction = s.LastAction
				updated = append(updated, contracts[i])
			}
			contractsMap[addr] = contracts[i]
		}
	}

	if err := ctx.ES.BulkUpdateField(updated, "TxCount", "LastAction"); err != nil {
		return err
	}

	for i := range operations {
		if !operations[i].IsTransaction() || !operations[i].IsCall() {
			continue
		}
		addr := elastic.Address{
			Address: operations[i].Destination,
			Network: operations[i].Network,
		}
		if cntr, ok := contractsMap[addr]; ok {
			rpc, err := ctx.GetRPC(cntr.Network)
			if err != nil {
				return err
			}
			if err := h.FixTokenMetadata(rpc, ctx.SharePath, &cntr, &operations[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
