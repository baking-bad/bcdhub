package protocols

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/migrations"
	"github.com/baking-bad/bcdhub/internal/postgres/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Migration struct {
	network types.Network
	ctx     *config.Context
}

func NewMigration(
	network types.Network,
	ctx *config.Context,
) Migration {
	return Migration{
		network: network,
		ctx:     ctx,
	}
}

func (m Migration) Do(
	ctx context.Context,
	tx models.Transaction,
	currentProtocol protocol.Protocol,
	head noderpc.Header,
) (protocol.Protocol, error) {
	if err := m.finilizeProtocol(ctx, tx, currentProtocol, head); err != nil {
		return currentProtocol, errors.Wrapf(err, "finalization of %s", currentProtocol.Hash)
	}

	newProto, err := m.newProtocol(ctx, tx, head)
	if err != nil {
		return currentProtocol, errors.Wrapf(err, "creating new protocol: %s", head.Hash)
	}

	if err := m.contractMigrations(ctx, tx, head, currentProtocol, newProto); err != nil {
		return currentProtocol, errors.Wrapf(err, "contracts migration to new protocol: %s", head.Protocol)
	}

	return newProto, nil
}

func (m Migration) finilizeProtocol(
	ctx context.Context,
	tx models.Transaction,
	proto protocol.Protocol,
	head noderpc.Header,
) error {
	if proto.EndLevel > 0 || head.Level == 1 {
		return nil
	}
	log.Info().
		Str("network", m.network.String()).
		Msgf("Finalizing the previous protocol: %s", proto.Alias)

	proto.EndLevel = head.Level - 1
	return tx.Protocol(ctx, &proto)
}

func (m Migration) newProtocol(
	ctx context.Context,
	tx models.Transaction,
	head noderpc.Header,
) (protocol.Protocol, error) {
	log.Info().
		Str("network", m.network.String()).
		Msgf("Creating new protocol %s starting at %d", head.Protocol, head.Level)

	newProtocol, err := Create(ctx, m.ctx.RPC, head)
	if err != nil {
		return newProtocol, err
	}
	err = tx.Protocol(ctx, &newProtocol)
	return newProtocol, err
}

func (m Migration) contractMigrations(
	ctx context.Context,
	tx models.Transaction,
	head noderpc.Header,
	currentProtocol protocol.Protocol,
	newProtocol protocol.Protocol,
) error {
	if head.Level == 1 {
		return m.vestingMigration(ctx, tx, head, currentProtocol)
	}

	if currentProtocol.SymLink == "" {
		return errors.Errorf("[%s] Protocol should be initialized", m.network)
	}

	if newProtocol.SymLink != currentProtocol.SymLink {
		return m.standartMigration(ctx, currentProtocol, newProtocol, head, tx)
	}

	log.Info().
		Str("network", m.network.String()).
		Msgf("Same symlink %s for %s / %s", newProtocol.SymLink, currentProtocol.Alias, newProtocol.Alias)

	return nil
}

func (m Migration) vestingMigration(ctx context.Context, _ models.Transaction, head noderpc.Header, currentProtocol protocol.Protocol) error {
	addresses, err := m.ctx.RPC.GetContractsByBlock(ctx, head.Level)
	if err != nil {
		return err
	}

	specific, err := Get(m.ctx, currentProtocol.Hash)
	if err != nil {
		return err
	}

	p := migrations.NewVestingParser(m.ctx, specific.ContractParser, currentProtocol)
	store := store.NewStore(m.ctx.StorageDB.DB, m.ctx.Stats)

	for _, address := range addresses {
		if !bcd.IsContract(address) {
			continue
		}

		data, err := m.ctx.RPC.GetContractData(ctx, address, head.Level)
		if err != nil {
			return err
		}

		if err := p.Parse(ctx, data, head, address, store); err != nil {
			return err
		}
	}

	return store.Save(ctx)
}

func (m Migration) standartMigration(ctx context.Context, currentProtocol, newProtocol protocol.Protocol, head noderpc.Header, tx models.Transaction) error {
	log.Info().Str("network", m.network.String()).Msg("Try to find migrations...")

	contracts, err := m.ctx.Contracts.AllExceptDelegators(ctx)
	if err != nil {
		return err
	}
	log.Info().Str("network", m.network.String()).Msgf("Now %2d contracts are indexed", len(contracts))

	specific, err := Get(m.ctx, newProtocol.Hash)
	if err != nil {
		return err
	}

	for i := range contracts {
		if !specific.MigrationParser.IsMigratable(contracts[i].Account.Address) && newProtocol.SymLink == bcd.SymLinkJakarta {
			if err := tx.JakartaVesting(ctx, &contracts[i]); err != nil {
				return errors.Wrap(err, "jakarta vesting migration error")
			}
			continue
		}

		log.Info().Str("network", m.network.String()).Msgf("Migrate %s...", contracts[i].Account.Address)
		script, err := m.ctx.RPC.GetScriptJSON(ctx, contracts[i].Account.Address, newProtocol.StartLevel)
		if err != nil {
			return err
		}

		if err := specific.MigrationParser.Parse(
			ctx, script, &contracts[i], currentProtocol, newProtocol, head.Timestamp, tx,
		); err != nil {
			return err
		}

		switch newProtocol.SymLink {
		case bcd.SymLinkBabylon:
			err = tx.BabylonUpdateNonDelegator(ctx, &contracts[i])
		case bcd.SymLinkJakarta:
			err = tx.JakartaUpdateNonDelegator(ctx, &contracts[i])
		}

		if err != nil {
			return errors.Wrapf(err, "migration of contract error: %s", contracts[i].Account.Address)
		}

	}

	// only delegator contracts
	switch newProtocol.SymLink {
	case bcd.SymLinkBabylon:
		err = tx.ToBabylon(ctx)
	case bcd.SymLinkJakarta:
		err = tx.ToJakarta(ctx)
	}
	return err
}
