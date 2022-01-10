package storage

import (
	"reflect"
	"testing"
)

func Test_decodeData(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    []byte
		wantErr bool
	}{
		{
			name:  "test 1",
			value: `{"bytes": "7b226e61"}`,
			want:  []byte{0x7b, 0x22, 0x6e, 0x61},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeData([]byte(tt.value))
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeData() = %v, want %v", got, tt.want)
			}
		})
	}
}
