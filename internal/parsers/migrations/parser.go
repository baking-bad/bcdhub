package migrations

import (
	"time"

	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/go-pg/pg/v10"
)

// Parser -
type Parser interface {
	Parse(script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx pg.DBI) error
	IsMigratable(address string) bool
}
