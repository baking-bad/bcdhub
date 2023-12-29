package ticket

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/stretchr/testify/require"
)

func TestTicket_Hash(t *testing.T) {
	tests := []struct {
		name string
		t    Ticket
		want string
	}{
		{
			name: "test 1",
			t: Ticket{
				TicketerID: 1,
				Ticketer: account.Account{
					Address: "address1",
				},
				ContentType: []byte(`{}`),
				Content:     []byte(`{}`),
			},
			want: "49e3556aeeb72ede783c3a975bb10d8d19e14f0ab6b9d481de9f5ebeb0861a54",
		}, {
			name: "test 2",
			t: Ticket{
				TicketerID: 2,
				Ticketer: account.Account{
					Address: "address2",
				},
				ContentType: []byte(`{}`),
				Content:     []byte(`{}`),
			},
			want: "bcb6d6dc0d03d874520f0948f127a421eafe27a97950e39a118a03179dfce460",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.GetHash()
			require.Equal(t, tt.want, got)
		})
	}
}
