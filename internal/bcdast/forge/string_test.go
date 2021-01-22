package forge

import (
	"encoding/hex"
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
