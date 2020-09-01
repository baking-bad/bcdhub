package compilers

import (
	"fmt"
	"path/filepath"

	"github.com/baking-bad/bcdhub/internal/helpers"
)

// Paths to compilers
const (
	LigoPath      = "ligo"
	MichelsonPath = "/usr/local/bin/tezos-client"
	SmartpyPath   = "/root/smartpy-cli/SmartPy.sh"
)

// Compiler -
type Compiler interface {
	Compile(path string) (*Data, error)
	Language() string
}

// Data -
type Data struct {
	Script   string
	Language string
}

// BuildFromFile -
func BuildFromFile(path string) (*Data, error) {
	cmpls := map[string]Compiler{
		".ligo":   new(ligo),
		".religo": new(ligo),
		".mligo":  new(ligo),
		".tz":     new(michelson),
		".py":     new(smartpy),
	}

	ext := filepath.Ext(path)
	compiler, ok := cmpls[ext]
	if !ok {
		return nil, fmt.Errorf("invalid file extension %v", path)
	}

	return compiler.Compile(path)
}

// IsValidExtension -
func IsValidExtension(ext string) bool {
	return helpers.StringInArray(ext, []string{".ligo", ".religo", ".mligo", ".tz", ".py"})
}
