package models

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/go-pg/pg/v10"
)

// Model -
type Model interface {
	GetID() int64
	GetIndex() string
	Save(tx pg.DBI) error
}

type Constraint interface {
	*account.Account | *bigmapaction.BigMapAction | *bigmapdiff.BigMapDiff | *bigmapdiff.BigMapState |
		*block.Block | *contract.Contract | *contract.Script | *contract.GlobalConstant | *contract.ScriptConstants |
		*migration.Migration | *operation.Operation | *protocol.Protocol | domains.BigMapDiff

	Model
}
