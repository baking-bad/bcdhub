package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

const yes = "yes"

// BigRussianBoss -
type BigRussianBoss struct{}

// Key -
func (m *BigRussianBoss) Key() string {
	return "big_russian_boss"
}

// Description -
func (m *BigRussianBoss) Description() string {
	return "Script for filling missing or manual metadata"
}

// Do - migrate function
func (m *BigRussianBoss) Do(ctx *config.Context) error {
	if err := m.fillTZIP(ctx); err != nil {
		return err
	}
	if err := m.createTZIP(ctx); err != nil {
		return err
	}
	if err := m.fillAliases(ctx); err != nil {
		return err
	}
	if err := m.eventsAndTokenBalances(ctx); err != nil {
		return err
	}
	return nil
}

func (m *BigRussianBoss) fillTZIP(ctx *config.Context) error {
	answer, err := ask("Do you want to fill TZIP data from repository? (yes/no)")
	if err != nil {
		return err
	}
	if answer == yes {
		if err := new(FillTZIP).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *BigRussianBoss) createTZIP(ctx *config.Context) error {
	answer, err := ask("Do you want to create missing TZIP data? (yes/no)")
	if err != nil {
		return err
	}
	if answer == yes {
		if err := new(CreateTZIP).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *BigRussianBoss) fillAliases(ctx *config.Context) error {
	answer, err := ask("Do you want to fill aliases? (yes/no)")
	if err != nil {
		return err
	}
	if answer == yes {
		if err := new(GetAliases).Do(ctx); err != nil {
			return err
		}

		if err := new(SetAliases).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *BigRussianBoss) eventsAndTokenBalances(ctx *config.Context) error {
	answer, err := ask("Do you want to execute all events (initial_storage, extended_storage, parameter_events) and recalculate token balances? (yes/no)")
	if err != nil {
		return err
	}
	if answer != yes {
		return nil
	}

	logger.Info("executing all initial storages")
	initStorageEvents := new(InitialStorageEvents)
	if err := initStorageEvents.Do(ctx); err != nil {
		return err
	}

	logger.Info("executing all extended storages")
	extStorageEvents := new(ExtendedStorageEvents)
	if err := extStorageEvents.Do(ctx); err != nil {
		return err
	}

	logger.Info("executing all parameter events")
	parameterEvents := new(ParameterEvents)
	if err := parameterEvents.Do(ctx); err != nil {
		return err
	}

	uniqueContracts := make(map[string]string)
	for _, contracts := range []map[string]string{
		initStorageEvents.AffectedContracts(),
		extStorageEvents.AffectedContracts(),
		parameterEvents.AffectedContracts(),
	} {
		for address, network := range contracts {
			uniqueContracts[address] = network
		}
	}

	logger.Info("Found %v affected contracts. Starting token balance recalculation", len(uniqueContracts))
	if err := new(TokenBalanceRecalc).DoBatch(ctx, uniqueContracts); err != nil {
		return err
	}

	return nil
}
