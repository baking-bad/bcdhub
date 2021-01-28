package tokenbalance

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenBalance_Set(t *testing.T) {
	tests := []struct {
		name  string
		tb    *TokenBalance
		value float64
		want  string
	}{
		{
			name: "50000000000000000000",
			tb: &TokenBalance{
				Value: big.NewInt(0),
			},
			value: 50000000000000000000,
			want:  "50000000000000000000",
		}, {
			name: "1111111111111111",
			tb: &TokenBalance{
				Value: big.NewInt(0),
			},
			value: 1111111111111111,
			want:  "1111111111111111",
		}, {
			name:  "1111111111111111",
			tb:    &TokenBalance{},
			value: 1111111111111111,
			want:  "1111111111111111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tb.Set(tt.value)
			assert.Equal(t, tt.want, tt.tb.Value.String())
		})
	}
}

func TestTokenBalance_Add(t *testing.T) {
	tests := []struct {
		name  string
		tb    *TokenBalance
		value float64
		want  string
	}{
		{
			name: "50000000000000000000",
			tb: &TokenBalance{
				Value: big.NewInt(10),
			},
			value: 50000000000000000000,
			want:  "50000000000000000010",
		}, {
			name: "1111111111111111",
			tb: &TokenBalance{
				Value: big.NewInt(0),
			},
			value: 1111111111111111,
			want:  "1111111111111111",
		}, {
			name:  "1111111111111111",
			tb:    &TokenBalance{},
			value: 1111111111111111,
			want:  "1111111111111111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tb.Add(tt.value)
			assert.Equal(t, tt.want, tt.tb.Value.String())
		})
	}
}

func TestTokenBalance_Sub(t *testing.T) {
	tests := []struct {
		name  string
		tb    *TokenBalance
		value float64
		want  string
	}{
		{
			name: "50000000000000000000",
			tb: &TokenBalance{
				Value: big.NewInt(0),
			},
			value: 50000000000000000000,
			want:  "-50000000000000000000",
		}, {
			name: "1111111111111111",
			tb: &TokenBalance{
				Value: big.NewInt(1111111111111112),
			},
			value: 1111111111111111,
			want:  "1",
		}, {
			name:  "1111111111111111",
			tb:    &TokenBalance{},
			value: 1111111111111111,
			want:  "-1111111111111111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tb.Sub(tt.value)
			assert.Equal(t, tt.want, tt.tb.Value.String())
		})
	}
}
