package operations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Migration -
type Migration struct {
	contracts contract.Repository
}

// NewMigration -
func NewMigration(contracts contract.Repository) Migration {
	return Migration{contracts}
}

// Parse -
func (m Migration) Parse(ctx context.Context, data noderpc.Operation, operation *operation.Operation, protocol string, store parsers.Store) error {
	switch protocol {
	case
		"ProtoGenesisGenesisGenesisGenesisGenesisGenesk612im",
		"PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i",
		"PtBMwNZT94N7gXKw4i273CKcSaBrrBnqnt3RATExNKr9KNX2USV",
		"ProtoDemoNoopsDemoNoopsDemoNoopsDemoNoopsDemo6XBoYp",
		"PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex",
		"Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P",
		"PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt",
		"PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP",
		"Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd",
		"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY",
		"PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS",
		"PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU",
		"PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
		"PryLyZ8A11FXDr1tRE9zQ7Di6Y8zX48RfFCFpkjC8Pt9yCBLhtN",
		"PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
		"PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq",
		"PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA":
		return m.fromBigMapDiffs(ctx, data, operation, store)
	case
		"ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK",
		"PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx",
		"PtHangzHogokSuiMHemCuowEavgYTP8J5qQ9fQS793MHYFpCY3r",
		"PsiThaCaT47Zboaw71QWScM8sXeMM7bbQFncK9FLqYc6EKdpjVP",
		"Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A",
		"PrrUA9dCzbqBzugjQyw65HLHKjhH3HMFSLLHLZjj5rkmkG13Fej",
		"PsrsRVg1Gycjn5LvMtoYSQah1znvYmGp8bHLxwYLBZaYFf2CEkV",
		"PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i",
		"PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV",
		"PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY",
		"PtKathmankSpLLDALzWw7CGD2j2MtyveTwboEYokqUCP4a1LxMg",
		"PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW",
		"PtMumbaiiFFEGbew1rRjzSPyzRbA51Tm3RVZL5suHPxSZYDhCEc",
		"PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1",
		"PtNairobiyssHuh87hEhfVBGCVrK3WnS8Z2FT4ymB5tAa4r1nQf",
		"ProxfordSW2S7fvchT1Zgj2avb5UES194neRyYVXoaDGvF9egt8",
		"ProxfordYmVfjWnRcgjWH36fW6PArwqykTFzotUxRs6gmTcZDuH",
		"PtParisA6ruu136piHaBC7cQLDP87JEqtczJWP2pLa5QCELGBH5",
		"PtParisBQscdCm6Cfow6ndeU6wKJyA3aV1j4D3gQBQMsTQyJCrz",
		"PtParisBxoLz5gzMmn3d9WBQNoPSZakgnkMC2VNuQ3KXfUtUQeZ",
		"PsParisCZo7KAh1Z1smVd9ZMZ1HHn5gkzbM94V3PLCpknFWhUAi",
		"PsQuebecnLByd3JwTiGadoG4nGWi3HYiLXUjkibeFV8dCFeVMUg",
		"PsRiotumaAMotcRoDWW1bysEhQy2n1M5fy8JgRp8jjRfHGmfeA7",
		"PtSeouLouXkxhg39oWzjxDWaCydNfR3RxCUrNe4Q9Ro8BTehcbh":
		return m.fromLazyStorageDiff(ctx, data, operation, store)
	default:
		return errors.Errorf("unknown protocol for migration parser: %s", protocol)
	}
}

func (m Migration) fromLazyStorageDiff(ctx context.Context, data noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	var lsd []noderpc.LazyStorageDiff
	switch {
	case data.Result != nil && data.Result.LazyStorageDiff != nil:
		lsd = data.Result.LazyStorageDiff
	case data.Metadata != nil && data.Metadata.OperationResult != nil && data.Metadata.OperationResult.LazyStorageDiff != nil:
		lsd = data.Metadata.OperationResult.LazyStorageDiff
	default:
		return nil
	}

	for i := range lsd {
		if lsd[i].Kind != types.LazyStorageDiffBigMap || lsd[i].Diff == nil || lsd[i].Diff.BigMap == nil {
			continue
		}

		if lsd[i].Diff.BigMap.Action != types.BigMapActionStringUpdate {
			continue
		}

		for j := range lsd[i].Diff.BigMap.Updates {
			migration, err := m.createMigration(ctx, lsd[i].Diff.BigMap.Updates[j].Value, operation)
			if err != nil {
				return err
			}
			if migration != nil {
				operation.Destination.MigrationsCount += 1
				store.AddMigrations(migration)
				log.Info().Fields(migration.LogFields()).Msg("Migration detected")
			}
		}
	}
	return nil
}

func (m Migration) fromBigMapDiffs(ctx context.Context, data noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	var bmd []noderpc.BigMapDiff
	switch {
	case data.Result != nil && data.Result.BigMapDiffs != nil:
		bmd = data.Result.BigMapDiffs
	case data.Metadata != nil && data.Metadata.OperationResult != nil && data.Metadata.OperationResult.BigMapDiffs != nil:
		bmd = data.Metadata.OperationResult.BigMapDiffs
	default:
		return nil
	}

	for i := range bmd {
		if bmd[i].Action != types.BigMapActionStringUpdate {
			continue
		}

		migration, err := m.createMigration(ctx, bmd[i].Value, operation)
		if err != nil {
			return err
		}
		if migration != nil {
			operation.Destination.MigrationsCount += 1
			store.AddMigrations(migration)
			log.Info().Fields(migration.LogFields()).Msg("Migration detected")
		}
	}
	return nil
}

func (m Migration) createMigration(ctx context.Context, value []byte, operation *operation.Operation) (*migration.Migration, error) {
	if len(value) == 0 {
		return nil, nil
	}
	var tree ast.UntypedAST
	if err := json.Unmarshal(value, &tree); err != nil {
		return nil, err
	}

	if len(tree) == 0 {
		return nil, nil
	}

	if !tree[0].IsLambda() {
		return nil, nil
	}

	c, err := m.contracts.Get(ctx, operation.Destination.Address)
	if err != nil {
		return nil, err
	}
	return &migration.Migration{
		ContractID: c.ID,
		Contract:   c,
		Level:      operation.Level,
		ProtocolID: operation.ProtocolID,
		Timestamp:  operation.Timestamp,
		Hash:       operation.Hash,
		Kind:       types.MigrationKindLambda,
	}, nil
}
