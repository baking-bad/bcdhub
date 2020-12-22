package tzip

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/genjidb/genji/document"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// GetTokenMetadata -
func (storage *Storage) GetTokenMetadata(ctx tzip.GetTokenMetadataContext) (tokens []tzip.TokenMetadata, err error) {
	tzips := make([]tzip.TZIP, 0)
	builder := buildGetTokenMetadataContext(ctx)
	if err = storage.db.GetAllByQuery(builder, &tzips); err != nil {
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
func (storage *Storage) GetDApps() ([]tzip.DApp, error) {
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewIsNotNull("dapps"),
	).SortAsc("dapps.order")

	tokens := make([]tzip.DApp, 0)
	err := storage.db.GetAllByQuery(builder, &tokens)
	return tokens, err
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
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewEq("dapps.slug", slug),
	)

	var model tzip.TZIP
	err := storage.db.GetOne(builder, &model)
	return &model, err
}

// GetAliasesMap -
func (storage *Storage) GetAliasesMap(network string) (map[string]string, error) {
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewEq("network", network),
		core.NewIsNotNull("name"),
	).End()

	res, err := storage.db.Query(builder.String())
	if err != nil {
		return nil, err
	}
	defer res.Close()

	aliases := make(map[string]string)
	res.Iterate(func(d document.Document) error {
		address, err := d.GetByField("address")
		if err != nil {
			return err
		}
		name, err := d.GetByField("name")
		if err != nil {
			return err
		}
		aliases[address.String()] = name.String()
		return nil
	})

	return aliases, nil
}

// GetAliases -
func (storage *Storage) GetAliases(network string) ([]tzip.TZIP, error) {
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewEq("network", network),
		core.NewIsNotNull("name"),
	)

	aliases := make([]tzip.TZIP, 0)
	err := storage.db.GetAllByQuery(builder, &aliases)
	return aliases, err
}

// GetAlias -
func (storage *Storage) GetAlias(network, address string) (*tzip.TZIP, error) {
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewEq("network", network),
		core.NewEq("address", address),
	)

	var data tzip.TZIP
	err := storage.db.GetOne(builder, &data)
	return &data, err
}

// GetWithEvents -
func (storage *Storage) GetWithEvents() ([]tzip.TZIP, error) {
	builder := core.NewBuilder().SelectAll(models.DocTZIP).And(
		core.NewIsNotNull("events"),
	)

	tzips := make([]tzip.TZIP, 0)
	err := storage.db.GetAllByQuery(builder, &tzips)
	return tzips, err
}
