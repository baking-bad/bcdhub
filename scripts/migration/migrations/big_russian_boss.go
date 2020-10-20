package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
)

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
	return nil
}

func (m *BigRussianBoss) fillTZIP(ctx *config.Context) error {
	yes, err := ask("Do you want to fill TZIP data from repository? (yes/no)")
	if err != nil {
		return err
	}
	if yes == "yes" {
		migration := FillTZIP{}
		if err := migration.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *BigRussianBoss) createTZIP(ctx *config.Context) error {
	yes, err := ask("Do you want to create missing TZIP data? (yes/no)")
	if err != nil {
		return err
	}
	if yes == "yes" {
		migration := CreateTZIP{}
		if err := migration.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}
