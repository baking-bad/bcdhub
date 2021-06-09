package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"gorm.io/gorm"
)

// FixZeroID -
type FixZeroID struct{}

// Key -
func (m *FixZeroID) Key() string {
	return "fix_zero_id"
}

// Description -
func (m *FixZeroID) Description() string {
	return "fix for zero identity"
}

// Do - migrate function
func (m *FixZeroID) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		logger.Info("setting new ids for token metadata...")

		var id int64
		limit := 1000
		offset := 0

		var end bool
		for !end {
			var tokens []tokenmetadata.TokenMetadata
			if err := tx.Model(&tokenmetadata.TokenMetadata{}).Offset(offset).Limit(limit).Order("timestamp asc").Find(&tokens).Error; err != nil {
				return err
			}

			for i := range tokens {
				id++
				if err := tx.Model(&tokens[i]).Update("id", id).Error; err != nil {
					return err
				}
			}

			offset += len(tokens)
			end = len(tokens) < limit
		}

		logger.Info("creating sequence...")
		return tx.Exec(`
			CREATE SEQUENCE token_metadata_id_seq;
			ALTER TABLE token_metadata ALTER COLUMN id SET DEFAULT nextval('token_metadata_id_seq');
			ALTER SEQUENCE token_metadata_id_seq OWNED BY token_metadata.id;
			SELECT setval('token_metadata_id_seq', COALESCE(max(id), 0)) FROM token_metadata;
		`).Error
	})
}
