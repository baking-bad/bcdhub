package meta

import (
	"testing"
)

func Test_addressValidator_Validate(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{
			name:  "Invalid length",
			value: "KT1",
			want:  false,
		}, {
			name:  "Invalid length 2",
			value: "tz1",
			want:  false,
		}, {
			name:  "Invalid string",
			value: "hello world",
			want:  false,
		}, {
			name:  "Invalid check sum KT",
			value: "KT1ED3b45wsZTsuHB57PYryihtMmxvAs95k1",
			want:  false,
		}, {
			name:  "Invalid check sum tz",
			value: "tz1bwKJe27WPPXwkbNbTfC4d2rkV7eCb5v41",
			want:  false,
		}, {
			name:  "Valid KT",
			value: "KT1ED3b45wsZTsuHB57PYryihtMmxvAs95kR",
			want:  true,
		}, {
			name:  "Valid tz",
			value: "tz1bwKJe27WPPXwkbNbTfC4d2rkV7eCb5v44",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &addressValidator{}
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("addressValidator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesValidator_Validate(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{
			name:  "Invalid length",
			value: "012",
			want:  false,
		}, {
			name:  "Invalid symbols",
			value: "012q",
			want:  false,
		}, {
			name:  "Valid",
			value: "0123456789abcdef",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &bytesValidator{}
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("bytesValidator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
