package metrics

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Handler -
type Handler struct {
	Contracts     contract.Repository
	BigMapDiffs   bigmapdiff.Repository
	Blocks        block.Repository
	Protocol      protocol.Repository
	Operations    operation.Repository
	Migrations    migration.Repository
	TokenBalances tokenbalance.Repository
	TokenMetadata tokenmetadata.Repository
	TZIP          tzip.Repository
	Storage       models.GeneralRepository
}

// New -
func New(
	contracts contract.Repository,
	bmdRepo bigmapdiff.Repository,
	blocksRepo block.Repository,
	protocolRepo protocol.Repository,
	operations operation.Repository,
	tbRepo tokenbalance.Repository,
	tmRepo tokenmetadata.Repository,
	tzipRepo tzip.Repository,
	migrationRepo migration.Repository,
	storage models.GeneralRepository,
) *Handler {
	return &Handler{contracts, bmdRepo, blocksRepo, protocolRepo, operations, migrationRepo, tbRepo, tmRepo, tzipRepo, storage}
}
