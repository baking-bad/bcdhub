package repository

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// FileSystem -
type FileSystem struct {
	root string
}

// NewFileSystem -
func NewFileSystem(root string) *FileSystem {
	return &FileSystem{root}
}

// GetAll -
func (fs *FileSystem) GetAll() ([]Item, error) {
	networks, err := ioutil.ReadDir(fs.root)
	if err != nil {
		return nil, err
	}

	items := make([]Item, 0)

	for _, network := range networks {
		if !network.IsDir() {
			continue
		}

		metadata, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", fs.root, network.Name()))
		if err != nil {
			return nil, err
		}

		for _, m := range metadata {
			if !m.IsDir() {
				continue
			}

			item, err := fs.Get(network.Name(), m.Name())
			if err != nil {
				return nil, err
			}

			items = append(items, item)
		}
	}
	return items, nil
}

// Get -
func (fs *FileSystem) Get(network, name string) (item Item, err error) {
	path := fmt.Sprintf("%s/%s/%s", fs.root, network, name)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := fmt.Sprintf("%s/%s", path, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return item, err
		}
		item.Metadata = data
		item.Network = types.NewNetwork(network)
		item.Address = strings.TrimSuffix(file.Name(), ".json")
		break
	}

	return
}
