package migrations

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Parser -
type Parser interface {
	Parse(ctx context.Context, script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx models.Transaction) error
	IsMigratable(address string) bool
}
