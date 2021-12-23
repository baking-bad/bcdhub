package noderpc

import (
	"os"
	"path/filepath"
)

// FS -
type FS struct {
	*NodeRPC

	shareDir string
	network  string
}

// NewFS -
func NewFS(uri, shareDir, network string) *FS {
	return &FS{
		NewNodeRPC(uri),
		shareDir,
		network,
	}
}

func (fs *FS) get(filename string, output interface{}) error {
	filePath := filepath.Join(fs.shareDir, "node_cache", fs.network, filename)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(output)
}

// GetOPG -
func (fs *FS) GetOPG(block int64) (group []OperationGroup, err error) {
	err = fs.get(filepath.Join(getBlockString(block), "operations.json"), &group)
	return
}

// GetHeader -
func (fs *FS) GetHeader(block int64) (header Header, err error) {
	err = fs.get(filepath.Join(getBlockString(block), "header.json"), &header)
	return
}
