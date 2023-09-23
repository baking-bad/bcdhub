package forge

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt_Unforge(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    int
		val     *big.Int
		wantErr bool
	}{
		{
			name: "Small int",
			data: []byte{0x06},
			want: 1,
			val:  big.NewInt(6),
		},
		{
			name: "Negative small int",
			data: []byte{0x46},
			want: 1,
			val:  big.NewInt(-6),
		},
		{
			name: "Medium int",
			data: []byte{0x84, 0x0e},
			want: 2,
			val:  big.NewInt(900),
		},
		{
			name: "Negative medium int",
			data: []byte{0xc4, 0x0e},
			want: 2,
			val:  big.NewInt(-900),
		},
		{
			name: "Large int",
			data: []byte{0xba, 0x9a, 0xf7, 0xea, 0x06},
			want: 5,
			val:  big.NewInt(917431994),
		},
		{
			name: "Negative large int",
			data: []byte{0xc0, 0xf9, 0xb9, 0xd4, 0xc7, 0x23},
			want: 6,
			val:  big.NewInt(-610913435200),
		}, {
			name: "int 8b937f02",
			data: []byte{0x8b, 0x93, 0x7f, 0x02},
			want: 3,
			val:  big.NewInt(1041611),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := NewInt()
			got, err := val.Unforge(tt.data)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
			require.EqualValues(t, val.IntValue.Cmp(tt.val), 0)
		})
	}
}

func TestInt_encode(t *testing.T) {
	tests := []struct {
		name    string
		data    *big.Int
		want    []byte
		wantErr bool
	}{
		{
			name: "Small int",
			data: big.NewInt(6),
			want: []byte{0x06},
		},
		{
			name: "Negative small int",
			data: big.NewInt(-6),
			want: []byte{0x46},
		},
		{
			name: "Medium int",
			data: big.NewInt(900),
			want: []byte{0x84, 0x0e},
		},
		{
			name: "Negative medium int",
			data: big.NewInt(-900),
			want: []byte{0xc4, 0x0e},
		},
		{
			name: "Large int",
			data: big.NewInt(917431994),
			want: []byte{0xba, 0x9a, 0xf7, 0xea, 0x06},
		},
		{
			name: "Negative large int",
			data: big.NewInt(-610913435200),
			want: []byte{0xc0, 0xf9, 0xb9, 0xd4, 0xc7, 0x23},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := NewInt()
			val.IntValue.Set(tt.data)
			got, err := val.encode()
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
