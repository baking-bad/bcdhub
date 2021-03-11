package ast

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

func TestBigMapKeyHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bytes",
			input:    `{"bytes":"000018896fcfc6690baefa9aedc6d759f9bf05727e8c"}`,
			expected: "expru2YV8AanTTUSV4K21P7X4DzbuWQFVk7NewDuP1A5uamffiiFA3",
		},
		{
			name:     "string",
			input:    `{"string":"Game one!"}`,
			expected: "exprtiRSZkLKYRess9GZ3ryb4cVQD36WLo2oysZBFxKTZ2jXqcHWGj",
		},
		{
			name:     "int",
			input:    `{"int":"505506"}`,
			expected: "exprufzwVGdAX7zG91UpiAkR2yVxEDE75tHD5YgSBmYMUx22teZTCM",
		},
		{
			name:     "string",
			input:    `{"string":"metadata"}`,
			expected: "exprtuf4ctHCKfnRvAxgU8rMeqPzfb8D8e51GWR3iHkoWsFBxD8u9h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node base.Node
			if err := json.UnmarshalFromString(tt.input, &node); err != nil {
				t.Errorf("UnmarshalFromString error: %v", err)
				return
			}
			result, err := BigMapKeyHash(&node)
			if err != nil {
				t.Errorf("error in Key, error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("error in Key, got: %v, expected: %v", result, tt.expected)
			}
		})
	}
}
