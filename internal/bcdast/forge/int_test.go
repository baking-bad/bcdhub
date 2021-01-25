package forge

import (
	"math/big"
	"reflect"
	"testing"
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := NewInt()
			got, err := val.Unforge(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int.Unforge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int.Unforge() = %v, want %v", got, tt.want)
				return
			}
			if val.IntValue.Cmp(tt.val) != 0 {
				t.Errorf("Int.Unforge() parsed value = %v, want %v", val.IntValue.Int64(), tt.val.Int64())
				return
			}
		})
	}
}

func TestInt_decode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *big.Int
		wantErr bool
	}{
		{
			name: "Small int",
			data: []byte{0x06},
			want: big.NewInt(6),
		},
		{
			name: "Negative small int",
			data: []byte{0x46},
			want: big.NewInt(-6),
		},
		{
			name: "Medium int",
			data: []byte{0x84, 0x0e},
			want: big.NewInt(900),
		},
		{
			name: "Negative medium int",
			data: []byte{0xc4, 0x0e},
			want: big.NewInt(-900),
		},
		{
			name: "Large int",
			data: []byte{0xba, 0x9a, 0xf7, 0xea, 0x06},
			want: big.NewInt(917431994),
		},
		{
			name: "Negative large int",
			data: []byte{0xc0, 0xf9, 0xb9, 0xd4, 0xc7, 0x23},
			want: big.NewInt(-610913435200),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := NewInt()
			if err := val.decode(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("Int.decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want.Cmp(val.IntValue) != 0 {
				t.Errorf("Int.decode() want = %v, got %v", tt.want, val.IntValue)
				return
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Int.encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int.encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
