package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// GetTokenMetadata -
func (storage *Storage) GetTokenMetadata(ctx tzip.GetTokenMetadataContext) (tokens []tzip.TokenMetadata, err error) {
	tzips := make([]tzip.TZIP, 0)

	query := storage.db.Query(models.DocTZIP)
	buildGetTokenMetadataContext(ctx, query)
	if err = storage.db.GetAllByQuery(query, &tzips); err != nil {
		return
	}
	if len(tzips) == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	tokens = make([]tzip.TokenMetadata, 0)
	for k := range tzips {
		if tzips[k].Tokens == nil {
			continue
		}

		for i := range tzips[k].Tokens.Static {
			tokens = append(tokens, tzip.TokenMetadata{
				Address:         tzips[k].Address,
				Network:         tzips[k].Network,
				Level:           tzips[k].Level,
				RegistryAddress: tzips[k].Tokens.Static[i].RegistryAddress,
				Symbol:          tzips[k].Tokens.Static[i].Symbol,
				Name:            tzips[k].Tokens.Static[i].Name,
				Decimals:        tzips[k].Tokens.Static[i].Decimals,
				TokenID:         tzips[k].Tokens.Static[i].TokenID,
				Extras:          tzips[k].Tokens.Static[i].Extras,
			})
		}
	}
	return
}

// Get -
func (storage *Storage) Get(network, address string) (t tzip.TZIP, err error) {
	t.Address = address
	t.Network = network
	err = storage.db.GetByID(&t)
	return
}

// GetDApps -
func (storage *Storage) GetDApps() (tokens []tzip.DApp, err error) {
	query := storage.db.Query(models.DocTZIP).
		Not().
		Where("dapps", reindexer.EMPTY, 0).
		Sort("dapps.order", false)
	err = storage.db.GetAllByQuery(query, &tokens)
	return
}

// GetDAppBySlug -
func (storage *Storage) GetDAppBySlug(slug string) (*tzip.DApp, error) {
	model, err := storage.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	return &model.DApps[0], err
}

// GetBySlug -
func (storage *Storage) GetBySlug(slug string) (*tzip.TZIP, error) {
	query := storage.db.Query(models.DocTZIP).Match("dapps.slug", slug)

	var model tzip.TZIP
	err := storage.db.GetOne(query, &model)
	return &model, err
}

// GetAliasesMap -
func (storage *Storage) GetAliasesMap(network string) (map[string]string, error) {
	it := storage.db.Query(models.DocTZIP).
		Select("address", "name").
		Match("network", network).
		Not().
		Match("name", "").Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}
	aliases := make(map[string]string)

	type res struct {
		Address string `reindex:"address"`
		Name    string `reindex:"name"`
	}
	for it.Next() {
		var r res
		it.NextObj(&r)
		aliases[r.Address] = r.Name
	}

	return aliases, nil
}

// GetAliases -
func (storage *Storage) GetAliases(network string) (aliases []tzip.TZIP, err error) {
	query := storage.db.Query(models.DocTZIP).
		Match("network", network).
		Not().
		Match("name", "")

	err = storage.db.GetAllByQuery(query, &aliases)
	return
}

// GetAlias -
func (storage *Storage) GetAlias(network, address string) (*tzip.TZIP, error) {
	query := storage.db.Query(models.DocTZIP).
		Match("network", network).
		Match("address", address)

	var data tzip.TZIP
	err := storage.db.GetOne(query, &data)
	return &data, err
}

// GetWithEvents -
func (storage *Storage) GetWithEvents() (tzips []tzip.TZIP, err error) {
	query := storage.db.Query(models.DocTZIP).
		Not().
		Where("events", reindexer.EMPTY, 0)
	err = storage.db.GetAllByQuery(query, &tzips)
	return
}
