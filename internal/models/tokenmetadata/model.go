package tokenmetadata

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
)

// TokenMetadata -
type TokenMetadata struct {
	// nolint
	tableName struct{} `pg:"token_metadata"`

	ID                 int64                  `pg:",notnull" json:"-"`
	Network            types.Network          `pg:",type:SMALLINT,unique:token_metadata,use_zero" json:"network"`
	Contract           string                 `pg:",unique:token_metadata" json:"contract"`
	TokenID            uint64                 `pg:",type:numeric(50,0),unique:token_metadata,use_zero" json:"token_id"`
	Level              int64                  `pg:",use_zero" json:"level"`
	Timestamp          time.Time              `json:"timestamp"`
	Symbol             string                 `json:"symbol"`
	Name               string                 `json:"name"`
	Decimals           *int64                 `json:"decimals,omitempty"`
	Description        string                 `json:"description,omitempty"`
	ArtifactURI        string                 `json:"artifact_uri,omitempty"`
	DisplayURI         string                 `json:"display_uri,omitempty"`
	ThumbnailURI       string                 `json:"thumbnail_uri,omitempty"`
	ExternalURI        string                 `json:"external_uri,omitempty"`
	IsTransferable     bool                   `pg:",default:true" json:"is_transferable"`
	IsBooleanAmount    bool                   `pg:",use_zero" json:"is_boolean_amount"`
	ShouldPreferSymbol bool                   `pg:",use_zero" json:"should_prefer_symbol"`
	Tags               pq.StringArray         `pg:",type:text[]" json:"tags,omitempty"`
	Creators           pq.StringArray         `pg:",type:text[]" json:"creators,omitempty"`
	Formats            types.Bytes            `pg:",type:bytea" json:"formats,omitempty"`
	Extras             map[string]interface{} `pg:",type:jsonb" json:"extras,omitempty"`
}

// ByName - TokenMetadata sorting filter by Name field
type ByName []TokenMetadata

func (tm ByName) Len() int      { return len(tm) }
func (tm ByName) Swap(i, j int) { tm[i], tm[j] = tm[j], tm[i] }
func (tm ByName) Less(i, j int) bool {
	if tm[i].Name == "" {
		return false
	} else if tm[j].Name == "" {
		return true
	}

	return tm[i].Name < tm[j].Name
}

// ByTokenID - TokenMetadata sorting filter by TokenID field
type ByTokenID []TokenMetadata

func (tm ByTokenID) Len() int           { return len(tm) }
func (tm ByTokenID) Swap(i, j int)      { tm[i], tm[j] = tm[j], tm[i] }
func (tm ByTokenID) Less(i, j int) bool { return tm[i].TokenID < tm[j].TokenID }

// GetID -
func (t *TokenMetadata) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *TokenMetadata) GetIndex() string {
	return "token_metadata"
}

// Save -
func (t *TokenMetadata) Save(tx pg.DBI) error {
	_, err := tx.Model(t).
		OnConflict("(network, contract, token_id) DO UPDATE").
		Set(`
			symbol               = excluded.symbol,
			name                 = excluded.name,
			decimals             = excluded.decimals,
			description          = excluded.description,
			artifact_uri         = excluded.artifact_uri,
			display_uri          = excluded.display_uri,
			thumbnail_uri        = excluded.thumbnail_uri,
			external_uri         = excluded.external_uri,
			is_transferable      = excluded.is_transferable,
			is_boolean_amount    = excluded.is_boolean_amount,
			should_prefer_symbol = excluded.should_prefer_symbol,
			tags                 = excluded.tags,
			creators             = excluded.creators,
			formats              = excluded.formats,
			extras               = excluded.extras
	`).Returning("id").Insert()
	return err
}

// LogFields -
func (t *TokenMetadata) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  t.Network.String(),
		"contract": t.Contract,
		"token_id": t.TokenID,
	}
}
