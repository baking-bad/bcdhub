package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetStrings(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []string
	}{
		{
			name: "test 1",
			data: []byte(`{"bytes":"62616c6c732e74657a"}`),
			want: []string{
				"balls.tez",
			},
		}, {
			name: "test 2",
			data: []byte(`{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"0000c0ca282a775946b5ecbe02e5cf73e25f6b62b70c"}]},[]]},{"prim":"Pair","args":[{"prim":"Some","args":[{"bytes":"62616c6c732e74657a"}]},[]]}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"2"},{"bytes":"0000753f63893674b6d523f925f0d787bf9270b95c33"}]},{"prim":"Some","args":[{"int":"3223"}]}]}]}`),
			want: []string{
				"tz1dDQc4KsTHEFe3USc66Wti2pBatZ3UDbD4",
				"balls.tez",
				"tz1WKygtstVY96oyc6Rmk945dMf33LeihgWT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStrings(tt.data)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
