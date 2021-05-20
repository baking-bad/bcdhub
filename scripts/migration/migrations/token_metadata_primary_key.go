package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"gorm.io/gorm"
)

// TokenMetadataPrimaryKey - migration that sets composite primary key for token metadata
type TokenMetadataPrimaryKey struct{}

// Key -
func (m *TokenMetadataPrimaryKey) Key() string {
	return "token_metadata_pk"
}

// Description -
func (m *TokenMetadataPrimaryKey) Description() string {
	return "sets composite primary key for token metadata"
}

// Do - migrate function
func (m *TokenMetadataPrimaryKey) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM token_metadata tm1 USING token_metadata tm2 WHERE tm1.id < tm2.id AND tm1.network = tm2.network AND tm1.contract = tm2.contract AND tm1.token_id = tm2.token_id;`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`ALTER TABLE public.token_metadata DROP CONSTRAINT token_metadata_pkey`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`ALTER TABLE public.token_metadata ADD CONSTRAINT token_metadata_pkey PRIMARY KEY (contract,network,token_id)`).Error; err != nil {
			return err
		}
		return nil
	})
}
