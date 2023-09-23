package migrations

import (
	"context"
	"time"

	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/uptrace/bun"
)

// Parser -
type Parser interface {
	Parse(ctx context.Context, script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx bun.IDB) error
	IsMigratable(address string) bool
}
