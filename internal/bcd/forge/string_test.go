package forge

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestString_Unforge(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    int
		wantErr bool
		val     string
	}{
		{
			name: "String",
			data: "000000096d696368656c696e65",
			val:  "micheline",
			want: 13,
		},
		{
			name: "Empty string",
			data: "00000000",
			val:  "",
			want: 4,
		},
		{
			name: "totalBurned",
			data: "0000000b746f74616c4275726e6564",
			val:  "totalBurned",
			want: 15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := hex.DecodeString(tt.data)
			require.NoError(t, err)

			got, err := new(String).Unforge(input)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestString_Forge(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{
			name: "String",
			s:    "micheline",
			want: "01000000096d696368656c696e65",
		}, {
			name: "Empty string",
			s:    "",
			want: "0100000000",
		}, {
			name: "totalBurned",
			s:    "totalBurned",
			want: "010000000b746f74616c4275726e6564",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(String)
			val := tt.s
			s.StringValue = &val

			got, err := s.Forge()
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}

			want, err := hex.DecodeString(tt.want)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
