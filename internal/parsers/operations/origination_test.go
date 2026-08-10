package operations

import (
	"context"
	stdJSON "encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/protocols"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// notFoundRPCError - reproduces the error the node client returns on HTTP 404,
// with the same wrapping as (*NodeRPC).parseResponse
func notFoundRPCError(uri string) error {
	return fmt.Errorf("%w (%s): %w", noderpc.ErrNodeRPCError, uri, fmt.Errorf("%w: %s", noderpc.ErrNotFound, uri))
}

func TestOrigination_Parse_scriptNotFound(t *testing.T) {
	const address = "KT1MsTf9bEgjVBb94b2R4fodp45DXCa3q1kV"
	const level = int64(243222)

	ticketUpdates := []noderpc.TicketUpdate{
		{
			TicketToken: noderpc.TicketToken{
				Ticketer:    address,
				ContentType: stdJSON.RawMessage(`{"prim":"string"}`),
				Content:     stdJSON.RawMessage(`{"string":"abc"}`),
			},
			Updates: []noderpc.TicketUpdateItem{
				{Account: "tz1WrbkDrzKVqcGXkjw4Qk4fXkjXpAJuNP1j", Amount: "1"},
			},
		},
	}

	tests := []struct {
		name string
		// script of the origination content
		script        stdJSON.RawMessage
		ticketUpdates []noderpc.TicketUpdate
		storageErr    error
		wantErr       bool
		wantStatus    types.OperationStatus
		wantAccount   bool
	}{
		{
			name:        "applied origination of a phantom contract",
			script:      stdJSON.RawMessage(`{"code":[],"storage":{"int":"0"}}`),
			storageErr:  notFoundRPCError("chains/main/blocks/243222/context/contracts/" + address + "/script"),
			wantStatus:  types.OperationStatusFailed,
			wantAccount: false,
		}, {
			name:       "applied origination of a contract without script",
			script:     nil,
			storageErr: notFoundRPCError("chains/main/blocks/243222/context/contracts/" + address + "/script"),
			wantErr:    true,
		}, {
			name:       "unrelated rpc error is not swallowed",
			script:     stdJSON.RawMessage(`{"code":[],"storage":{"int":"0"}}`),
			storageErr: noderpc.NewNodeUnavailiableError("node", 502),
			wantErr:    true,
		}, {
			name:          "phantom contract with ticket updates in the result",
			script:        stdJSON.RawMessage(`{"code":[],"storage":{"int":"0"}}`),
			ticketUpdates: ticketUpdates,
			storageErr:    notFoundRPCError("chains/main/blocks/243222/context/contracts/" + address + "/script"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			rpc := noderpc.NewMockINode(ctrl)
			rpc.EXPECT().
				GetScriptStorageRaw(gomock.Any(), address, level).
				Return(nil, tt.storageErr).
				Times(1)

			balance := int64(0)
			paidStorageSizeDiff := int64(100)

			params := &ParseParams{
				ctx:        &config.Context{RPC: rpc},
				specific:   &protocols.Specific{NeedReceiveRawStorage: true},
				stackTrace: stacktrace.New(),
				head: noderpc.Header{
					Level:     level,
					Timestamp: time.Now(),
				},
				protocol: &protocol.Protocol{
					Constants: &protocol.Constants{CostPerByte: 250},
				},
			}

			data := noderpc.Operation{
				Kind:    "origination",
				Source:  "tz1WrbkDrzKVqcGXkjw4Qk4fXkjXpAJuNP1j",
				Balance: &balance,
				Script:  tt.script,
				Result: &noderpc.OperationResult{
					Status:              "applied",
					Originated:          []string{address},
					PaidStorageSizeDiff: &paidStorageSizeDiff,
					TicketUpdates:       tt.ticketUpdates,
				},
			}

			store := parsers.NewTestStore()
			err := NewOrigination(params).Parse(context.Background(), data, store)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Len(t, store.Operations, 1)
			op := store.Operations[0]
			require.Equal(t, tt.wantStatus, op.Status)
			require.False(t, op.AllocatedDestinationContract)
			require.EqualValues(t, 0, op.Burned)

			// the contract does not exist, so neither a contract nor its account is stored
			require.Empty(t, store.Contracts)
			_, ok := store.Accounts[address]
			require.Equal(t, tt.wantAccount, ok)
		})
	}
}
