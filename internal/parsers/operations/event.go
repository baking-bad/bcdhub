package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Event -
type Event struct {
	*ParseParams
}

// NewEvent -
func NewEvent(params *ParseParams) Event {
	return Event{params}
}

// Parse -
func (p Event) Parse(data noderpc.Operation, store parsers.Store) error {
	source := account.Account{
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
	}

	event := operation.Operation{
		Hash:         p.hash,
		ProtocolID:   p.protocol.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         types.NewOperationKind(data.Kind),
		Initiator:    p.main.Initiator,
		Source:       source,
		Fee:          data.Fee,
		Counter:      p.main.Counter,
		GasLimit:     p.main.GasLimit,
		StorageLimit: p.main.StorageLimit,
		Nonce:        data.Nonce,
		ContentIndex: p.contentIdx,
		Tag:          types.NewNullString(data.Tag),
		Payload:      data.Payload,
		PayloadType:  data.Type,
		Internal:     true,
	}

	parseOperationResult(data, &event)

	event.SetBurned(*p.protocol.Constants)

	p.stackTrace.Add(event)

	store.AddOperations(&event)

	return nil
}
