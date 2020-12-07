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
			name:  "Invalid 0x bytes",
			value: "0xc",
			want:  false,
		}, {
			name:  "Valid",
			value: "0123456789abcdef",
			want:  true,
		}, {
			name:  "valid long",
			value: "c51117f21919bda5ce166ddf0903b34b07c1095ff5fba19165196819cbffce13c4340d2fcdda02f2bdd04fe3b6949729a28749d1b979699be484bdead6801a20",
			want:  true,
		}, {
			name:  "valid long with 0x",
			value: "0xc51117f21919bda5ce166ddf0903b34b07c1095ff5fba19165196819cbffce13c4340d2fcdda02f2bdd04fe3b6949729a28749d1b979699be484bdead6801a20",
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

func Test_base58Validator_Validate(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{
			name:  "invalid signature",
			value: "edsigtxngyy9YJuyR3agLQTLDkj5SR9ciL4nrhAootCoNnTAVzNyiksgxYFyxzaZFooCcBJ1212VD2Pt2JCmi5jGZ5JbCKti9Ag",
			want:  false,
		}, {
			name:  "signature",
			value: "edsigtxngyy9YJuyRS3gLQTLDkj5SR9ciL4nrhAootCoNnTAVzNyiksgxYFyxzaZFooCcBJ1212VD2Pt2JCmi5jGZ5JbCKti9Ag",
			want:  true,
		}, {
			name:  "signature 2",
			value: "edsigtYMZAhSbi3V2SdR7ij7v21ctW63Sr9DUfAZB443XS5zZegUNopKxYEkV87YuTsHsCJqTSmzxvcF4JER98Prc7Jmfk9JkdQ",
			want:  true,
		}, {
			name:  "key_hash",
			value: "tz1Qc94FFQqyjX4VMPTcCzpnmYWNQYWcBoNw",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &base58Validator{}
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("base58Validator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
