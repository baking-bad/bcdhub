package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/rs/zerolog/log"
)

// TransferTicket -
type TransferTicket struct {
	*ParseParams
}

// NewTransferTicket -
func NewTransferTicket(params *ParseParams) TransferTicket {
	return TransferTicket{params}
}

// Parse -
func (p TransferTicket) Parse(data noderpc.Operation, store parsers.Store) error {
	log.Info().Msg("transfer ticket parser")
	source := account.Account{
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
	}

	transferTicket := operation.Operation{
		Source:       source,
		Initiator:    source,
		StorageLimit: data.StorageLimit,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		Hash:         p.hash,
		ProtocolID:   p.protocol.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         types.NewOperationKind(data.Kind),
		Entrypoint:   types.NewNullString(data.Entrypoint),
		ContentIndex: p.contentIdx,
		Payload:      data.TicketContent,
		PayloadType:  data.TicketType,
	}

	if data.Destination != nil {
		transferTicket.Destination = account.Account{
			Address: *data.Destination,
			Type:    types.NewAccountType(*data.Destination),
		}
	}

	p.fillInternal(&transferTicket)
	transferTicket.SetBurned(*p.protocol.Constants)
	parseOperationResult(data, &transferTicket)
	p.stackTrace.Add(transferTicket)

	store.AddOperations(&transferTicket)

	return nil
}

func (p TransferTicket) fillInternal(tx *operation.Operation) {
	if p.main == nil {
		p.main = tx
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}
