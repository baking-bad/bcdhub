package ticket

import (
	"testing"

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
				TicketerID:  1,
				ContentType: []byte(`{}`),
				Content:     []byte(`{}`),
			},
			want: "1e938ee9817c472dab8b9a6601d7266592ad45186d24ddd1a92bba08854f71b2",
		}, {
			name: "test 2",
			t: Ticket{
				TicketerID:  2,
				ContentType: []byte(`{}`),
				Content:     []byte(`{}`),
			},
			want: "6238d736dd79c7274ffd96c912cd18df394a773400960859f0a3320610811a1e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.Hash()
			require.Equal(t, tt.want, got)
		})
	}
}
