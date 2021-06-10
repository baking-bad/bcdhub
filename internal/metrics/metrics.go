package metrics

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Handler -
type Handler struct {
	Contracts     contract.Repository
	Blocks        block.Repository
	Protocol      protocol.Repository
	Operations    operation.Repository
	TokenBalances tokenbalance.Repository
	TZIP          tzip.Repository
	Storage       models.GeneralRepository
}

// New -
func New(
	contracts contract.Repository,
	blocksRepo block.Repository,
	protocolRepo protocol.Repository,
	operations operation.Repository,
	tbRepo tokenbalance.Repository,
	tzipRepo tzip.Repository,
	storage models.GeneralRepository,
) *Handler {
	return &Handler{contracts, blocksRepo, protocolRepo, operations, tbRepo, tzipRepo, storage}
}
