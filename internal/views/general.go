package views

import (
	"context"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// errors
var (
	ErrNodeReturn = errors.New(`Node return error`)
)

// Args -
type Args struct {
	Protocol                 string
	Contract                 string
	Parameters               string
	Source                   string
	Initiator                string
	ChainID                  string
	HardGasLimitPerOperation int64
	Amount                   int64
}

// View -
type View interface {
	Return() []byte
	Execute(ctx context.Context, rpc noderpc.INode, args Args) ([]byte, error)
}

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}
