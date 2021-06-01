package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/repository"
)

// FillTZIP -
type FillTZIP struct{}

// Key -
func (m *FillTZIP) Key() string {
	return "fill_tzip"
}

// Description -
func (m *FillTZIP) Description() string {
	return "fill tzip metadata from filesystem repository"
}

// Do - migrate function
func (m *FillTZIP) Do(ctx *config.Context) error {
	root, err := ask("Enter full path to directory with TZIP data (if empty - /etc/bcd/off-chain-metadata):")
	if err != nil {
		return err
	}
	if root == "" {
		root = "/etc/bcd/off-chain-metadata"
	}

	fs := repository.NewFileSystem(root)

	networks := make(map[string]struct{})
	for _, network := range ctx.Config.Scripts.Networks {
		networks[network] = struct{}{}
	}

	inserts := make([]models.Model, 0)
	updates := make([]models.Model, 0)

	network, err := ask("Enter network if you want certain TZIP will be added (all if empty):")
	if err != nil {
		return err
	}

	if err := ctx.Storage.CreateIndexes(); err != nil {
		return err
	}

	if network == "" {
		items, err := fs.GetAll()
		if err != nil {
			return err
		}
		for i := range items {
			if _, ok := networks[items[i].Network.String()]; !ok {
				continue
			}

			if err := processTzipItem(ctx, items[i], &inserts, &updates); err != nil {
				return err
			}
		}
	} else {
		name, err := ask("Enter directory name of the TZIP (required):")
		if name == "" {
			err = errors.New("you have to enter TZIP name")
		}
		if err != nil {
			return err
		}
		item, err := fs.Get(network, name)
		if err != nil {
			return err
		}

		if err := processTzipItem(ctx, item, &inserts, &updates); err != nil {
			return err
		}
	}

	logger.WithField("new", len(inserts)).WithField("updates", len(updates)).Info("Saving metadata...")
	if err := ctx.Storage.Save(inserts); err != nil {
		return err
	}

	return ctx.Storage.Save(updates)
}

func processTzipItem(ctx *config.Context, item repository.Item, inserts, updates *[]models.Model) error {
	model, err := item.ToModel()
	if err != nil {
		return err
	}

	for _, token := range model.Tokens.Static {
		*inserts = append(*inserts, &tokenmetadata.TokenMetadata{
			Network:   model.Network,
			Contract:  model.Address,
			Level:     0,
			Timestamp: model.Timestamp,
			TokenID:   token.TokenID,
			Symbol:    token.Symbol,
			Name:      token.Name,
			Decimals:  token.Decimals,
			Extras:    token.Extras,
		})
	}

	copyModel, err := ctx.TZIP.Get(model.Network, model.Address)
	switch {
	case err == nil:
		model.ID = copyModel.ID
		if copyModel.Name != "" {
			model.Name = copyModel.Name
		}

		if copyModel.Slug != "" {
			model.Slug = copyModel.Slug
		}
		*updates = append(*updates, &model.TZIP)

		for i := range model.DApps {
			d, err := ctx.DApps.Get(model.DApps[i].Slug)
			switch {
			case err == nil:
				*updates = append(*updates, &d)
			case ctx.Storage.IsRecordNotFound(err):
				*inserts = append(*inserts, &model.DApps[i])
			default:
				logger.Error(err)
				return err
			}
		}
	case ctx.Storage.IsRecordNotFound(err):
		*inserts = append(*inserts, &model.TZIP)

		for i := range model.DApps {
			*inserts = append(*inserts, &model.DApps[i])
		}
	default:
		logger.Error(err)
		return err
	}

	*updates = append(*updates, &model.TZIP)

	return nil
}
