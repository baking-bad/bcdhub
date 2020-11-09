package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
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
	root, err := ask("Enter full path to directory with TZIP data (if empty - current directory):")
	if err != nil {
		return err
	}
	if root == "" {
		root = "."
	}
	fs := repository.NewFileSystem(root)

	network, err := ask("Enter network if you want certain TZIP will be added (all if empty):")
	if err != nil {
		return err
	}

	blocks, err := ctx.ES.GetLastBlocks()
	if err != nil {
		return err
	}

	networks := make(map[string]struct{})
	for i := range blocks {
		networks[blocks[i].Network] = struct{}{}
	}

	data := make([]elastic.Model, 0)
	if network == "" {
		items, err := fs.GetAll()
		if err != nil {
			return err
		}
		for i := range items {
			if _, ok := networks[items[i].Network]; !ok {
				continue
			}

			model, err := items[i].ToModel()
			if err != nil {
				return err
			}
			data = append(data, model)
		}
	} else {
		name, err := ask("Enter directory name of the TZIP (mandatory):")
		if name == "" {
			err = errors.New("You have to enter TZIP name")
		}
		if err != nil {
			return err
		}
		item, err := fs.Get(network, name)
		if err != nil {
			return err
		}
		model, err := item.ToModel()
		if err != nil {
			return err
		}
		data = append(data, model)
	}

	return ctx.ES.BulkInsert(data)
}