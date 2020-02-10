package pack

import (
	"fmt"
	"testing"

	"github.com/tidwall/gjson"
)

func TestMicheline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "bytes",
			input: `{
				"bytes": "000018896fcfc6690baefa9aedc6d759f9bf05727e8c"
			}`,
			expected: "050a00000016000018896fcfc6690baefa9aedc6d759f9bf05727e8c",
		},
		{
			name: "string",
			input: `{
				"string": "Game one!"
			}`,
			expected: "05010000000947616d65206f6e6521",
		},
		{
			name: "int",
			input: `{
				"int": "505506"
			}`,
			expected: "0500a2da3d",
		},
		{
			name: "string 2",
			input: `{
				"string": "sMeta"
			}`,
			expected: "050100000005734d657461",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := gjson.Parse(tt.input)
			result, err := Micheline(parsed)
			if err != nil {
				t.Errorf("error in Micheline, error: %v", err)
			}
			if fmt.Sprintf("%x", result) != tt.expected {
				t.Errorf("error in Micheline, got: %x, expected: %v", result, tt.expected)
			}
		})
	}
}
