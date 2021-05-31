package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

// DecimalsAmount -
type DecimalsAmount struct{}

// Key -
func (m *DecimalsAmount) Key() string {
	return "decimals_balances"
}

// Description -
func (m *DecimalsAmount) Description() string {
	return "set amount in transfers and balance in token_balances to decimals"
}

// Do - migrate function
func (m *DecimalsAmount) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(m.migrate)
}

func (m *DecimalsAmount) migrate(tx *gorm.DB) error {
	logger.Info("migrate token balances ...")
	if err := m.migrateTokenBalances(tx); err != nil {
		return err
	}
	logger.Info("migrate transfers ...")
	return m.migrateTransfers(tx)
}

func (m *DecimalsAmount) migrateTokenBalances(tx *gorm.DB) error {
	limit := int64(1000)

	var count int64
	if err := tx.Model(&tokenbalance.TokenBalance{}).Count(&count).Error; err != nil {
		return err
	}

	parts := int(count / limit)
	if int64(parts)*limit != count {
		parts += 1
	}
	bar := progressbar.NewOptions(parts, progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())

	var offset int64
	for offset < count {
		if err := bar.Add(1); err != nil {
			return err
		}
		var balances []tokenbalance.TokenBalance
		if err := tx.Model(&tokenbalance.TokenBalance{}).Offset(int(offset)).Limit(int(limit)).Find(&balances).Error; err != nil {
			return err
		}

		for i := range balances {
			balances[i].IsLedger = true
			if err := balances[i].Save(tx); err != nil {
				return err
			}
		}

		offset += int64(len(balances))
	}
	return nil
}

func (m *DecimalsAmount) migrateTransfers(tx *gorm.DB) error {
	limit := int64(1000)

	var count int64
	if err := tx.Model(&transfer.Transfer{}).Count(&count).Error; err != nil {
		return err
	}

	parts := int(count / limit)
	if int64(parts)*limit != count {
		parts += 1
	}
	bar := progressbar.NewOptions(parts, progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())

	var offset int64
	for offset < count {
		if err := bar.Add(1); err != nil {
			return err
		}
		var transfers []transfer.Transfer
		if err := tx.Model(&transfer.Transfer{}).Offset(int(offset)).Limit(int(limit)).Find(&transfers).Error; err != nil {
			return err
		}

		for i := range transfers {
			if err := transfers[i].Save(tx); err != nil {
				return err
			}
		}

		offset += int64(len(transfers))
	}
	return nil
}
