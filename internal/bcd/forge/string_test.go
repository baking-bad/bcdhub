package forge

import (
	"encoding/hex"
	"reflect"
	"testing"
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
			if err != nil {
				t.Errorf("String.Unforge() DecodeString error = %v", err)
				return
			}
			s := &String{}
			got, err := s.Unforge(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Unforge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String.Unforge() = %v, want %v", got, tt.want)
			}
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
			s.StringValue = &tt.s

			got, err := s.Forge()
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Forge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, err := hex.DecodeString(tt.want)
			if err != nil {
				t.Errorf("String.Forge() DecodeString error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("String.Forge() = %v, want %v", got, want)
			}
		})
	}
}
