package dapp

import (
	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
)

// DApp -
type DApp struct {
	// nolint
	tableName struct{} `pg:"dapps"`

	ID               int64          `json:"-"`
	Name             string         `json:"name" pg:",unique"`
	ShortDescription string         `json:"short_description"`
	FullDescription  string         `json:"full_description"`
	WebSite          string         `json:"web_site"`
	Slug             string         `json:"slug,omitempty"`
	Authors          pq.StringArray `json:"authors" pg:",type:text[]"`
	SocialLinks      pq.StringArray `json:"social_links" pg:",type:text[]"`
	Interfaces       pq.StringArray `json:"interfaces" pg:",type:text[]"`
	Categories       pq.StringArray `json:"categories" pg:",type:text[]"`
	Contracts        DAppContracts  `json:"contracts" pg:",type:jsonb"`
	Order            int64          `json:"order"`
	Soon             bool           `json:"soon" pg:",use_zero"`

	Pictures  Pictures  `json:"pictures,omitempty" pg:",type:jsonb"`
	DexTokens DexTokens `json:"dex_tokens,omitempty" pg:",type:jsonb"`
}

// GetID -
func (d *DApp) GetID() int64 {
	return d.ID
}

// GetIndex -
func (d *DApp) GetIndex() string {
	return "dapps"
}

// Save -
func (d *DApp) Save(tx pg.DBI) error {
	_, err := tx.Model(d).
		OnConflict("(name) DO UPDATE").
		Set(`
		"short_description" = excluded.short_description,
		"full_description" = excluded.full_description,
		"web_site" = excluded.web_site,
		"slug" = excluded.slug,
		"authors" = excluded.authors,
		"social_links" = excluded.social_links,
		"interfaces" = excluded.interfaces,
		"categories" = excluded.categories,
		"contracts" = excluded.contracts,
		"order" = excluded.order,
		"soon" = excluded.soon,
		"pictures" = excluded.pictures,
		"dex_tokens" = excluded.dex_tokens
	`).
		Returning("id").Insert()
	return err
}

// LogFields -
func (d *DApp) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"name": d.Name,
	}
}

// Picture -
type Picture struct {
	Link string `json:"link"`
	Type string `json:"type"`
}

// Pictures -
type Pictures []Picture

// DexToken -
type DexToken struct {
	TokenID  uint64 `json:"token_id"`
	Contract string `json:"contract"`
}

// DexTokens -
type DexTokens []DexToken

// DAppContract -
type DAppContract struct {
	Address    string   `json:"address"`
	Entrypoint []string `json:"dex_volume_entrypoints,omitempty"`
	WithTokens bool     `json:"with_tokens,omitempty"`
}

// DAppContracts -
type DAppContracts []DAppContract
