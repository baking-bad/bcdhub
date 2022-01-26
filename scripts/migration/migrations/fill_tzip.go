package migrations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/repository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-pg/pg/v10"
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
	storages := []string{
		"File system",
		"GitHub",
	}

	for i := range storages {
		fmt.Printf("  [%d] %s\r\n", i, storages[i])
	}
	storageIndex, err := ask("Enter storage type #")
	if err != nil {
		return err
	}
	idx, err := strconv.Atoi(storageIndex)
	if err != nil {
		return err
	}

	root := "/etc/bcd/off-chain-metadata"
	switch idx {
	case 0:
		newRoot, err := ask("Enter full path to directory with TZIP data (if empty - /etc/bcd/off-chain-metadata): ")
		if err != nil {
			return err
		}
		if newRoot != "" {
			root = newRoot
		}

	case 1:
		gitHubUser, err := ask("GitHub user: ")
		if err != nil {
			return err
		}

		if _, err := os.Stat(root); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err := os.MkdirAll(root, os.ModePerm); err != nil {
				return err
			}
			repositoryPath, err := ask("Enter git repository path: ")
			if err != nil {
				return err
			}

			if _, err := git.PlainClone(root, false, &git.CloneOptions{
				URL:      repositoryPath,
				Progress: os.Stdout,
				Auth: &http.BasicAuth{
					Username: gitHubUser,
					Password: os.Getenv("GITHUB_TOKEN"),
				},
			}); err != nil {
				return err
			}
		} else {
			repo, err := git.PlainOpen(root)
			if err != nil {
				return err
			}
			workTree, err := repo.Worktree()
			if err != nil {
				return err
			}
			if err = workTree.Pull(&git.PullOptions{
				RemoteName: "origin",
				Progress:   os.Stdout,
				Auth: &http.BasicAuth{
					Username: gitHubUser,
					Password: os.Getenv("GITHUB_TOKEN"),
				},
			}); err != nil {
				if !errors.Is(err, git.NoErrAlreadyUpToDate) {
					return err
				}
				logger.Info().Msg(err.Error())
			}
		}

	default:
		return errors.New("invalid storage index")
	}

	fs := repository.NewFileSystem(root)

	networks := make(map[string]struct{})
	for _, network := range ctx.Config.Scripts.Networks {
		networks[network] = struct{}{}
	}

	network, err := ask("Enter network if you want certain TZIP will be added (all if empty):")
	if err != nil {
		return err
	}

	if err := ctx.Storage.CreateTables(); err != nil {
		return err
	}

	if network == "" {
		items, err := fs.GetAll()
		if err != nil {
			return err
		}
		return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
			for i := range items {
				if _, ok := networks[items[i].Network.String()]; !ok {
					continue
				}

				if err := processTzipItem(ctx, items[i], tx); err != nil {
					return err
				}
			}
			return nil
		})

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

		return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
			return processTzipItem(ctx, item, tx)
		})
	}
}

func processTzipItem(ctx *config.Context, item repository.Item, tx pg.DBI) error {
	model, err := item.ToModel()
	if err != nil {
		return err
	}

	for _, token := range model.Tokens.Static {
		tm := &tokenmetadata.TokenMetadata{
			Network:   item.Network,
			Contract:  item.Address,
			Level:     0,
			Timestamp: model.Timestamp,
			TokenID:   token.TokenID,
			Symbol:    token.Symbol,
			Name:      token.Name,
			Decimals:  token.Decimals,
			Extras:    token.Extras,
		}
		if err := tm.Save(tx); err != nil {
			return err
		}
	}

	copyModel, err := ctx.ContractMetadata.Get(item.Network, item.Address)
	switch {
	case err == nil:
		model.ID = copyModel.ID
		if copyModel.Name != "" {
			model.Name = copyModel.Name
		}

		if copyModel.Slug != "" {
			model.Slug = copyModel.Slug
		}
	case ctx.Storage.IsRecordNotFound(err):
	default:
		return err
	}

	if model.ContractMetadata.Name != "" {
		if model.ContractMetadata.Slug == "" {
			model.ContractMetadata.Slug = helpers.Slug(model.ContractMetadata.Name)
		}
		if err := model.ContractMetadata.Save(tx); err != nil {
			return err
		}

		acc := account.Account{
			Alias:   model.ContractMetadata.Name,
			Address: model.ContractMetadata.Address,
			Network: model.ContractMetadata.Network,
			Type:    types.NewAccountType(model.ContractMetadata.Address),
		}
		if _, err := tx.Model(&acc).OnConflict("(network, address) DO UPDATE").Set(`alias = excluded.alias`).Returning("id").Insert(); err != nil {
			return err
		}
	}

	for i := range model.DApps {
		d, err := ctx.DApps.Get(model.DApps[i].Slug)
		switch {
		case err == nil:
			model.DApps[i].ID = d.ID
		case ctx.Storage.IsRecordNotFound(err):
		default:
			logger.Err(err)
			return err
		}

		if err := model.DApps[i].Save(tx); err != nil {
			return err
		}
	}

	return nil
}
